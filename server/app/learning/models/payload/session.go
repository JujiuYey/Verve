package payload

// 创建学习会话请求
type CreateSessionRequest struct {
	DocumentID string `json:"document_id"` // Wiki文档ID
}

// 提交解释请求
type ReviewExplanationRequest struct {
	RequestID   string `json:"request_id"`  // 请求幂等标识
	Explanation string `json:"explanation"` // 学习者解释
}
