package payload

// 开始一节学习会话请求
type CreateSessionRequest struct {
	ObjectiveID string `json:"objective_id"` // 小目标ID
}

// 提交练习验证请求
type SubmitExerciseRequest struct {
	Type       string `json:"type"`        // explain/choice/cloze/paste_output/code_snippet
	Prompt     string `json:"prompt"`      // 题目 / 要求
	UserAnswer string `json:"user_answer"` // 用户作答
}
