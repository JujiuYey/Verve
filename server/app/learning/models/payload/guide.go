package payload

type GenerateGuideRequest struct {
	ObjectiveID string `json:"objective_id"` // 小目标ID
	Markdown    string `json:"markdown"`     // markdown内容
}
