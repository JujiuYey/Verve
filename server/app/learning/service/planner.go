package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/cloudwego/eino/adk"

	learning_db "sag-wiki/app/learning/models/db"
	learning_model "sag-wiki/app/learning/model"
	"sag-wiki/infrastructure/database"
)

// 学习路线生成服务(Planner)
type PlannerService struct {
	db *database.DatabaseService
}

func NewPlannerService(db *database.DatabaseService) *PlannerService {
	return &PlannerService{db: db}
}

type plannerOutput struct {
	Stages []plannerStage `json:"stages"`
}

type plannerStage struct {
	Title      string             `json:"title"`
	Objectives []plannerObjective `json:"objectives"`
}

type plannerObjective struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// GeneratePath 调 Planner agent 生成路线并落库(path + objectives)
func (s *PlannerService) GeneratePath(ctx context.Context, goal *learning_db.LearningGoal) error {
	agent, err := learning_model.NewPlannerAgent(ctx, s.db.ModelConfigs)
	if err != nil {
		return err
	}
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})

	text, err := collectText(runner.Query(ctx, goal.Title))
	if err != nil {
		return err
	}

	parsed, err := parsePlannerOutput(text)
	if err != nil {
		return err
	}
	if len(parsed.Stages) == 0 {
		return errors.New("planner 未产出有效路线")
	}

	// 路线概要
	overview := make([]map[string]interface{}, 0, len(parsed.Stages))
	for _, st := range parsed.Stages {
		overview = append(overview, map[string]interface{}{"title": st.Title})
	}

	path := &learning_db.LearningPath{
		GoalID:   goal.ID,
		UserID:   goal.UserID,
		Overview: overview,
		Status:   "active",
	}
	if err := s.db.Paths.Create(ctx, path); err != nil {
		return err
	}

	// 小目标(扁平化,按顺序)
	var objectives []*learning_db.LearningObjective
	order := 0
	for _, st := range parsed.Stages {
		stageTitle := st.Title
		for _, o := range st.Objectives {
			detail := o.Detail
			status := "pending"
			if order == 0 {
				status = "active" // 第一个小目标设为当前
			}
			objectives = append(objectives, &learning_db.LearningObjective{
				PathID:       path.ID,
				UserID:       goal.UserID,
				StageTitle:   &stageTitle,
				Title:        o.Title,
				Detail:       &detail,
				OrderIndex:   order,
				Status:       status,
				MasteryLevel: "none",
			})
			order++
		}
	}
	if err := s.db.Objectives.BulkCreate(ctx, objectives); err != nil {
		return err
	}

	// 当前小目标 = 第一个
	if len(objectives) > 0 {
		path.CurrentObjectiveID = &objectives[0].ID
		if err := s.db.Paths.Update(ctx, path); err != nil {
			return err
		}
	}
	return nil
}

// collectText 把 agent 事件流收集成完整文本(对齐现有 chat.go 的事件处理)
func collectText(iter *adk.AsyncIterator[*adk.AgentEvent]) (string, error) {
	var sb strings.Builder
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return "", event.Err
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		mo := event.Output.MessageOutput
		if mo.MessageStream != nil {
			for {
				chunk, err := mo.MessageStream.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					return "", err
				}
				if chunk != nil {
					sb.WriteString(chunk.Content)
				}
			}
		} else if mo.Message != nil {
			sb.WriteString(mo.Message.Content)
		}
	}
	return sb.String(), nil
}

// parsePlannerOutput 容错解析 LLM 输出的 JSON(去掉可能的 markdown fence / 前后噪声)
func parsePlannerOutput(text string) (*plannerOutput, error) {
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

	var out plannerOutput
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
