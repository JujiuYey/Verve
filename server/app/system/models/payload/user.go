package payload

// 创建用户请求载体
type CreateUserRequest struct {
	Username            string   `json:"username"`
	Email               string   `json:"email"`
	FullName            *string  `json:"full_name"`
	Avatar              *string  `json:"avatar"`
	PrimaryDepartmentID string   `json:"primary_department_id"`
	RoleIDs             []string `json:"role_ids"`
}

// 更新用户请求载体
type UpdateUserRequest struct {
	ID                  string   `json:"id"`
	Email               string   `json:"email"`
	FullName            *string  `json:"full_name"`
	Avatar              *string  `json:"avatar"`
	Status              string   `json:"status"`
	PrimaryDepartmentID string   `json:"primary_department_id"`
	RoleIDs             []string `json:"role_ids"`
}

// 修改密码请求载体
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
