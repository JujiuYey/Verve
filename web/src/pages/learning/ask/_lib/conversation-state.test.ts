import { describe, expect, it } from "vitest";

import {
  findQuestionForAssistant,
  initialKnowledgeQAConversationState,
  knowledgeQAConversationReducer,
  serializeKnowledgeQAHistory,
  type KnowledgeQAAssistantMessage,
} from "./conversation-state";

describe("knowledgeQAConversationReducer", () => {
  it("keeps sources while moving from retrieval to a completed structured answer", () => {
    let state = knowledgeQAConversationReducer(initialKnowledgeQAConversationState, {
      type: "start",
      question: "事务是什么？",
      userId: "user-1",
      assistantId: "assistant-1",
    });
    state = knowledgeQAConversationReducer(state, {
      type: "event",
      assistantId: "assistant-1",
      event: {
        type: "sources",
        sources: [
          {
            documentId: "doc-1",
            documentTitle: "database.md",
            folderPath: "数据库",
            headingPath: "事务",
            score: 0.91,
          },
        ],
      },
    });
    state = knowledgeQAConversationReducer(state, {
      type: "event",
      assistantId: "assistant-1",
      event: { type: "status", phase: "generating" },
    });
    state = knowledgeQAConversationReducer(state, {
      type: "event",
      assistantId: "assistant-1",
      event: {
        type: "answer",
        knowledgeAnswer: "事务是一组操作。",
        learningAdvice: "复习隔离性。",
      },
    });

    const assistant = state.messages[1] as KnowledgeQAAssistantMessage;
    expect(assistant.status).toBe("completed");
    expect(assistant.sources).toHaveLength(1);
    expect(assistant.knowledgeAnswer).toBe("事务是一组操作。");
    expect(state.activeAssistantId).toBeUndefined();
    expect(serializeKnowledgeQAHistory(state.messages)).toEqual([
      { role: "user", content: "事务是什么？" },
      {
        role: "assistant",
        content: "知识回答：\n事务是一组操作。\n\n学习建议：\n复习隔离性。",
      },
    ]);
  });

  it("excludes failed turns and retries without duplicating the user question", () => {
    let state = knowledgeQAConversationReducer(initialKnowledgeQAConversationState, {
      type: "start",
      question: "解释 MVCC",
      userId: "user-1",
      assistantId: "assistant-1",
    });
    state = knowledgeQAConversationReducer(state, {
      type: "fail",
      assistantId: "assistant-1",
      message: "检索失败",
    });
    expect(serializeKnowledgeQAHistory(state.messages)).toEqual([]);
    expect(findQuestionForAssistant(state.messages, "assistant-1")).toBe("解释 MVCC");

    state = knowledgeQAConversationReducer(state, {
      type: "retry",
      assistantId: "assistant-1",
      replacementId: "assistant-2",
    });
    expect(state.messages).toHaveLength(2);
    expect(state.messages[0]).toMatchObject({ id: "user-1", content: "解释 MVCC" });
    expect(state.messages[1]).toMatchObject({ id: "assistant-2", status: "retrieving" });
  });

  it("ignores events from an aborted or replaced request", () => {
    let state = knowledgeQAConversationReducer(initialKnowledgeQAConversationState, {
      type: "start",
      question: "解释索引",
      userId: "user-1",
      assistantId: "assistant-1",
    });
    state = knowledgeQAConversationReducer(state, {
      type: "fail",
      assistantId: "assistant-1",
      message: "连接中断",
    });
    state = knowledgeQAConversationReducer(state, {
      type: "retry",
      assistantId: "assistant-1",
      replacementId: "assistant-2",
    });
    const unchanged = knowledgeQAConversationReducer(state, {
      type: "event",
      assistantId: "assistant-1",
      event: { type: "answer", knowledgeAnswer: "迟到回答", learningAdvice: "迟到建议" },
    });
    expect(unchanged).toBe(state);
  });
});
