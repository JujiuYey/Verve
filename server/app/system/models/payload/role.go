package payload

// 创建角色请求载体
type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

// 更新角色请求载体
type UpdateRoleRequest struct {
	ID          string  `json:"id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}
