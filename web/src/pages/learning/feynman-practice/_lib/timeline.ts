import type {
  LearningAgentType,
  SubmitTurnRequest,
  TimelineItem,
  WikiDocumentChangeRequest,
} from "@/api/learning";

const terminalStatusRank: Record<WikiDocumentChangeRequest["status"], number> = {
  proposed: 0,
  failed: 1,
  conflict: 1,
  cancelled: 2,
  applied: 2,
};

export function mergeTimeline(current: TimelineItem[], server: TimelineItem[]): TimelineItem[] {
  const remaining = [...current];
  const merged = server.map((serverItem) => {
    const currentIndex = remaining.findIndex(
      (item) =>
        item.turn.id === serverItem.turn.id || item.turn.request_id === serverItem.turn.request_id,
    );
    if (currentIndex < 0) return serverItem;
    const currentItem = remaining.splice(currentIndex, 1)[0];
    return preserveNewerArtifact(currentItem, serverItem);
  });
  return [...merged, ...remaining.filter((item) => item.turn.id.startsWith("local-"))];
}

export function filterTimelineByAgent(
  timeline: TimelineItem[],
  agentType: LearningAgentType,
): TimelineItem[] {
  return timeline.filter((item) => item.turn.agent_type === agentType);
}

function preserveNewerArtifact(current: TimelineItem, server: TimelineItem): TimelineItem {
  if (
    current.artifact?.type !== "wiki_change_request" ||
    server.artifact?.type !== "wiki_change_request"
  ) {
    return server;
  }
  const currentRequest = current.artifact.data;
  const serverRequest = server.artifact.data;
  if (
    terminalStatusRank[currentRequest.status] > terminalStatusRank[serverRequest.status] ||
    new Date(currentRequest.updated_at).getTime() > new Date(serverRequest.updated_at).getTime()
  ) {
    return { ...server, artifact: current.artifact };
  }
  return server;
}

export function curatorRegenerationInput(
  item: TimelineItem,
): Pick<SubmitTurnRequest, "content" | "replaces_change_request_id"> {
  if (item.artifact?.type !== "wiki_change_request") {
    throw new Error("timeline item is not a curator change request");
  }
  return {
    content: item.user_message.content,
    replaces_change_request_id: item.artifact.data.id,
  };
}
