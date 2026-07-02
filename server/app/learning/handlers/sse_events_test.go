package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// makeEvent wraps a single Message into an AgentEvent as a non-streaming variant.
func makeEvent(agentName string, msg *schema.Message) *adk.AgentEvent {
	return &adk.AgentEvent{
		AgentName: agentName,
		Output: &adk.AgentOutput{
			MessageOutput: &adk.MessageVariant{
				IsStreaming: false,
				Message:     msg,
			},
		},
	}
}

// parseFrame 拆 "data: {...}\n\n" 一行,返回 JSON 内容。
func parseFrame(t *testing.T, raw string) (string, map[string]interface{}) {
	t.Helper()
	line := strings.TrimSpace(raw)
	if !strings.HasPrefix(line, "data: ") {
		t.Fatalf("frame does not start with 'data: ': %q", raw)
	}
	payload := strings.TrimPrefix(line, "data: ")
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		t.Fatalf("invalid JSON in frame %q: %v", payload, err)
	}
	return payload, m
}

// splitFrames 把 SSE 输出按 \n\n 切分,过滤空帧。
func splitFrames(raw string) []string {
	frames := []string{}
	for _, f := range strings.Split(raw, "\n\n") {
		if strings.TrimSpace(f) == "" {
			continue
		}
		frames = append(frames, f)
	}
	return frames
}

// TestWriteLearningSSEContentEmitsReasoningToolAndText 验证单次完整 agent 输出按以下顺序产生 SSE 帧:
//
//	reasoning → tool_call → tool_result → stream_chunk → [DONE]
func TestWriteLearningSSEContentEmitsReasoningToolAndText(t *testing.T) {
	iter, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()

	// 1. assistant: reasoning + tool_call (无 Content)
	gen.Send(makeEvent("LearningCoach", &schema.Message{
		Role:             schema.Assistant,
		ReasoningContent: "让我先查一下文件夹",
		ToolCalls: []schema.ToolCall{
			{ID: "call-1", Type: "function", Function: schema.FunctionCall{
				Name:      "list_folders",
				Arguments: `{"limit":50}`,
			}},
		},
	}))

	// 2. tool: list_folders 返回
	gen.Send(makeEvent("LearningCoach", &schema.Message{
		Role:       schema.Tool,
		ToolCallID: "call-1",
		ToolName:   "list_folders",
		Content:    `[{"id":"folder-go","name":"Go 基础"}]`,
	}))

	// 3. assistant: 文本主体(含 <ACTION> 标签)
	gen.Send(makeEvent("LearningCoach", &schema.Message{
		Role:    schema.Assistant,
		Content: "我们继续 Go 基础。<ACTION>{\"type\":\"navigate_to_practice\",\"objective_id\":\"obj-1\",\"label\":\"进入练习\"}</ACTION>",
	}))

	gen.Close()

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	acc := writeLearningSSEContent(w, iter)
	_ = w.Flush()

	// 断言累积内容包含 <ACTION>(供后续 ParseCoachAction 使用)
	if !strings.Contains(acc, "<ACTION>") {
		t.Fatalf("acc should keep <ACTION> tag for ParseCoachAction, got %q", acc)
	}
	if !strings.Contains(acc, "我们继续 Go 基础。") {
		t.Fatalf("acc missing visible text: %q", acc)
	}

	frames := splitFrames(buf.String())

	wantOrder := []string{
		SSEReasoning,
		SSEToolCall,
		SSEToolResult,
		SSEStreamChunk,
	}
	if len(frames) != len(wantOrder) {
		t.Fatalf("expected %d frames, got %d:\n%s", len(wantOrder), len(frames), buf.String())
	}
	for i, want := range wantOrder {
		_, payload := parseFrame(t, frames[i])
		gotType, _ := payload["type"].(string)
		if gotType != want {
			t.Fatalf("frame %d type = %q, want %q (full=%v)", i, gotType, want, payload)
		}
	}

	// 校验每条帧的关键字段
	_, reasoning := parseFrame(t, frames[0])
	if reasoning["content"] != "让我先查一下文件夹" {
		t.Fatalf("reasoning content = %v", reasoning["content"])
	}
	if reasoning["agent"] != "LearningCoach" {
		t.Fatalf("reasoning agent = %v", reasoning["agent"])
	}

	_, toolCall := parseFrame(t, frames[1])
	if toolCall["id"] != "call-1" {
		t.Fatalf("tool_call id = %v", toolCall["id"])
	}
	if toolCall["name"] != "list_folders" {
		t.Fatalf("tool_call name = %v", toolCall["name"])
	}
	if toolCall["arguments"] != `{"limit":50}` {
		t.Fatalf("tool_call arguments = %v", toolCall["arguments"])
	}

	_, toolResult := parseFrame(t, frames[2])
	if toolResult["tool_call_id"] != "call-1" {
		t.Fatalf("tool_result tool_call_id = %v", toolResult["tool_call_id"])
	}
	if toolResult["tool_name"] != "list_folders" {
		t.Fatalf("tool_result tool_name = %v", toolResult["tool_name"])
	}
	if !strings.Contains(toolResult["content"].(string), "folder-go") {
		t.Fatalf("tool_result content missing: %v", toolResult["content"])
	}

	_, chunk := parseFrame(t, frames[3])
	if !strings.Contains(chunk["content"].(string), "我们继续 Go 基础。") {
		t.Fatalf("stream_chunk content missing: %v", chunk["content"])
	}
	if !strings.Contains(chunk["content"].(string), "<ACTION>") {
		t.Fatalf("stream_chunk should keep raw content incl. <ACTION> for ParseCoachAction: %v", chunk["content"])
	}
}

// TestDispatchMessageChunkNilSafe 验证 nil chunk 不会 panic,且不会发出任何帧。
func TestDispatchMessageChunkNilSafe(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	var acc strings.Builder
	state := &thinkParseState{}
	dispatchMessageChunk(w, nil, "agent", &acc, state)
	_ = w.Flush()
	if buf.Len() != 0 {
		t.Fatalf("expected empty output, got %q", buf.String())
	}
	if acc.Len() != 0 {
		t.Fatalf("acc should stay empty, got %q", acc.String())
	}
}

// TestDispatchMessageChunkPlainAssistant 仅 assistant + Content 的最简分支。
func TestDispatchMessageChunkPlainAssistant(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	var acc strings.Builder
	state := &thinkParseState{}
	dispatchMessageChunk(w, &schema.Message{
		Role:    schema.Assistant,
		Content: "hello",
	}, "agent", &acc, state)
	_ = w.Flush()

	frames := splitFrames(buf.String())
	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d: %q", len(frames), buf.String())
	}
	_, payload := parseFrame(t, frames[0])
	if payload["type"] != SSEStreamChunk {
		t.Fatalf("type = %v, want %q", payload["type"], SSEStreamChunk)
	}
	if payload["content"] != "hello" {
		t.Fatalf("content = %v", payload["content"])
	}
	if acc.String() != "hello" {
		t.Fatalf("acc = %q, want %q", acc.String(), "hello")
	}
}

// TestDispatchInlineThinkBlockInSingleChunk 单 chunk 同时含 <think>...</think> 与正文。
func TestDispatchInlineThinkBlockInSingleChunk(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	var acc strings.Builder
	state := &thinkParseState{}
	dispatchMessageChunk(w, &schema.Message{
		Role:    schema.Assistant,
		Content: "<think>分析中</think>你好",
	}, "agent", &acc, state)
	_ = w.Flush()

	frames := splitFrames(buf.String())
	want := []struct {
		typ, content string
	}{
		{SSEReasoning, "分析中"},
		{SSEStreamChunk, "你好"},
	}
	if len(frames) != len(want) {
		t.Fatalf("expected %d frames, got %d:\n%s", len(want), len(frames), buf.String())
	}
	for i, w := range want {
		_, payload := parseFrame(t, frames[i])
		if payload["type"] != w.typ {
			t.Fatalf("frame %d type = %v, want %q", i, payload["type"], w.typ)
		}
		if payload["content"] != w.content {
			t.Fatalf("frame %d content = %v, want %q", i, payload["content"], w.content)
		}
	}
	if acc.String() != "你好" {
		t.Fatalf("acc = %q, want %q", acc.String(), "你好")
	}
}

// TestDispatchThinkOpenSplitAcrossChunks <think> 标签跨 chunk: 前几个 chunk 是思考,
// 后续 chunk 切到正文。验证 reasoning 与 stream_chunk 按顺序产出,acc 不含 <think> 内容。
func TestDispatchThinkOpenSplitAcrossChunks(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	var acc strings.Builder
	state := &thinkParseState{}

	chunks := []string{
		"<think>用户继续学习,我",
		"需要先决定推荐哪个小节",
		"。没有学习小节记录,所",
		"以问用户。</think>请选择:",
	}
	for _, c := range chunks {
		dispatchMessageChunk(w, &schema.Message{Role: schema.Assistant, Content: c}, "agent", &acc, state)
	}
	_ = w.Flush()

	// 每个思考中的 chunk 都会独立发一条 reasoning 事件(增量流式),
	// 最后一帧是 stream_chunk。
	frames := splitFrames(buf.String())
	if len(frames) != 5 {
		t.Fatalf("expected 5 frames, got %d:\n%s", len(frames), buf.String())
	}
	want := []struct {
		typ, content string
	}{
		{SSEReasoning, "用户继续学习,我"},
		{SSEReasoning, "需要先决定推荐哪个小节"},
		{SSEReasoning, "。没有学习小节记录,所"},
		{SSEReasoning, "以问用户。"},
		{SSEStreamChunk, "请选择:"},
	}
	for i, w := range want {
		_, payload := parseFrame(t, frames[i])
		if payload["type"] != w.typ {
			t.Fatalf("frame %d type = %v, want %q", i, payload["type"], w.typ)
		}
		if payload["content"] != w.content {
			t.Fatalf("frame %d content = %v, want %q", i, payload["content"], w.content)
		}
	}
	if acc.String() != "请选择:" {
		t.Fatalf("acc = %q, want %q", acc.String(), "请选择:")
	}
}

// TestDispatchThinkTagSplitMidTag 标签自身跨 chunk: chunk 1 是 "<th",
// chunk 2 是 "ink>..." 验证 <think> 仍被正确识别为 open tag。
func TestDispatchThinkTagSplitMidTag(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	var acc strings.Builder
	state := &thinkParseState{}

	dispatchMessageChunk(w, &schema.Message{Role: schema.Assistant, Content: "在思考<th"}, "agent", &acc, state)
	dispatchMessageChunk(w, &schema.Message{Role: schema.Assistant, Content: "ink>分析</think>你好"}, "agent", &acc, state)
	_ = w.Flush()

	frames := splitFrames(buf.String())
	want := []struct {
		typ, content string
	}{
		{SSEStreamChunk, "在思考"},
		{SSEReasoning, "分析"},
		{SSEStreamChunk, "你好"},
	}
	if len(frames) != len(want) {
		t.Fatalf("expected %d frames, got %d:\n%s", len(want), len(frames), buf.String())
	}
	for i, w := range want {
		_, payload := parseFrame(t, frames[i])
		if payload["type"] != w.typ {
			t.Fatalf("frame %d type = %v, want %q", i, payload["type"], w.typ)
		}
		if payload["content"] != w.content {
			t.Fatalf("frame %d content = %v, want %q", i, payload["content"], w.content)
		}
	}
	if acc.String() != "在思考你好" {
		t.Fatalf("acc = %q, want %q", acc.String(), "在思考你好")
	}
}

// TestDispatchCloseTagSplitMidTag 验证 </think> 跨 chunk 也能被识别。
func TestDispatchCloseTagSplitMidTag(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	var acc strings.Builder
	state := &thinkParseState{}

	dispatchMessageChunk(w, &schema.Message{Role: schema.Assistant, Content: "<think>reasoning</th"}, "agent", &acc, state)
	dispatchMessageChunk(w, &schema.Message{Role: schema.Assistant, Content: "ink>after"}, "agent", &acc, state)
	_ = w.Flush()

	frames := splitFrames(buf.String())
	want := []struct {
		typ, content string
	}{
		{SSEReasoning, "reasoning"},
		{SSEStreamChunk, "after"},
	}
	if len(frames) != len(want) {
		t.Fatalf("expected %d frames, got %d:\n%s", len(want), len(frames), buf.String())
	}
	for i, w := range want {
		_, payload := parseFrame(t, frames[i])
		if payload["type"] != w.typ {
			t.Fatalf("frame %d type = %v, want %q", i, payload["type"], w.typ)
		}
		if payload["content"] != w.content {
			t.Fatalf("frame %d content = %v, want %q", i, payload["content"], w.content)
		}
	}
	if acc.String() != "after" {
		t.Fatalf("acc = %q, want %q", acc.String(), "after")
	}
}

// TestWriteLearningSSEContentExtractsInlineThinkingFromRealFlow 模拟用户截图里的真实数据流:
// 多个 stream_chunk 内联 <think>...</think>,验证后端把它们切成 reasoning 与 stream_chunk 两类事件。
func TestWriteLearningSSEContentExtractsInlineThinkingFromRealFlow(t *testing.T) {
	iter, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()

	gen.Send(makeEvent("LearningCoach", &schema.Message{
		Role:    schema.Assistant,
		Content: "<think>The user said \"继续学习\" (continue learning). Let",
	}))
	gen.Send(makeEvent("LearningCoach", &schema.Message{
		Role:    schema.Assistant,
		Content: " me analyze the context:\n\n1. There are many wiki folders",
	}))
	gen.Send(makeEvent("LearningCoach", &schema.Message{
		Role:    schema.Assistant,
		Content: " available\n\nSo the situation is: there are wiki docs.",
	}))
	gen.Send(makeEvent("LearningCoach", &schema.Message{
		Role:    schema.Assistant,
		Content: "</think>\n\n你好!目前你的 Verve 还没有生成任何学习小节",
	}))
	gen.Close()

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	acc := writeLearningSSEContent(w, iter)
	_ = w.Flush()

	// 累积的 acc 不应包含 <think> 标签或思考内容,只保留可见正文
	if strings.Contains(acc, "<think>") || strings.Contains(acc, "The user said") {
		t.Fatalf("acc should not contain thinking content, got %q", acc)
	}
	if !strings.Contains(acc, "你好!") {
		t.Fatalf("acc missing visible text: %q", acc)
	}

	frames := splitFrames(buf.String())
	if len(frames) < 2 {
		t.Fatalf("expected at least reasoning + stream_chunk frames, got %d:\n%s", len(frames), buf.String())
	}

	// 第一帧必须是 reasoning
	_, first := parseFrame(t, frames[0])
	if first["type"] != SSEReasoning {
		t.Fatalf("first frame type = %v, want %q", first["type"], SSEReasoning)
	}

	// 最后一帧必须是 stream_chunk 且包含可见正文
	_, last := parseFrame(t, frames[len(frames)-1])
	if last["type"] != SSEStreamChunk {
		t.Fatalf("last frame type = %v, want %q", last["type"], SSEStreamChunk)
	}
	if !strings.Contains(last["content"].(string), "你好!") {
		t.Fatalf("last stream_chunk should contain visible text: %v", last["content"])
	}
}