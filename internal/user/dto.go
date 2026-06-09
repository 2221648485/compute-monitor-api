package user

import "time"

// CreateUserRequest 是创建后台用户的请求体。
type CreateUserRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=64"`
	Password    string `json:"password" binding:"required,min=8,max=72"`
	DisplayName string `json:"display_name" binding:"required,max=64"`
	Email       string `json:"email" binding:"omitempty,email,max=128"`
	Phone       string `json:"phone" binding:"omitempty,max=32"`
	Role        string `json:"role" binding:"required"`
	Status      int    `json:"status"`
}

// UpdateUserRequest 是修改后台用户基础信息的请求体，不在这里修改密码。
type UpdateUserRequest struct {
	DisplayName string `json:"display_name" binding:"required,max=64"`
	Email       string `json:"email" binding:"omitempty,email,max=128"`
	Phone       string `json:"phone" binding:"omitempty,max=32"`
	Role        string `json:"role" binding:"required"`
}

// UpdateUserStatusRequest 是启用或禁用后台用户的请求体。
type UpdateUserStatusRequest struct {
	Status int `json:"status" binding:"required"`
}

// ResetPasswordRequest 是管理员重置用户密码的请求体。
type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// UserListRequest 是后台用户列表查询参数。
type UserListRequest struct {
	Keyword string `form:"keyword"`
	Role    string `form:"role"`
	Status  *int   `form:"status"`
	Page    int    `form:"page"`
	Size    int    `form:"size"`
}

// UserResponse 是对外返回的用户信息，不能包含 password_hash。
type UserResponse struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	Role        string     `json:"role"`
	Status      int        `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// UserListResponse 是用户分页列表返回体。
type UserListResponse struct {
	Items []UserResponse `json:"items"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
}

// ToResponse 把数据库模型转换成接口返回模型。
func ToResponse(u User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		Phone:       u.Phone,
		Role:        u.Role,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// ToResponses 批量转换用户返回模型。
func ToResponses(users []User) []UserResponse {
	items := make([]UserResponse, 0, len(users))
	for _, item := range users {
		items = append(items, ToResponse(item))
	}
	return items
}
