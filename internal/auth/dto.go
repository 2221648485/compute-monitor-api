package auth

import userpkg "compute-monitor-api/internal/user"

// LoginRequest 是登录请求体。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 是登录成功后返回给前端的 token 和当前用户信息。
type LoginResponse struct {
	AccessToken      string               `json:"access_token"`
	RefreshToken     string               `json:"refresh_token"`
	TokenType        string               `json:"token_type"`
	ExpiresIn        int64                `json:"expires_in"`
	RefreshExpiresIn int64                `json:"refresh_expires_in"`
	User             userpkg.UserResponse `json:"user"`
}

// RefreshTokenRequest 是使用 refresh token 换取新 token 的请求体。
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse 是刷新 token 的响应体。
type RefreshTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}

// ChangePasswordRequest 是当前登录用户修改自己密码的请求体。
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

// IDResponse 是只需要返回当前用户 ID 的操作响应。
type IDResponse struct {
	ID int64 `json:"id"`
}
