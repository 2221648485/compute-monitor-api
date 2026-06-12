package user

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

// BootstrapAdminOptions 是初始化第一个后台管理员账号的配置。
type BootstrapAdminOptions struct {
	Enabled     bool
	Username    string
	Password    string
	DisplayName string
	Email       string
	Role        string
	Status      int
}

// EnsureBootstrapAdmin 确保系统里有一个可登录的初始管理员。
// 这个逻辑只用于系统初始化，不应该替代正式的用户管理接口。
func EnsureBootstrapAdmin(ctx context.Context, repository Repository, hasher PasswordHasher, opts BootstrapAdminOptions) error {
	if !opts.Enabled {
		return nil
	}

	username := strings.TrimSpace(opts.Username)
	if username == "" {
		username = "admin"
	}
	if _, err := repository.FindByUsername(ctx, username); err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) && !errors.Is(err, ErrUserNotFound) {
		return err
	}

	password := opts.Password
	if password == "" {
		password = "Admin@123456"
	}
	if err := validatePassword(password); err != nil {
		return err
	}

	role := strings.TrimSpace(opts.Role)
	if role == "" {
		role = RoleAdmin
	}
	if err := validateRole(role); err != nil {
		return err
	}

	status := opts.Status
	if status != UserStatusDisabled {
		status = UserStatusEnabled
	}

	displayName := strings.TrimSpace(opts.DisplayName)
	if displayName == "" {
		displayName = "系统管理员"
	}

	passwordHash, err := hasher.Hash(password)
	if err != nil {
		return err
	}

	_, err = repository.Create(ctx, User{
		Username:     username,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		Email:        strings.TrimSpace(opts.Email),
		Role:         role,
		Status:       status,
	})
	return err
}
