package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/adk"

	learning_db "verve/app/learning/models/db"
	"verve/infrastructure/database"
	"verve/infrastructure/llm"
)

// 学习验证服务(Examiner)
type ExaminerService struct {
	db *database.DatabaseService
}

func NewExaminerService(db *database.DatabaseService) *ExaminerService {
	return &ExaminerService{db: db}
}

type JudgeResult struct {
	Verdict            string   `json:"verdict"`
	MasteryAfter       string   `json:"mastery_after"`
	Feedback           string   `json:"feedback"`
	Evidence           string   `json:"evidence"`
	WeakPoints         []string `json:"weak_points"`
	NextRecommendation string   `json:"next_recommendation"`
	ReviewRequired     bool     `json:"review_required"`
}

// Judge 调 Examiner agent 判定一次作答
func (s *ExaminerService) Judge(ctx context.Context, obj *learning_db.LearningObjective, exType, prompt, userAnswer string) (*JudgeResult, error) {
	agent, err := llm.NewExaminerAgent(ctx, nil)
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

	result, err := parseLearningExaminerOutput(text)
	if err != nil {
		log.Printf("❌ Examiner 输出 JSON 解析失败: objective_id=%s err=%v raw=%q", obj.ID, err, truncateForExaminerLog(text, 2000))
		return nil, err
	}
	if result.Verdict == "" {
		log.Printf("❌ Examiner 未产出 verdict: objective_id=%s parsed=%+v raw=%q", obj.ID, *result, truncateForExaminerLog(text, 2000))
		return nil, errors.New("examiner 未产出有效判定")
	}
	log.Printf("✅ Examiner 判定完成: objective_id=%s verdict=%s mastery_after=%s feedback_chars=%d weak_points=%d",
		obj.ID,
		result.Verdict,
		result.MasteryAfter,
		len(result.Feedback),
		len(result.WeakPoints),
	)
	return result, nil
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

func parseLearningExaminerOutput(text string) (*JudgeResult, error) {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var lastErr error
	for _, candidate := range jsonObjectCandidates(text) {
		var result JudgeResult
		if err := json.Unmarshal([]byte(candidate), &result); err != nil {
			lastErr = err
			continue
		}
		if result.Verdict == "" {
			lastErr = errors.New("JSON 对象不是 JudgeResult")
			continue
		}
		normalizeJudgeResult(&result)
		return &result, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("未找到 JSON 对象")
}

func normalizeJudgeResult(result *JudgeResult) {
	if result.WeakPoints == nil {
		result.WeakPoints = []string{}
	}
	result.WeakPoints = uniqueNonEmptyStrings(result.WeakPoints)
	if strings.TrimSpace(result.Evidence) == "" {
		result.Evidence = strings.TrimSpace(result.Feedback)
	}
	if strings.TrimSpace(result.NextRecommendation) == "" {
		switch result.Verdict {
		case "pass":
			result.NextRecommendation = "进入下一个小目标前,补一个自己的例子巩固本次解释。"
		case "partial":
			result.NextRecommendation = "先补齐反馈里指出的缺口,再用自己的话重新解释一遍。"
		default:
			result.NextRecommendation = "回到资料中的关键定义,先完成一次小范围复述。"
		}
	}
	if result.Verdict != "pass" {
		result.ReviewRequired = true
	}
}

func MergeLearningProfileState(profile *learning_db.LearningProfile, obj *learning_db.LearningObjective, result *JudgeResult) {
	mastery := strings.TrimSpace(result.MasteryAfter)
	if mastery != "" {
		profile.CurrentLevel = &mastery
	}
	if result.Verdict == "pass" && strings.TrimSpace(obj.Title) != "" {
		profile.CompletedTopics = appendUnique(profile.CompletedTopics, obj.Title)
	}
	profile.WeakPoints = appendUnique(profile.WeakPoints, result.WeakPoints...)
	if strings.TrimSpace(result.NextRecommendation) != "" {
		next := strings.TrimSpace(result.NextRecommendation)
		profile.NextGoal = &next
	}
	if strings.TrimSpace(result.Evidence) != "" {
		habit := strings.TrimSpace(result.Evidence)
		profile.VerificationHabits = &habit
	}
}

func uniqueNonEmptyStrings(values []string) []string {
	return appendUnique(nil, values...)
}

func appendUnique(base []string, values ...string) []string {
	seen := make(map[string]bool, len(base)+len(values))
	out := make([]string, 0, len(base)+len(values))
	for _, value := range append(base, values...) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func truncateForExaminerLog(text string, limit int) string {
	text = strings.TrimSpace(text)
	if len(text) <= limit {
		return text
	}
	return fmt.Sprintf("%s...(truncated %d chars)", text[:limit], len(text)-limit)
}
