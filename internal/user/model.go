package user

import "time"

const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
)

const (
	UserStatusEnabled  = 1
	UserStatusDisabled = 0
)

// User 是后台管理用户表模型，认证模块和用户管理模块都复用这一份模型。
type User struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username     string     `gorm:"column:username;type:varchar(64);uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	DisplayName  string     `gorm:"column:display_name;type:varchar(64);not null" json:"display_name"`
	Email        string     `gorm:"column:email;type:varchar(128)" json:"email"`
	Phone        string     `gorm:"column:phone;type:varchar(32)" json:"phone"`
	Role         string     `gorm:"column:role;type:varchar(32);not null" json:"role"`
	Status       int        `gorm:"column:status;not null" json:"status"`
	TokenVersion int        `gorm:"column:token_version;not null;default:1" json:"-"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定 GORM 映射到后台用户表。
func (User) TableName() string {
	return "admin_users"
}

// IsEnabled 判断用户是否允许登录和访问后台接口。
func (u User) IsEnabled() bool {
	return u.Status == UserStatusEnabled
}

// IsAdmin 判断用户是否是管理员角色。
func (u User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// LoginLog 记录用户登录行为，后续可以用于审计、安全分析和后台登录日志查询。
type LoginLog struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64     `gorm:"column:user_id;index;not null" json:"user_id"`
	Username  string    `gorm:"column:username;type:varchar(64);not null" json:"username"`
	IP        string    `gorm:"column:ip;type:varchar(64)" json:"ip"`
	UserAgent string    `gorm:"column:user_agent;type:varchar(512)" json:"user_agent"`
	Success   bool      `gorm:"column:success;not null" json:"success"`
	Reason    string    `gorm:"column:reason;type:varchar(255)" json:"reason"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 指定 GORM 映射到登录日志表。
func (LoginLog) TableName() string {
	return "admin_login_logs"
}
