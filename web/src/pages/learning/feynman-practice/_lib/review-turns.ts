import type { LearningExplanationReview } from "@/api/learning";

export function mergeReviewTurns(
  current: LearningExplanationReview[],
  serverTurns: LearningExplanationReview[],
) {
  const remaining = [...current];
  const merged = serverTurns.map((serverTurn) => {
    const exactIndex = remaining.findIndex((turn) => turn.id === serverTurn.id);
    if (exactIndex >= 0) {
      remaining.splice(exactIndex, 1);
      return serverTurn;
    }

    const localIndex = remaining.findIndex(
      (turn) =>
        turn.id.startsWith("local-") &&
        turn.session_id === serverTurn.session_id &&
        turn.document_id === serverTurn.document_id &&
        turn.explanation === serverTurn.explanation,
    );
    if (localIndex >= 0) {
      remaining.splice(localIndex, 1);
    }
    return serverTurn;
  });

  return [...merged, ...remaining.filter((turn) => turn.id.startsWith("local-"))];
}
