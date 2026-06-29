package tools

import (
	"context"
	"log"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	learning_db "sag-wiki/app/learning/models/db"
	learning_repo "sag-wiki/app/learning/repository"
)

// ==================== Inputs / Outputs ====================

type GetObjectiveInput struct {
	ObjectiveID string `json:"objective_id" jsonschema_description:"小目标 ID"`
}

type GetObjectiveOutput struct {
	Title        string `json:"title"`
	Detail       string `json:"detail"`
	MasteryLevel string `json:"mastery_level"`
	Status       string `json:"status"`
}

type UpdateMasteryInput struct {
	ObjectiveID  string `json:"objective_id" jsonschema_description:"小目标 ID"`
	MasteryLevel string `json:"mastery_level" jsonschema_description:"掌握层级: none/seen/heard/explained/written/verified"`
	Status       string `json:"status" jsonschema_description:"状态: pending/active/completed/review,可留空"`
}

type UpdateMasteryOutput struct {
	Success bool `json:"success"`
}

type RecordExerciseInput struct {
	SessionID    string `json:"session_id" jsonschema_description:"会话 ID"`
	ObjectiveID  string `json:"objective_id" jsonschema_description:"小目标 ID"`
	UserID       string `json:"user_id" jsonschema_description:"用户 ID"`
	Type         string `json:"type" jsonschema_description:"验证类型: explain/choice/cloze/paste_output/code_snippet"`
	Prompt       string `json:"prompt" jsonschema_description:"题目 / 要求"`
	UserAnswer   string `json:"user_answer" jsonschema_description:"用户作答"`
	Verdict      string `json:"verdict" jsonschema_description:"判定: pass/partial/fail"`
	MasteryAfter string `json:"mastery_after" jsonschema_description:"判定后掌握层级"`
	Feedback     string `json:"feedback" jsonschema_description:"反馈"`
}

type RecordExerciseOutput struct {
	Success bool `json:"success"`
}

// ==================== Factory ====================

// NewLearningTools 构造监督/陪练 agent 用的工具(注入 repository)
func NewLearningTools(objRepo *learning_repo.ObjectiveRepository, exRepo *learning_repo.ExerciseRepository) []tool.BaseTool {
	return []tool.BaseTool{
		newGetObjectiveTool(objRepo),
		newUpdateMasteryTool(objRepo),
		newRecordExerciseTool(exRepo),
	}
}

func newGetObjectiveTool(repo *learning_repo.ObjectiveRepository) tool.InvokableTool {
	t, err := utils.InferTool("get_objective", "获取小目标详情(标题/要点/当前掌握层级)",
		func(ctx context.Context, input *GetObjectiveInput) (*GetObjectiveOutput, error) {
			obj, err := repo.FindOne(ctx, input.ObjectiveID)
			if err != nil {
				return nil, err
			}
			detail := ""
			if obj.Detail != nil {
				detail = *obj.Detail
			}
			return &GetObjectiveOutput{
				Title:        obj.Title,
				Detail:       detail,
				MasteryLevel: obj.MasteryLevel,
				Status:       obj.Status,
			}, nil
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newUpdateMasteryTool(repo *learning_repo.ObjectiveRepository) tool.InvokableTool {
	t, err := utils.InferTool("update_mastery", "更新小目标的掌握层级与状态",
		func(ctx context.Context, input *UpdateMasteryInput) (*UpdateMasteryOutput, error) {
			obj, err := repo.FindOne(ctx, input.ObjectiveID)
			if err != nil {
				return nil, err
			}
			obj.MasteryLevel = input.MasteryLevel
			if input.Status != "" {
				obj.Status = input.Status
			}
			if err := repo.Update(ctx, obj); err != nil {
				return nil, err
			}
			return &UpdateMasteryOutput{Success: true}, nil
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newRecordExerciseTool(repo *learning_repo.ExerciseRepository) tool.InvokableTool {
	t, err := utils.InferTool("record_exercise", "记录一次练习与验证结果",
		func(ctx context.Context, input *RecordExerciseInput) (*RecordExerciseOutput, error) {
			ua, verdict, ma, fb := input.UserAnswer, input.Verdict, input.MasteryAfter, input.Feedback
			ex := &learning_db.LearningExercise{
				SessionID:    input.SessionID,
				ObjectiveID:  input.ObjectiveID,
				UserID:       input.UserID,
				Type:         input.Type,
				Prompt:       input.Prompt,
				UserAnswer:   &ua,
				Verdict:      &verdict,
				MasteryAfter: &ma,
				Feedback:     &fb,
			}
			if err := repo.Create(ctx, ex); err != nil {
				return nil, err
			}
			return &RecordExerciseOutput{Success: true}, nil
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}
