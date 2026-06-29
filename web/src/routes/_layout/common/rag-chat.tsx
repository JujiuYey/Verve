import { createFileRoute } from "@tanstack/react-router";

import { RagChatPage } from "@/pages/ai/rag-chat";

export const Route = createFileRoute("/_layout/common/rag-chat")({
  component: RagChatPage,
});
