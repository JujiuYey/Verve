import { describe, expect, it } from "vitest";

import type { LearningExplanationReview } from "@/api/learning";

import { mergeReviewTurns } from "./review-turns";

function review(id: string, explanation = "同一段解释"): LearningExplanationReview {
  return {
    id,
    session_id: "session-1",
    document_id: "document-1",
    user_id: "user-1",
    explanation,
    heard_summary: "听到的内容",
    clear_points: [],
    confusing_points: [],
    misconceptions: [],
    follow_up_question: "",
    explanation_summary: "解释摘要",
    ready_to_wrap_up: false,
    context_sufficient: true,
    created_at: "2026-07-11T10:00:00Z",
  };
}

describe("mergeReviewTurns", () => {
  it("consumes only one local placeholder for each new persisted duplicate explanation", () => {
    const persistedFirst = review("review-1");
    const localSecond = review("local-2");
    const localThird = review("local-3");

    const result = mergeReviewTurns(
      [persistedFirst, localSecond, localThird],
      [persistedFirst, review("review-2")],
    );

    expect(result.map((turn) => turn.id)).toEqual(["review-1", "review-2", "local-3"]);
  });

  it("preserves an unmatched local turn when the server snapshot is stale", () => {
    const persisted = review("review-1");
    const pendingLocal = review("local-2");

    const result = mergeReviewTurns([persisted, pendingLocal], [persisted]);

    expect(result.map((turn) => turn.id)).toEqual(["review-1", "local-2"]);
  });
});
