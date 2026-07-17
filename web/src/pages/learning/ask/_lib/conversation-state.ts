import type {
  KnowledgeQAHistoryMessage,
  KnowledgeQASource,
  KnowledgeQAStreamEvent,
} from "@/api/learning/knowledge-qa";

export type KnowledgeQAUserMessage = {
  id: string;
  role: "user";
  content: string;
};

export type KnowledgeQAAssistantMessage = {
  id: string;
  questionId: string;
  role: "assistant";
  status: "retrieving" | "generating" | "completed" | "failed";
  knowledgeAnswer: string;
  learningAdvice: string;
  sources: KnowledgeQASource[];
  error?: string;
};

export type KnowledgeQAConversationMessage = KnowledgeQAUserMessage | KnowledgeQAAssistantMessage;

export type KnowledgeQAConversationState = {
  messages: KnowledgeQAConversationMessage[];
  activeAssistantId?: string;
};

export const initialKnowledgeQAConversationState: KnowledgeQAConversationState = { messages: [] };

export type KnowledgeQAConversationAction =
  | { type: "start"; question: string; userId: string; assistantId: string }
  | { type: "retry"; assistantId: string; replacementId: string }
  | { type: "event"; assistantId: string; event: KnowledgeQAStreamEvent }
  | { type: "fail"; assistantId: string; message: string };

export function knowledgeQAConversationReducer(
  state: KnowledgeQAConversationState,
  action: KnowledgeQAConversationAction,
): KnowledgeQAConversationState {
  switch (action.type) {
    case "start": {
      const user: KnowledgeQAUserMessage = {
        id: action.userId,
        role: "user",
        content: action.question,
      };
      const assistant = emptyAssistant(action.assistantId, action.userId);
      return {
        messages: [...state.messages, user, assistant],
        activeAssistantId: action.assistantId,
      };
    }
    case "retry": {
      const target = state.messages.find(
        (message): message is KnowledgeQAAssistantMessage =>
          message.role === "assistant" && message.id === action.assistantId,
      );
      if (!target || target.status !== "failed") return state;
      return {
        messages: state.messages.map((message) =>
          message.id === action.assistantId
            ? emptyAssistant(action.replacementId, target.questionId)
            : message,
        ),
        activeAssistantId: action.replacementId,
      };
    }
    case "event":
      if (state.activeAssistantId !== action.assistantId) return state;
      return applyEvent(state, action.assistantId, action.event);
    case "fail":
      if (state.activeAssistantId !== action.assistantId) return state;
      return {
        messages: updateAssistant(state.messages, action.assistantId, (message) => ({
          ...message,
          status: "failed",
          error: action.message,
        })),
        activeAssistantId: undefined,
      };
  }
}

export function serializeKnowledgeQAHistory(
  messages: KnowledgeQAConversationMessage[],
): KnowledgeQAHistoryMessage[] {
  const result: KnowledgeQAHistoryMessage[] = [];
  for (let index = 0; index < messages.length - 1; index += 1) {
    const user = messages[index];
    const assistant = messages[index + 1];
    if (
      user.role !== "user" ||
      assistant.role !== "assistant" ||
      assistant.questionId !== user.id ||
      assistant.status !== "completed"
    ) {
      continue;
    }
    result.push(
      { role: "user", content: user.content },
      {
        role: "assistant",
        content: `知识回答：\n${assistant.knowledgeAnswer}\n\n学习建议：\n${assistant.learningAdvice}`,
      },
    );
    index += 1;
  }
  return result;
}

export function findQuestionForAssistant(
  messages: KnowledgeQAConversationMessage[],
  assistantId: string,
): string | undefined {
  const assistant = messages.find(
    (message): message is KnowledgeQAAssistantMessage =>
      message.role === "assistant" && message.id === assistantId,
  );
  if (!assistant) return undefined;
  const question = messages.find(
    (message): message is KnowledgeQAUserMessage =>
      message.role === "user" && message.id === assistant.questionId,
  );
  return question?.content;
}

function emptyAssistant(id: string, questionId: string): KnowledgeQAAssistantMessage {
  return {
    id,
    questionId,
    role: "assistant",
    status: "retrieving",
    knowledgeAnswer: "",
    learningAdvice: "",
    sources: [],
  };
}

function applyEvent(
  state: KnowledgeQAConversationState,
  assistantId: string,
  event: KnowledgeQAStreamEvent,
): KnowledgeQAConversationState {
  if (event.type === "error") {
    return knowledgeQAConversationReducer(state, {
      type: "fail",
      assistantId,
      message: event.message,
    });
  }
  return {
    messages: updateAssistant(state.messages, assistantId, (message) => {
      switch (event.type) {
        case "status":
          return { ...message, status: "generating" };
        case "sources":
          return { ...message, sources: event.sources };
        case "answer":
          return {
            ...message,
            status: "completed",
            knowledgeAnswer: event.knowledgeAnswer,
            learningAdvice: event.learningAdvice,
            error: undefined,
          };
      }
    }),
    activeAssistantId: event.type === "answer" ? undefined : state.activeAssistantId,
  };
}

function updateAssistant(
  messages: KnowledgeQAConversationMessage[],
  assistantId: string,
  update: (message: KnowledgeQAAssistantMessage) => KnowledgeQAAssistantMessage,
) {
  return messages.map((message) =>
    message.role === "assistant" && message.id === assistantId ? update(message) : message,
  );
}
