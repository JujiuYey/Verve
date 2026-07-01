package service

import (
	"context"
	"crypto/sha256"
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

// GuideService 调用导学 agent 阅读资料并生成本节学习目标。
type GuideService struct {
	db *database.DatabaseService
}

func NewGuideService(db *database.DatabaseService) *GuideService {
	return &GuideService{db: db}
}

type GuideResult struct {
	Summary            string               `json:"summary"`
	MasteryGoals       []string             `json:"mastery_goals"`
	PracticePoints     []GuidePracticePoint `json:"practice_points"`
	ReadingSteps       []string             `json:"reading_steps"`
	Pitfalls           []string             `json:"pitfalls"`
	SelfCheckQuestions []string             `json:"self_check_questions"`
	Evidence           []string             `json:"evidence"`
}

type GuidePracticePoint struct {
	Title    string   `json:"title"`
	Goal     string   `json:"goal"`
	Evidence []string `json:"evidence"`
}

type GuideResponse struct {
	*GuideResult
	ContentHash string `json:"content_hash"`
	Cached      bool   `json:"cached"`
}

func (s *GuideService) FindCached(ctx context.Context, obj *learning_db.LearningObjective, contentHash string) (*GuideResponse, error) {
	guide, err := s.db.Guides.FindByObjectiveAndHash(ctx, obj.ID, contentHash)
	if err != nil {
		return nil, err
	}
	var result GuideResult
	if err := json.Unmarshal(guide.Result, &result); err != nil {
		return nil, err
	}
	normalizeGuideResult(&result)
	return &GuideResponse{
		GuideResult: &result,
		ContentHash: guide.ContentHash,
		Cached:      true,
	}, nil
}

func (s *GuideService) Generate(ctx context.Context, obj *learning_db.LearningObjective, markdown string) (*GuideResponse, error) {
	contentHash := GuideContentHash(markdown)
	agent, err := llm.NewGuideAgent(ctx)
	if err != nil {
		log.Printf("❌ Guide Agent 初始化失败: objective_id=%s err=%v", obj.ID, err)
		return nil, err
	}

	query := buildGuideQuery(obj, markdown)
	log.Printf("📘 Guide 开始导学: objective_id=%s objective_title=%q markdown_chars=%d",
		obj.ID,
		obj.Title,
		len(markdown),
	)

	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})
	text, err := collectText(runner.Query(ctx, query))
	if err != nil {
		log.Printf("❌ Guide Agent 调用失败: objective_id=%s err=%v", obj.ID, err)
		return nil, err
	}
	log.Printf("📘 Guide 原始输出: objective_id=%s output_chars=%d output_preview=%q", obj.ID, len(text), truncateForPlannerLog(text, 1200))

	result, err := parseGuideOutput(text)
	if err != nil {
		log.Printf("❌ Guide 输出 JSON 解析失败: objective_id=%s err=%v raw=%q", obj.ID, err, truncateForPlannerLog(text, 2000))
		return nil, err
	}
	if len(result.MasteryGoals) == 0 && result.Summary == "" {
		return nil, errors.New("guide 未产出有效导学内容")
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	if err := s.db.Guides.Upsert(ctx, &learning_db.LearningGuide{
		ObjectiveID: obj.ID,
		UserID:      obj.UserID,
		ContentHash: contentHash,
		Result:      resultJSON,
	}); err != nil {
		return nil, err
	}

	return &GuideResponse{
		GuideResult: result,
		ContentHash: contentHash,
		Cached:      false,
	}, nil
}

func buildGuideQuery(obj *learning_db.LearningObjective, markdown string) string {
	var sb strings.Builder
	sb.WriteString("当前小目标:")
	sb.WriteString(obj.Title)
	if obj.StageTitle != nil && strings.TrimSpace(*obj.StageTitle) != "" {
		sb.WriteString("\n阶段:")
		sb.WriteString(*obj.StageTitle)
	}
	if obj.Detail != nil && strings.TrimSpace(*obj.Detail) != "" {
		sb.WriteString("\n要点:")
		sb.WriteString(*obj.Detail)
	}
	sb.WriteString("\n\nMarkdown 学习资料:\n")
	sb.WriteString(truncateGuideMarkdown(markdown, 18000))
	return sb.String()
}

func parseGuideOutput(text string) (*GuideResult, error) {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var lastErr error
	for _, candidate := range jsonObjectCandidates(text) {
		var out GuideResult
		if err := json.Unmarshal([]byte(candidate), &out); err != nil {
			lastErr = err
			continue
		}
		if out.Summary == "" && len(out.MasteryGoals) == 0 && len(out.PracticePoints) == 0 {
			lastErr = errors.New("JSON 对象不是 GuideResult")
			continue
		}
		normalizeGuideResult(&out)
		return &out, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("未找到 JSON 对象")
}

func jsonObjectCandidates(text string) []string {
	var candidates []string
	for start := 0; start < len(text); start++ {
		if text[start] != '{' {
			continue
		}

		depth := 0
		inString := false
		escaped := false
		for end := start; end < len(text); end++ {
			ch := text[end]
			if inString {
				if escaped {
					escaped = false
					continue
				}
				if ch == '\\' {
					escaped = true
					continue
				}
				if ch == '"' {
					inString = false
				}
				continue
			}

			switch ch {
			case '"':
				inString = true
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					candidates = append(candidates, text[start:end+1])
					start = end
					end = len(text)
				}
			}
		}
	}
	return candidates
}

func normalizeGuideResult(result *GuideResult) {
	if result.MasteryGoals == nil {
		result.MasteryGoals = []string{}
	}
	if result.PracticePoints == nil {
		result.PracticePoints = []GuidePracticePoint{}
	}
	if result.ReadingSteps == nil {
		result.ReadingSteps = []string{}
	}
	if result.Pitfalls == nil {
		result.Pitfalls = []string{}
	}
	if result.SelfCheckQuestions == nil {
		result.SelfCheckQuestions = []string{}
	}
	if result.Evidence == nil {
		result.Evidence = []string{}
	}
	for i := range result.PracticePoints {
		if result.PracticePoints[i].Evidence == nil {
			result.PracticePoints[i].Evidence = []string{}
		}
	}
}

func GuideContentHash(markdown string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(markdown)))
	return fmt.Sprintf("%x", sum)
}

func truncateGuideMarkdown(markdown string, limit int) string {
	markdown = strings.TrimSpace(markdown)
	if len(markdown) <= limit {
		return markdown
	}
	return markdown[:limit] + "\n\n...(资料过长,已截断)"
}
