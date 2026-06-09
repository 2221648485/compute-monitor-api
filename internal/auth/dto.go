package auth

import userpkg "compute-monitor-api/internal/user"

// LoginRequest 是登录请求体。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 是登录成功后返回给前端的 token 和当前用户信息。
type LoginResponse struct {
	AccessToken string               `json:"access_token"`
	TokenType   string               `json:"token_type"`
	ExpiresIn   int64                `json:"expires_in"`
	User        userpkg.UserResponse `json:"user"`
}

// RefreshTokenResponse 是刷新 token 的响应体。
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// ChangePasswordRequest 是当前登录用户修改自己密码的请求体。
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}
