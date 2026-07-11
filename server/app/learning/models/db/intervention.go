package db

import (
	"time"

	"github.com/uptrace/bun"
)

// 教学干预
type LearningTeachingIntervention struct {
	bun.BaseModel `bun:"table:learning_teaching_interventions,alias:lti"`

	ID                 string                 `bun:"id,pk,type:varchar(32)" json:"id"`                                        // 干预ID
	TurnID             string                 `bun:"turn_id,notnull,type:varchar(32)" json:"turn_id"`                         // 教学轮次ID
	QuestionSummary    string                 `bun:"question_summary,notnull" json:"question_summary"`                        // 卡点摘要
	KnowledgeGaps      []string               `bun:"knowledge_gaps,type:jsonb,notnull" json:"knowledge_gaps"`                 // 前置知识缺口
	ExplanationSummary string                 `bun:"explanation_summary,notnull" json:"explanation_summary"`                  // 教学内容摘要
	KeyPoints          []string               `bun:"key_points,type:jsonb,notnull" json:"key_points"`                         // 讲解关键点
	Examples           []string               `bun:"examples,type:jsonb,notnull" json:"examples"`                             // 教学示例
	Evidence           map[string]interface{} `bun:"evidence,type:jsonb,notnull" json:"evidence"`                             // 文档与检索依据
	CreatedAt          time.Time              `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"` // 创建时间
}
