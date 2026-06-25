package payload

// 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 可以是用户名或邮箱
	Password string `json:"password" binding:"required"`
}

// 登录响应
type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         UserInfo `json:"user"`
}

// 用户信息
type UserInfo struct {
	ID                  string   `json:"id"`
	Username            string   `json:"username"`
	Email               string   `json:"email"`
	FullName            *string  `json:"full_name"`
	Avatar              *string  `json:"avatar"`
	Status              string   `json:"status"`
	Roles               []string `json:"roles"`
	PrimaryDepartmentID string   `json:"primary_department_id"`
}

// 刷新 token 请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// 刷新 token 响应
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
