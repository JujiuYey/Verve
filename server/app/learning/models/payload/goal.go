package payload

// 创建学习目标请求(一句话主题驱动)
type CreateGoalRequest struct {
	Title string `json:"title"`
}

// 更新学习目标请求
type UpdateGoalRequest struct {
	ID     string  `json:"id"`
	Title  *string `json:"title"`
	Status *string `json:"status"` // active / archived / completed
}
