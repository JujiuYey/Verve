package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	"sag-wiki/app/ai/model"
	"sag-wiki/app/ai/repository"
	"sag-wiki/app/ai/service"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"
)

type ChatHandler struct {
	repo              repository.ModelConfigRepository
	dbService         *database.DatabaseService
	agentRunner       *model.AgentRunner
	retrievalService  *service.RetrievalService
}

func NewChatHandler(dbService *database.DatabaseService, repo repository.ModelConfigRepository, retrievalService *service.RetrievalService) *ChatHandler {
	return &ChatHandler{
		repo:             repo,
		dbService:        dbService,
		retrievalService: retrievalService,
	}
}

// ensureAgent 确保 agent 已初始化（懒加载）
func (h *ChatHandler) ensureAgent(ctx context.Context) error {
	// 如果已经初始化，直接返回
	if h.agentRunner != nil {
		return nil
	}

	// 尝试初始化
	a, err := model.NewAgentWithSystemTools(ctx, h.dbService, h.repo)
	if err != nil {
		// 检查是否是"未配置默认模型"的错误
		if strings.Contains(err.Error(), "获取默认模型配置失败") {
			return fmt.Errorf("未配置默认模型，请先在系统设置中添加模型配置")
		}
		return fmt.Errorf("agent 初始化失败: %w", err)
	}

	h.agentRunner = model.NewAgentRunner(ctx, a)
	return nil
}

// Chat SSE 对话接口
func (h *ChatHandler) Chat(c *fiber.Ctx) error {
	var req struct {
		Query      string `json:"query"`
		FolderID   string `json:"folder_id"`
		DocumentID string `json:"document_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}

	if req.Query == "" {
		return response.BadRequestCtx(c)
	}

	// 确保 agent 已初始化（懒加载）
	if err := h.ensureAgent(c.Context()); err != nil {
		return response.InternalServerCtx(c, err.Error())
	}

	// 构建最终查询：如果指定了 folder_id 或 document_id，先进行 RAG 检索
	query := req.Query
	if h.retrievalService != nil && (req.FolderID != "" || req.DocumentID != "") {
		userID, ok := c.Locals("user_id").(string)
		if !ok {
			userID = ""
		}

		searchReq := &service.SearchRequest{
			Query:      req.Query,
			FolderID:   req.FolderID,
			DocumentID: req.DocumentID,
			UserID:     userID,
			Limit:      10,
		}

		results, err := h.retrievalService.Search(c.Context(), searchReq)
		if err == nil && len(results) > 0 {
			// 组装检索到的上下文
			var contextBuilder strings.Builder
			contextBuilder.WriteString("以下是与问题相关的参考资料：\n\n")
			for i, result := range results {
				if result.ChunkInfo != nil {
					contextBuilder.WriteString(fmt.Sprintf("【参考资料 %d】\n%s\n\n", i+1, result.ChunkInfo.ChunkText))
				}
			}
			contextBuilder.WriteString("请基于以上参考资料回答用户问题。如果参考资料中没有相关信息，请基于已有知识回答。\n\n")
			contextBuilder.WriteString("用户问题：" + req.Query)
			query = contextBuilder.String()
		}
	}

	iter := h.agentRunner.Query(c.Context(), query)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		writeAgentStreamSSE(w, iter)
	}))
	return nil
}

func writeAgentStreamSSE(w *bufio.Writer, iter *adk.AsyncIterator[*adk.AgentEvent]) {
	for {
		event, ok := iter.Next()
		if !ok {
			// Send [DONE] marker
			_, _ = w.Write([]byte("data: [DONE]\n\n"))
			_ = w.Flush()
			return
		}

		if event.Err != nil {
			data, _ := json.Marshal(map[string]string{"type": "error", "content": event.Err.Error()})
			_, _ = w.Write([]byte("data: "))
			_, _ = w.Write(data)
			_, _ = w.Write([]byte("\n\n"))
			_ = w.Flush()
			return
		}

		// Process output
		if event.Output != nil && event.Output.MessageOutput != nil {
			msgOutput := event.Output.MessageOutput

			// Handle streaming message
			if stream := msgOutput.MessageStream; stream != nil {
				for {
					chunk, err := stream.Recv()
					if errors.Is(err, io.EOF) {
						break
					}
					if err != nil {
						data, _ := json.Marshal(map[string]string{"type": "error", "content": err.Error()})
						_, _ = w.Write([]byte("data: "))
						_, _ = w.Write(data)
						_, _ = w.Write([]byte("\n\n"))
						_ = w.Flush()
						return
					}
					if chunk == nil {
						continue
					}

					eventType := "stream_chunk"
					if chunk.Role == schema.Tool {
						eventType = "tool_result_chunk"
					}

					data, _ := json.Marshal(map[string]interface{}{
						"type":    eventType,
						"content": chunk.Content,
						"agent":   event.AgentName,
					})
					_, _ = w.Write([]byte("data: "))
					_, _ = w.Write(data)
					_, _ = w.Write([]byte("\n\n"))
					_ = w.Flush()
				}
			}

			// Handle complete message
			if msg := msgOutput.Message; msg != nil {
				eventType := "message"
				if msg.Role == schema.Tool {
					eventType = "tool_result"
				}

				sseEvent := map[string]interface{}{
					"type":    eventType,
					"content": msg.Content,
					"agent":   event.AgentName,
				}
				if len(msg.ToolCalls) > 0 {
					sseEvent["tool_calls"] = msg.ToolCalls
				}

				data, _ := json.Marshal(sseEvent)
				_, _ = w.Write([]byte("data: "))
				_, _ = w.Write(data)
				_, _ = w.Write([]byte("\n\n"))
				_ = w.Flush()
			}
		}

		// Process action events
		if event.Action != nil {
			if event.Action.Exit {
				data, _ := json.Marshal(map[string]interface{}{
					"type":    "action",
					"action":  "exit",
					"content": "Agent execution completed",
					"agent":   event.AgentName,
				})
				_, _ = w.Write([]byte("data: "))
				_, _ = w.Write(data)
				_, _ = w.Write([]byte("\n\n"))
				_ = w.Flush()
			}

			if event.Action.TransferToAgent != nil {
				data, _ := json.Marshal(map[string]interface{}{
					"type":    "action",
					"action":  "transfer",
					"content": fmt.Sprintf("Transfer to agent: %s", event.Action.TransferToAgent.DestAgentName),
					"agent":   event.AgentName,
				})
				_, _ = w.Write([]byte("data: "))
				_, _ = w.Write(data)
				_, _ = w.Write([]byte("\n\n"))
				_ = w.Flush()
			}
		}
	}
}
