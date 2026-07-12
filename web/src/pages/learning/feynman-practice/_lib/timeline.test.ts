import { describe, expect, it } from "vitest";

import type { TimelineItem, WikiDocumentChangeRequest } from "@/api/learning";

import { curatorRegenerationInput, mergeTimeline } from "./timeline";

function item(id: string, requestId: string, status = "completed"): TimelineItem {
  return {
    turn: {
      id,
      session_id: "session-1",
      request_id: requestId,
      agent_type: "listener",
      status: status as "processing" | "completed" | "failed",
      started_at: "2026-07-12T00:00:00Z",
      created_at: "2026-07-12T00:00:00Z",
      updated_at: "2026-07-12T00:00:00Z",
    },
    user_message: {
      id: `${id}-user`,
      session_id: "session-1",
      turn_id: id,
      role: "user",
      content: "解释",
      created_at: "2026-07-12T00:00:00Z",
    },
  };
}

function curator(status: WikiDocumentChangeRequest["status"]): TimelineItem {
  const result = item("turn-curator", "request-curator");
  result.turn.agent_type = "curator";
  result.user_message.content = "补充关闭规则";
  result.artifact = {
    type: "wiki_change_request",
    data: {
      id: "change-1",
      document_id: "doc-1",
      requested_by: "user-1",
      source_type: "learning_turn",
      source_id: "turn-curator",
      request_id: "request-curator",
      base_version: 2,
      instruction: "补充关闭规则",
      change_summary: "补充关闭规则",
      proposed_content: "# 文档",
      proposed_diff: "+新内容",
      status,
      created_at: "2026-07-12T00:00:00Z",
      updated_at: status === "proposed" ? "2026-07-12T00:00:00Z" : "2026-07-12T00:01:00Z",
    },
  };
  return result;
}

describe("mergeTimeline", () => {
  it("uses server ordering and consumes one local placeholder per request id", () => {
    const local = item("local-request-2", "request-2", "processing");
    const server1 = item("turn-1", "request-1");
    const server2 = item("turn-2", "request-2");
    expect(mergeTimeline([local], [server1, server2]).map((entry) => entry.turn.id)).toEqual([
      "turn-1",
      "turn-2",
    ]);
  });

  it("keeps newer local artifact status when a server snapshot is stale", () => {
    const merged = mergeTimeline([curator("applied")], [curator("proposed")]);
    expect((merged[0].artifact?.data as WikiDocumentChangeRequest).status).toBe("applied");
  });

  it("replaces an artifact when the server status is newer", () => {
    const merged = mergeTimeline([curator("proposed")], [curator("cancelled")]);
    expect((merged[0].artifact?.data as WikiDocumentChangeRequest).status).toBe("cancelled");
  });
});

it("builds curator regeneration input from the original instruction", () => {
  expect(curatorRegenerationInput(curator("conflict"))).toEqual({
    content: "补充关闭规则",
    replaces_change_request_id: "change-1",
  });
});
