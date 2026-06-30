package payload

type GenerateGuideRequest struct {
	ObjectiveID string `json:"objective_id"`
	Markdown    string `json:"markdown"`
}
