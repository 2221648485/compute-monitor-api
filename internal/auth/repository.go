package auth

import (
	"context"

	userpkg "compute-monitor-api/internal/user"
)

// UserRepository 是认证模块需要的用户数据能力。
// 这里定义小接口，是为了让 auth 只依赖自己真正需要的方法。
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (userpkg.User, error)
	FindByID(ctx context.Context, id int64) (userpkg.User, error)
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	UpdateLastLoginAt(ctx context.Context, id int64) error
	CreateLoginLog(ctx context.Context, log userpkg.LoginLog) error
}
