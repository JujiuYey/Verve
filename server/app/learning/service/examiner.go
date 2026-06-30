package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/adk"

	learning_model "sag-wiki/app/learning/model"
	learning_db "sag-wiki/app/learning/models/db"
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
		log.Printf("❌ Examiner Agent 初始化失败: objective_id=%s type=%s err=%v", obj.ID, exType, err)
		return nil, err
	}
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})

	query := buildExaminerQuery(obj, exType, prompt, userAnswer)
	log.Printf("🧪 Examiner 开始判定: objective_id=%s objective_title=%q type=%s prompt_chars=%d answer_chars=%d",
		obj.ID,
		obj.Title,
		exType,
		len(prompt),
		len(userAnswer),
	)

	text, err := collectText(runner.Query(ctx, query))
	if err != nil {
		log.Printf("❌ Examiner Agent 调用失败: objective_id=%s type=%s err=%v", obj.ID, exType, err)
		return nil, err
	}
	log.Printf("🧪 Examiner 原始输出: objective_id=%s output_chars=%d output_preview=%q", obj.ID, len(text), truncateForExaminerLog(text, 1200))

	var result JudgeResult
	if err := unmarshalJSON(text, &result); err != nil {
		log.Printf("❌ Examiner 输出 JSON 解析失败: objective_id=%s err=%v raw=%q", obj.ID, err, truncateForExaminerLog(text, 2000))
		return nil, err
	}
	if result.Verdict == "" {
		log.Printf("❌ Examiner 未产出 verdict: objective_id=%s parsed=%+v raw=%q", obj.ID, result, truncateForExaminerLog(text, 2000))
		return nil, errors.New("examiner 未产出有效判定")
	}
	log.Printf("✅ Examiner 判定完成: objective_id=%s verdict=%s mastery_after=%s feedback_chars=%d",
		obj.ID,
		result.Verdict,
		result.MasteryAfter,
		len(result.Feedback),
	)
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

func truncateForExaminerLog(text string, limit int) string {
	text = strings.TrimSpace(text)
	if len(text) <= limit {
		return text
	}
	return fmt.Sprintf("%s...(truncated %d chars)", text[:limit], len(text)-limit)
}
