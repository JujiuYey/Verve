package payload

// 开始一节学习会话请求
type CreateSessionRequest struct {
	DocumentID string `json:"document_id"`
}

type ReviewExplanationRequest struct {
	Explanation string `json:"explanation"`
}
