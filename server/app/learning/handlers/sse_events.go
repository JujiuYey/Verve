package handlers

import (
	"bufio"
	"encoding/json"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// SSE 事件类型常量(对齐前端 LearningStreamEvent.type)
const (
	SSEReasoning   = "reasoning"
	SSEStreamChunk = "stream_chunk"
	SSEMessage     = "message"
	SSEToolCall    = "tool_call"
	SSEToolResult  = "tool_result"
	SSEAction      = "action"
	SSEError       = "error"

	thinkOpenTag  = "<think>"
	thinkCloseTag = "</think>"
)

// writeSSEEvent 写一条 data: <json>\n\n 帧
func writeSSEEvent(w *bufio.Writer, eventType string, payload map[string]interface{}) error {
	payload["type"] = eventType
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, _ = w.Write([]byte("data: "))
	_, _ = w.Write(data)
	_, _ = w.Write([]byte("\n\n"))
	_ = w.Flush()
	return nil
}

// thinkParseState 跨 chunk 追踪 <think>...</think> 解析状态。
// thinkBuf 缓存待处理文本(可能含跨 chunk 的不完整标签前缀);
// inThink 表示当前是否位于 <think> 块内。
type thinkParseState struct {
	inThink  bool
	thinkBuf strings.Builder
}

// dispatchMessageChunk 检查单个 chunk 并发出 0..N 条 SSE 帧:
//
//	reasoning   : chunk.ReasoningContent 非空 (LLM 提供的结构化 reasoning)
//	tool_call   : chunk.Role == Assistant 且 len(chunk.ToolCalls) > 0,每个 tool call 一条
//	tool_result : chunk.Role == Tool
//	reasoning   : Content 中匹配 <think>...</think> 的片段
//	stream_chunk: Content 中位于 <think>...</think> 之外的可见正文
//
// 可见正文同时累积到 acc,供调用方后续做 ParseCoachAction 等全文解析。
// state 用于在多个 chunk 之间追踪未闭合的 <think> 标签。
func dispatchMessageChunk(
	w *bufio.Writer,
	chunk *schema.Message,
	agent string,
	acc *strings.Builder,
	state *thinkParseState,
) {
	if chunk == nil {
		return
	}

	// 1. 思考过程(LLM 已分离的结构化 reasoning,优先发出)
	if chunk.ReasoningContent != "" {
		_ = writeSSEEvent(w, SSEReasoning, map[string]interface{}{
			"content": chunk.ReasoningContent,
			"agent":   agent,
		})
	}

	// 2. assistant 发起的工具调用
	if chunk.Role == schema.Assistant && len(chunk.ToolCalls) > 0 {
		for _, tc := range chunk.ToolCalls {
			payload := map[string]interface{}{
				"id":        tc.ID,
				"type":      tc.Type,
				"name":      tc.Function.Name,
				"arguments": tc.Function.Arguments,
			}
			if tc.Type == "" {
				payload["type"] = "function"
			}
			_ = writeSSEEvent(w, SSEToolCall, payload)
		}
	}

	// 3. tool 结果(整段正文就是工具返回)
	if chunk.Role == schema.Tool {
		_ = writeSSEEvent(w, SSEToolResult, map[string]interface{}{
			"tool_call_id": chunk.ToolCallID,
			"tool_name":    chunk.ToolName,
			"content":      chunk.Content,
		})
		return
	}

	// 4. 文本主体(可能内联 <think>...</think>)
	if chunk.Content != "" {
		state.thinkBuf.WriteString(chunk.Content)
		flushThinkBlocks(w, state, agent, acc)
	}
}

// flushThinkBlocks 消费 thinkBuf 中累积的文本,反复切出完整 <think>...</think>
// 段落发 reasoning 事件,其余发 stream_chunk 并累加到 acc。
// inThink 标记当前是否在 <think> 块内,跨调用保持。
//
// 为处理 "<thin" 跨 chunk 切分,保留尾部最多 len(tag)-1 个可能的前缀字符。
func flushThinkBlocks(w *bufio.Writer, state *thinkParseState, agent string, acc *strings.Builder) {
	for {
		s := state.thinkBuf.String()
		if s == "" {
			return
		}
		if state.inThink {
			idx := strings.Index(s, thinkCloseTag)
			if idx < 0 {
				cut := safeTagSuffix(s, thinkCloseTag)
				if cut > 0 {
					emitReasoningIfNonEmpty(w, s[:len(s)-cut], agent)
					state.thinkBuf.Reset()
					state.thinkBuf.WriteString(s[len(s)-cut:])
				} else {
					emitReasoningIfNonEmpty(w, s, agent)
					state.thinkBuf.Reset()
				}
				return
			}
			emitReasoningIfNonEmpty(w, s[:idx], agent)
			state.thinkBuf.Reset()
			state.thinkBuf.WriteString(s[idx+len(thinkCloseTag):])
			state.inThink = false
		} else {
			idx := strings.Index(s, thinkOpenTag)
			if idx < 0 {
				cut := safeTagSuffix(s, thinkOpenTag)
				if cut > 0 {
					emitChunkIfNonEmpty(w, s[:len(s)-cut], agent, acc)
					state.thinkBuf.Reset()
					state.thinkBuf.WriteString(s[len(s)-cut:])
				} else {
					emitChunkIfNonEmpty(w, s, agent, acc)
					state.thinkBuf.Reset()
				}
				return
			}
			emitChunkIfNonEmpty(w, s[:idx], agent, acc)
			state.thinkBuf.Reset()
			state.thinkBuf.WriteString(s[idx+len(thinkOpenTag):])
			state.inThink = true
		}
	}
}

func emitReasoningIfNonEmpty(w *bufio.Writer, content, agent string) {
	if content == "" {
		return
	}
	_ = writeSSEEvent(w, SSEReasoning, map[string]interface{}{
		"content": content,
		"agent":   agent,
	})
}

func emitChunkIfNonEmpty(w *bufio.Writer, content, agent string, acc *strings.Builder) {
	if content == "" {
		return
	}
	_ = writeSSEEvent(w, SSEStreamChunk, map[string]interface{}{
		"content": content,
		"agent":   agent,
	})
	acc.WriteString(content)
}

// safeTagSuffix 返回 s 末尾可作为 tag 前缀的最大字符数。
// 例如 s="hello <thi", tag="<think>" 返回 4。
// 用于跨 chunk 时把可能未写完的 tag 前缀留在 carry 里。
func safeTagSuffix(s, tag string) int {
	max := len(tag) - 1
	if max > len(s) {
		max = len(s)
	}
	for n := max; n > 0; n-- {
		if strings.HasSuffix(s, tag[:n]) {
			return n
		}
	}
	return 0
}
