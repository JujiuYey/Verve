package service

import (
	"context"
	"strings"

	learning_db "sag-wiki/app/learning/models/db"
	"sag-wiki/infrastructure/database"
)

type LearningOrchestratorService struct {
	db *database.DatabaseService
}

func NewLearningOrchestratorService(db *database.DatabaseService) *LearningOrchestratorService {
	return &LearningOrchestratorService{db: db}
}

type OrchestrateLearningResponse struct {
	Intent       string                          `json:"intent"`
	Summary      string                          `json:"summary"`
	HabitSummary string                          `json:"habit_summary"`
	Actions      []LearningOrchestratorAction    `json:"actions"`
	Recent       []*learning_db.LearningExercise `json:"recent"`
}

type LearningOrchestratorAction struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Priority    int     `json:"priority"`
	Label       string  `json:"label"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	GoalID      *string `json:"goal_id,omitempty"`
	ObjectiveID *string `json:"objective_id,omitempty"`
	Intent      *string `json:"intent,omitempty"`
	Reason      string  `json:"reason"`
}

type objectiveCandidate struct {
	goal      *learning_db.LearningGoal
	path      *learning_db.LearningPath
	objective *learning_db.LearningObjective
	profile   *learning_db.LearningProfile
}

func (s *LearningOrchestratorService) Orchestrate(ctx context.Context, userID, intent string) (*OrchestrateLearningResponse, error) {
	intent = strings.TrimSpace(intent)
	goals, err := s.db.Goals.FindActiveByUser(ctx, userID, 8)
	if err != nil {
		return nil, err
	}
	recent, err := s.db.Exercises.FindRecentByUser(ctx, userID, 5)
	if err != nil {
		return nil, err
	}

	candidates := s.loadObjectiveCandidates(ctx, goals)
	actions := make([]LearningOrchestratorAction, 0, 5)

	if intent != "" {
		actions = append(actions, LearningOrchestratorAction{
			ID:          "create-from-intent",
			Type:        "create_goal",
			Priority:    100,
			Label:       "生成新路线",
			Title:       intent,
			Description: "把这个学习意图交给 Planner agent，生成路线后再选择小目标练习。",
			Intent:      &intent,
			Reason:      "用户刚输入了新的学习意图。",
		})
	}

	if current := firstCurrentCandidate(candidates); current != nil {
		actions = append(actions, continueAction(current, 90))
		if review := reviewActionFromProfile(current, 80); review != nil {
			actions = append(actions, *review)
		}
	}

	if weak := reviewActionFromRecent(candidates, recent, 70); weak != nil {
		actions = append(actions, *weak)
	}

	for _, candidate := range candidates {
		if len(actions) >= 5 {
			break
		}
		if candidate.objective == nil || candidate.goal == nil {
			continue
		}
		if hasActionForObjective(actions, candidate.objective.ID) {
			continue
		}
		actions = append(actions, continueAction(&candidate, 50-len(actions)))
	}

	if len(actions) == 0 {
		actions = append(actions, LearningOrchestratorAction{
			ID:          "empty-create-goal",
			Type:        "create_goal",
			Priority:    10,
			Label:       "开始学习",
			Title:       "写下今天想学什么",
			Description: "还没有可继续的学习路线，先输入一个学习目标。",
			Reason:      "当前没有 active 学习目标。",
		})
	}

	return &OrchestrateLearningResponse{
		Intent:       intent,
		Summary:      buildOrchestratorSummary(intent, candidates, recent),
		HabitSummary: buildHabitSummary(candidates, recent),
		Actions:      actions,
		Recent:       recent,
	}, nil
}

func (s *LearningOrchestratorService) loadObjectiveCandidates(ctx context.Context, goals []*learning_db.LearningGoal) []objectiveCandidate {
	candidates := make([]objectiveCandidate, 0, len(goals))
	for _, goal := range goals {
		path, err := s.db.Paths.FindByGoal(ctx, goal.ID)
		if err != nil || path == nil {
			candidates = append(candidates, objectiveCandidate{goal: goal})
			continue
		}

		var objective *learning_db.LearningObjective
		if path.CurrentObjectiveID != nil {
			if obj, err := s.db.Objectives.FindOne(ctx, *path.CurrentObjectiveID); err == nil {
				objective = obj
			}
		}
		if objective == nil {
			objectives, _ := s.db.Objectives.FindByPath(ctx, path.ID)
			objective = firstUnfinishedObjective(objectives)
		}

		var profile *learning_db.LearningProfile
		if p, err := s.db.Profiles.FindByGoal(ctx, goal.ID); err == nil {
			profile = p
		}
		candidates = append(candidates, objectiveCandidate{
			goal:      goal,
			path:      path,
			objective: objective,
			profile:   profile,
		})
	}
	return candidates
}

func firstUnfinishedObjective(objectives []*learning_db.LearningObjective) *learning_db.LearningObjective {
	for _, obj := range objectives {
		if obj.Status != "completed" {
			return obj
		}
	}
	if len(objectives) > 0 {
		return objectives[0]
	}
	return nil
}

func firstCurrentCandidate(candidates []objectiveCandidate) *objectiveCandidate {
	for i := range candidates {
		if candidates[i].objective != nil {
			return &candidates[i]
		}
	}
	for i := range candidates {
		if candidates[i].goal != nil {
			return &candidates[i]
		}
	}
	return nil
}

func continueAction(candidate *objectiveCandidate, priority int) LearningOrchestratorAction {
	goalID := candidate.goal.ID
	if candidate.objective == nil {
		return LearningOrchestratorAction{
			ID:          "open-goal-" + goalID,
			Type:        "open_goal",
			Priority:    priority,
			Label:       "打开路线",
			Title:       candidate.goal.Title,
			Description: "路线还在生成或暂时没有当前小目标，先打开路线查看进度。",
			GoalID:      &goalID,
			Reason:      "该学习目标是最近的 active 目标。",
		}
	}

	objectiveID := candidate.objective.ID
	return LearningOrchestratorAction{
		ID:          "continue-" + objectiveID,
		Type:        "continue_objective",
		Priority:    priority,
		Label:       "推荐继续",
		Title:       candidate.objective.Title,
		Description: objectiveDescription(candidate.objective),
		GoalID:      &goalID,
		ObjectiveID: &objectiveID,
		Reason:      "这是当前学习路线里正在推进的小目标。",
	}
}

func reviewActionFromProfile(candidate *objectiveCandidate, priority int) *LearningOrchestratorAction {
	if candidate.profile == nil || candidate.objective == nil {
		return nil
	}
	nextGoal := strings.TrimSpace(stringValue(candidate.profile.NextGoal))
	if nextGoal == "" {
		return nil
	}
	goalID := candidate.goal.ID
	objectiveID := candidate.objective.ID
	return &LearningOrchestratorAction{
		ID:          "profile-review-" + objectiveID,
		Type:        "review_objective",
		Priority:    priority,
		Label:       "补弱点",
		Title:       nextGoal,
		Description: "基于你的学习画像，先补这个缺口再进入下一轮复述。",
		GoalID:      &goalID,
		ObjectiveID: &objectiveID,
		Reason:      "学习画像记录了上次验证后的 next_goal。",
	}
}

func reviewActionFromRecent(candidates []objectiveCandidate, recent []*learning_db.LearningExercise, priority int) *LearningOrchestratorAction {
	for _, exercise := range recent {
		verdict := stringValue(exercise.Verdict)
		if verdict != "fail" && verdict != "partial" {
			continue
		}
		for i := range candidates {
			candidate := &candidates[i]
			if candidate.objective == nil || candidate.objective.ID != exercise.ObjectiveID {
				continue
			}
			goalID := candidate.goal.ID
			objectiveID := candidate.objective.ID
			title := candidate.objective.Title
			if feedback := strings.TrimSpace(stringValue(exercise.Feedback)); feedback != "" {
				title = feedback
			}
			return &LearningOrchestratorAction{
				ID:          "recent-review-" + exercise.ID,
				Type:        "review_objective",
				Priority:    priority,
				Label:       "复习一下",
				Title:       title,
				Description: "最近一次验证还没完全通过，建议先回到这个小目标复述。",
				GoalID:      &goalID,
				ObjectiveID: &objectiveID,
				Reason:      "最近验证结果是 " + verdict + "。",
			}
		}
	}
	return nil
}

func hasActionForObjective(actions []LearningOrchestratorAction, objectiveID string) bool {
	for _, action := range actions {
		if action.ObjectiveID != nil && *action.ObjectiveID == objectiveID {
			return true
		}
	}
	return false
}

func objectiveDescription(obj *learning_db.LearningObjective) string {
	if obj == nil {
		return ""
	}
	if obj.Detail != nil && strings.TrimSpace(*obj.Detail) != "" {
		return strings.TrimSpace(*obj.Detail)
	}
	if obj.StageTitle != nil && strings.TrimSpace(*obj.StageTitle) != "" {
		return "阶段：" + strings.TrimSpace(*obj.StageTitle)
	}
	return "进入练习工作台，先用自己的话解释这个小目标。"
}

func buildOrchestratorSummary(intent string, candidates []objectiveCandidate, recent []*learning_db.LearningExercise) string {
	if intent != "" {
		return "我会先把你的新意图变成一条学习路线，同时保留当前进度作为继续选项。"
	}
	if len(candidates) == 0 {
		return "还没有可继续的学习路线，先告诉我今天想学什么。"
	}
	if hasRecentWeakExercise(recent) {
		return "你最近有未完全通过的验证，建议先补弱点，再推进新内容。"
	}
	return "已根据当前路线和最近验证，给你排好下一步。"
}

func buildHabitSummary(candidates []objectiveCandidate, recent []*learning_db.LearningExercise) string {
	for _, candidate := range candidates {
		if candidate.profile != nil {
			if habit := strings.TrimSpace(stringValue(candidate.profile.VerificationHabits)); habit != "" {
				return habit
			}
		}
	}
	if len(recent) == 0 {
		return "还没有足够的验证记录，先完成一次费曼复述后我会开始追踪你的学习习惯。"
	}
	passCount := 0
	weakCount := 0
	for _, exercise := range recent {
		switch stringValue(exercise.Verdict) {
		case "pass":
			passCount++
		case "partial", "fail":
			weakCount++
		}
	}
	if weakCount > 0 {
		return "最近验证里有薄弱点，我会优先安排复述和补讲。"
	}
	if passCount > 0 {
		return "最近验证通过较多，可以继续推进下一小目标。"
	}
	return "最近已有练习记录，但还缺少明确判定。"
}

func hasRecentWeakExercise(recent []*learning_db.LearningExercise) bool {
	for _, exercise := range recent {
		verdict := stringValue(exercise.Verdict)
		if verdict == "partial" || verdict == "fail" {
			return true
		}
	}
	return false
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
