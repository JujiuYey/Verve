package payload

import (
	learning_db "verve/app/learning/models/db"
)

const (
	ArtifactExplanationReview    = "explanation_review"
	ArtifactTeachingIntervention = "teaching_intervention"
	ArtifactWikiChangeRequest    = "wiki_change_request"
)

type TurnArtifact struct {
	Type string `json:"type"` // 产物类型
	Data any    `json:"data"` // 产物数据
}

type TimelineItem struct {
	Turn             *learning_db.LearningTurn    `json:"turn"`                        // 学习轮次
	UserMessage      *learning_db.LearningMessage `json:"user_message"`                // 用户消息
	AssistantMessage *learning_db.LearningMessage `json:"assistant_message,omitempty"` // Agent回复
	Artifact         *TurnArtifact                `json:"artifact,omitempty"`          // 结构化产物
}
