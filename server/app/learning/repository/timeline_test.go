package repository

import (
	"testing"
	"time"

	learning_db "verve/app/learning/models/db"
	learning_payload "verve/app/learning/models/payload"
	wiki_db "verve/app/wiki/models/db"
)

func TestAssembleTimelineOrdersTurnsAndAttachesTypedArtifacts(t *testing.T) {
	now := time.Now()
	turns := []*learning_db.LearningTurn{
		{ID: "turn-2", AgentType: learning_db.LearningAgentTeacher, CreatedAt: now.Add(time.Second)},
		{ID: "turn-1", AgentType: learning_db.LearningAgentListener, CreatedAt: now},
		{ID: "turn-3", AgentType: learning_db.LearningAgentCurator, CreatedAt: now.Add(2 * time.Second)},
	}
	messages := []*learning_db.LearningMessage{
		{ID: "m2", TurnID: "turn-1", Role: "assistant"}, {ID: "m1", TurnID: "turn-1", Role: "user"},
		{ID: "m3", TurnID: "turn-2", Role: "user"}, {ID: "m4", TurnID: "turn-2", Role: "assistant"},
		{ID: "m5", TurnID: "turn-3", Role: "user"},
	}
	reviews := []*learning_db.LearningExplanationReview{{TurnID: "turn-1"}}
	interventions := []*learning_db.LearningTeachingIntervention{{TurnID: "turn-2"}}
	requests := []*wiki_db.DocumentChangeRequest{{ID: "cr-1", SourceID: "turn-3"}}

	items := assembleTimeline(turns, messages, reviews, interventions, requests)
	if len(items) != 3 || items[0].Turn.ID != "turn-1" || items[2].Turn.ID != "turn-3" {
		t.Fatalf("items = %#v", items)
	}
	wantTypes := []string{learning_payload.ArtifactExplanationReview, learning_payload.ArtifactTeachingIntervention, learning_payload.ArtifactWikiChangeRequest}
	for i, want := range wantTypes {
		if items[i].Artifact == nil || items[i].Artifact.Type != want {
			t.Fatalf("item %d artifact = %#v", i, items[i].Artifact)
		}
	}
	if items[2].AssistantMessage != nil {
		t.Fatal("failed or processing turns may omit assistant messages")
	}
}
