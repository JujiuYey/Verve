package payload

// 创建学习会话请求
type CreateSessionRequest struct {
	DocumentID string `json:"document_id"` // Wiki文档ID
}

// 提交解释请求
type ReviewExplanationRequest struct {
	Explanation string `json:"explanation"` // 学习者解释
}
