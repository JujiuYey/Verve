package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/cloudwego/eino/adk"

	learning_db "sag-wiki/app/learning/models/db"
	learning_model "sag-wiki/app/learning/model"
	"sag-wiki/infrastructure/database"
)

// 学习验证服务(Examiner)
type ExaminerService struct {
	db *database.DatabaseService
}

func NewExaminerService(db *database.DatabaseService) *ExaminerService {
	return &ExaminerService{db: db}
}

type JudgeResult struct {
	Verdict      string `json:"verdict"`
	MasteryAfter string `json:"mastery_after"`
	Feedback     string `json:"feedback"`
}

// Judge 调 Examiner agent 判定一次作答
func (s *ExaminerService) Judge(ctx context.Context, obj *learning_db.LearningObjective, exType, prompt, userAnswer string) (*JudgeResult, error) {
	agent, err := learning_model.NewExaminerAgent(ctx, s.db.ModelConfigs, nil)
	if err != nil {
		return nil, err
	}
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})

	text, err := collectText(runner.Query(ctx, buildExaminerQuery(obj, exType, prompt, userAnswer)))
	if err != nil {
		return nil, err
	}

	var result JudgeResult
	if err := unmarshalJSON(text, &result); err != nil {
		return nil, err
	}
	if result.Verdict == "" {
		return nil, errors.New("examiner 未产出有效判定")
	}
	return &result, nil
}

func buildExaminerQuery(obj *learning_db.LearningObjective, exType, prompt, userAnswer string) string {
	var sb strings.Builder
	sb.WriteString("小目标:")
	sb.WriteString(obj.Title)
	sb.WriteString("\n验证类型:")
	sb.WriteString(exType)
	sb.WriteString("\n题目:")
	sb.WriteString(prompt)
	sb.WriteString("\n学习者作答:")
	sb.WriteString(userAnswer)
	return sb.String()
}

// unmarshalJSON 容错解析 LLM 的 JSON 输出(去掉 markdown fence / 前后噪声)
func unmarshalJSON(text string, v interface{}) error {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	if i := strings.Index(text, "{"); i > 0 {
		text = text[i:]
	}
	if j := strings.LastIndex(text, "}"); j >= 0 && j < len(text)-1 {
		text = text[:j+1]
	}
	return json.Unmarshal([]byte(text), v)
}
