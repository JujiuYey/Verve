package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/cloudwego/eino/adk"

	learning_db "verve/app/learning/models/db"
	"verve/infrastructure/database"
	"verve/infrastructure/llm"
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
	fallback := buildRuleBasedOrchestratorResponse(intent, candidates, recent)
	agentResult, err := s.runOrchestratorAgent(ctx, intent, candidates, recent, fallback.Actions)
	if err != nil {
		log.Printf("⚠️ Learning Orchestrator agent 失败，使用规则兜底: user_id=%s intent_chars=%d err=%v", userID, len(intent), err)
		return fallback, nil
	}

	agentResult.Intent = intent
	agentResult.Recent = recent
	return agentResult, nil
}

func buildRuleBasedOrchestratorResponse(intent string, candidates []objectiveCandidate, recent []*learning_db.LearningExercise) *OrchestrateLearningResponse {
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
	}
}

type orchestratorAgentOutput struct {
	Summary      string                       `json:"summary"`
	HabitSummary string                       `json:"habit_summary"`
	Actions      []LearningOrchestratorAction `json:"actions"`
}

type orchestratorQuery struct {
	Intent           string                        `json:"intent"`
	CurrentLearning  []orchestratorLearningContext `json:"current_learning"`
	RecentExercises  []orchestratorExerciseContext `json:"recent_exercises"`
	CandidateActions []LearningOrchestratorAction  `json:"candidate_actions"`
}

type orchestratorLearningContext struct {
	GoalID             string   `json:"goal_id"`
	GoalTitle          string   `json:"goal_title"`
	GoalSource         string   `json:"goal_source"`
	ObjectiveID        string   `json:"objective_id,omitempty"`
	ObjectiveTitle     string   `json:"objective_title,omitempty"`
	ObjectiveDetail    string   `json:"objective_detail,omitempty"`
	ObjectiveStatus    string   `json:"objective_status,omitempty"`
	MasteryLevel       string   `json:"mastery_level,omitempty"`
	CurrentLevel       string   `json:"current_level,omitempty"`
	CompletedTopics    []string `json:"completed_topics,omitempty"`
	WeakPoints         []string `json:"weak_points,omitempty"`
	VerificationHabits string   `json:"verification_habits,omitempty"`
	NextGoal           string   `json:"next_goal,omitempty"`
}

type orchestratorExerciseContext struct {
	ID           string `json:"id"`
	ObjectiveID  string `json:"objective_id"`
	Type         string `json:"type"`
	Prompt       string `json:"prompt"`
	Verdict      string `json:"verdict,omitempty"`
	MasteryAfter string `json:"mastery_after,omitempty"`
	Feedback     string `json:"feedback,omitempty"`
	CreatedAt    string `json:"created_at"`
}

func (s *LearningOrchestratorService) runOrchestratorAgent(
	ctx context.Context,
	intent string,
	candidates []objectiveCandidate,
	recent []*learning_db.LearningExercise,
	candidateActions []LearningOrchestratorAction,
) (*OrchestrateLearningResponse, error) {
	agent, err := llm.NewOrchestratorAgent(ctx)
	if err != nil {
		return nil, err
	}

	query, err := buildOrchestratorAgentQuery(intent, candidates, recent, candidateActions)
	if err != nil {
		return nil, err
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})
	text, err := collectText(runner.Query(ctx, query))
	if err != nil {
		return nil, err
	}

	output, err := parseOrchestratorAgentOutput(text)
	if err != nil {
		return nil, err
	}
	actions := normalizeAgentActions(output.Actions, candidateActions)
	if len(actions) == 0 {
		return nil, fmt.Errorf("orchestrator agent 未返回有效动作")
	}
	return &OrchestrateLearningResponse{
		Summary:      firstNonEmpty(output.Summary, buildOrchestratorSummary(intent, candidates, recent)),
		HabitSummary: firstNonEmpty(output.HabitSummary, buildHabitSummary(candidates, recent)),
		Actions:      actions,
	}, nil
}

func buildOrchestratorAgentQuery(
	intent string,
	candidates []objectiveCandidate,
	recent []*learning_db.LearningExercise,
	candidateActions []LearningOrchestratorAction,
) (string, error) {
	query := orchestratorQuery{
		Intent:           intent,
		CurrentLearning:  buildLearningContexts(candidates),
		RecentExercises:  buildExerciseContexts(recent),
		CandidateActions: candidateActions,
	}
	data, err := json.Marshal(query)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func buildLearningContexts(candidates []objectiveCandidate) []orchestratorLearningContext {
	contexts := make([]orchestratorLearningContext, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.goal == nil {
			continue
		}
		item := orchestratorLearningContext{
			GoalID:     candidate.goal.ID,
			GoalTitle:  candidate.goal.Title,
			GoalSource: candidate.goal.Source,
		}
		if candidate.objective != nil {
			item.ObjectiveID = candidate.objective.ID
			item.ObjectiveTitle = candidate.objective.Title
			item.ObjectiveDetail = stringValue(candidate.objective.Detail)
			item.ObjectiveStatus = candidate.objective.Status
			item.MasteryLevel = candidate.objective.MasteryLevel
		}
		if candidate.profile != nil {
			item.CurrentLevel = stringValue(candidate.profile.CurrentLevel)
			item.CompletedTopics = candidate.profile.CompletedTopics
			item.WeakPoints = candidate.profile.WeakPoints
			item.VerificationHabits = stringValue(candidate.profile.VerificationHabits)
			item.NextGoal = stringValue(candidate.profile.NextGoal)
		}
		contexts = append(contexts, item)
	}
	return contexts
}

func buildExerciseContexts(recent []*learning_db.LearningExercise) []orchestratorExerciseContext {
	contexts := make([]orchestratorExerciseContext, 0, len(recent))
	for _, exercise := range recent {
		contexts = append(contexts, orchestratorExerciseContext{
			ID:           exercise.ID,
			ObjectiveID:  exercise.ObjectiveID,
			Type:         exercise.Type,
			Prompt:       truncateForOrchestrator(exercise.Prompt, 240),
			Verdict:      stringValue(exercise.Verdict),
			MasteryAfter: stringValue(exercise.MasteryAfter),
			Feedback:     truncateForOrchestrator(stringValue(exercise.Feedback), 300),
			CreatedAt:    exercise.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return contexts
}

func parseOrchestratorAgentOutput(text string) (*orchestratorAgentOutput, error) {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var lastErr error
	for _, candidate := range jsonObjectCandidates(text) {
		var output orchestratorAgentOutput
		if err := json.Unmarshal([]byte(candidate), &output); err != nil {
			lastErr = err
			continue
		}
		if output.Summary == "" && output.HabitSummary == "" && len(output.Actions) == 0 {
			lastErr = fmt.Errorf("JSON 对象不是 Orchestrator 输出")
			continue
		}
		return &output, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("未找到 JSON 对象")
}

func normalizeAgentActions(agentActions []LearningOrchestratorAction, candidateActions []LearningOrchestratorAction) []LearningOrchestratorAction {
	allowed := make(map[string]LearningOrchestratorAction, len(candidateActions))
	for _, action := range candidateActions {
		allowed[action.ID] = action
	}

	seen := make(map[string]bool, len(agentActions))
	actions := make([]LearningOrchestratorAction, 0, len(agentActions))
	for _, agentAction := range agentActions {
		base, ok := allowed[agentAction.ID]
		if !ok || seen[agentAction.ID] {
			continue
		}
		seen[agentAction.ID] = true
		base.Priority = clampPriority(agentAction.Priority, base.Priority)
		base.Label = firstNonEmpty(agentAction.Label, base.Label)
		base.Title = firstNonEmpty(agentAction.Title, base.Title)
		base.Description = firstNonEmpty(agentAction.Description, base.Description)
		base.Reason = firstNonEmpty(agentAction.Reason, base.Reason)
		actions = append(actions, base)
	}

	sort.SliceStable(actions, func(i, j int) bool {
		return actions[i].Priority > actions[j].Priority
	})
	if len(actions) > 5 {
		actions = actions[:5]
	}
	return actions
}

func clampPriority(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	if value > 100 {
		return 100
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func truncateForOrchestrator(text string, limit int) string {
	text = strings.TrimSpace(text)
	if len(text) <= limit {
		return text
	}
	return text[:limit] + "...(truncated)"
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
