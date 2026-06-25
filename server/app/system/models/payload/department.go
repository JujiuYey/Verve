package payload

// 创建部门请求载体
type CreateDepartmentRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	ParentID    *string `json:"parent_id"`
}

// 更新部门请求载体
type UpdateDepartmentRequest struct {
	ID          string  `json:"id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	ParentID    *string `json:"parent_id"`
}
