package payload

// 学习调度请求。Intent 为空时只返回基于历史与当前进度的继续选项。
type OrchestrateLearningRequest struct {
	Intent string `json:"intent"`
}
