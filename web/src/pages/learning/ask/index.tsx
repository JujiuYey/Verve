import { useCallback, useEffect, useReducer, useRef } from "react";

import { knowledgeQAStream } from "@/api/learning";

import { KnowledgeQAWorkspace } from "./_components/knowledge-qa-workspace";
import {
  findQuestionForAssistant,
  initialKnowledgeQAConversationState,
  knowledgeQAConversationReducer,
  serializeKnowledgeQAHistory,
} from "./_lib/conversation-state";

export function KnowledgeQAPage() {
  const [state, dispatch] = useReducer(
    knowledgeQAConversationReducer,
    initialKnowledgeQAConversationState,
  );
  const stateRef = useRef(state);
  const controllerRef = useRef<AbortController | null>(null);
  const activeRequestRef = useRef<string | undefined>(undefined);
  stateRef.current = state;

  useEffect(
    () => () => {
      activeRequestRef.current = undefined;
      controllerRef.current?.abort();
    },
    [],
  );

  const runRequest = useCallback(
    (
      question: string,
      assistantId: string,
      history = serializeKnowledgeQAHistory(stateRef.current.messages),
    ) => {
      const controller = new AbortController();
      controllerRef.current?.abort();
      controllerRef.current = controller;
      activeRequestRef.current = assistantId;

      void knowledgeQAStream({
        message: question,
        history,
        signal: controller.signal,
        onEvent: (event) => {
          if (controller.signal.aborted || activeRequestRef.current !== assistantId) return;
          dispatch({ type: "event", assistantId, event });
        },
      })
        .catch((error: unknown) => {
          if (controller.signal.aborted || activeRequestRef.current !== assistantId) return;
          dispatch({
            type: "fail",
            assistantId,
            message: error instanceof Error ? error.message : "问答请求失败，请重试",
          });
        })
        .finally(() => {
          if (activeRequestRef.current === assistantId) {
            activeRequestRef.current = undefined;
            controllerRef.current = null;
          }
        });
    },
    [],
  );

  const handleSend = useCallback(
    (rawQuestion: string) => {
      const question = rawQuestion.trim();
      if (!question || activeRequestRef.current) return;
      const userId = crypto.randomUUID();
      const assistantId = crypto.randomUUID();
      const history = serializeKnowledgeQAHistory(stateRef.current.messages);
      dispatch({ type: "start", question, userId, assistantId });
      runRequest(question, assistantId, history);
    },
    [runRequest],
  );

  const handleRetry = useCallback(
    (failedAssistantId: string) => {
      if (activeRequestRef.current) return;
      const question = findQuestionForAssistant(stateRef.current.messages, failedAssistantId);
      if (!question) return;
      const assistantId = crypto.randomUUID();
      const history = serializeKnowledgeQAHistory(stateRef.current.messages);
      dispatch({ type: "retry", assistantId: failedAssistantId, replacementId: assistantId });
      runRequest(question, assistantId, history);
    },
    [runRequest],
  );

  return (
    <KnowledgeQAWorkspace
      busy={Boolean(state.activeAssistantId)}
      messages={state.messages}
      onRetry={handleRetry}
      onSend={handleSend}
    />
  );
}
