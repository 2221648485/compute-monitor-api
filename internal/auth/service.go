package auth

import (
	"context"
	"errors"
	"strings"

	userpkg "compute-monitor-api/internal/user"

	"gorm.io/gorm"
)

// Service 负责认证业务逻辑，包含登录、当前用户和修改密码。
type Service struct {
	repository     UserRepository
	passwordHasher PasswordHasher
	tokenManager   TokenManager
}

// NewService 创建认证服务。
func NewService(repository UserRepository, passwordHasher PasswordHasher, tokenManager TokenManager) *Service {
	return &Service{
		repository:     repository,
		passwordHasher: passwordHasher,
		tokenManager:   tokenManager,
	}
}

// Login 校验用户名密码，签发 JWT，并记录登录日志。
func (s *Service) Login(ctx context.Context, req LoginRequest, ip string, userAgent string) (LoginResponse, error) {
	username := strings.TrimSpace(req.Username)
	user, err := s.repository.FindByUsername(ctx, username)
	if err != nil {
		s.writeLoginLog(ctx, 0, username, ip, userAgent, false, "invalid credential")
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, userpkg.ErrUserNotFound) {
			return LoginResponse{}, ErrInvalidCredential
		}
		return LoginResponse{}, err
	}

	if !user.IsEnabled() {
		s.writeLoginLog(ctx, user.ID, user.Username, ip, userAgent, false, "user disabled")
		return LoginResponse{}, ErrUserDisabled
	}
	if !s.passwordHasher.Compare(req.Password, user.PasswordHash) {
		s.writeLoginLog(ctx, user.ID, user.Username, ip, userAgent, false, "invalid credential")
		return LoginResponse{}, ErrInvalidCredential
	}

	token, expiresIn, err := s.tokenManager.Generate(ctx, user)
	if err != nil {
		return LoginResponse{}, err
	}
	_ = s.repository.UpdateLastLoginAt(ctx, user.ID)
	s.writeLoginLog(ctx, user.ID, user.Username, ip, userAgent, true, "")

	return LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        userpkg.ToResponse(user),
	}, nil
}

// GetCurrentUser 查询当前登录用户信息。
func (s *Service) GetCurrentUser(ctx context.Context, userID int64) (userpkg.UserResponse, error) {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, userpkg.ErrUserNotFound) {
			return userpkg.UserResponse{}, ErrInvalidCredential
		}
		return userpkg.UserResponse{}, err
	}
	if !user.IsEnabled() {
		return userpkg.UserResponse{}, ErrUserDisabled
	}
	return userpkg.ToResponse(user), nil
}

// ChangePassword 修改当前登录用户自己的密码。
func (s *Service) ChangePassword(ctx context.Context, userID int64, req ChangePasswordRequest) error {
	if req.NewPassword != req.ConfirmPassword {
		return ErrPasswordMismatch
	}
	if len(req.NewPassword) < 8 || len(req.NewPassword) > 72 {
		return ErrPasswordTooWeak
	}

	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, userpkg.ErrUserNotFound) {
			return ErrInvalidCredential
		}
		return err
	}
	if !s.passwordHasher.Compare(req.OldPassword, user.PasswordHash) {
		return ErrInvalidCredential
	}

	passwordHash, err := s.passwordHasher.Hash(req.NewPassword)
	if err != nil {
		return err
	}
	return s.repository.UpdatePassword(ctx, userID, passwordHash)
}

func (s *Service) writeLoginLog(ctx context.Context, userID int64, username string, ip string, userAgent string, success bool, reason string) {
	_ = s.repository.CreateLoginLog(ctx, userpkg.LoginLog{
		UserID:    userID,
		Username:  username,
		IP:        ip,
		UserAgent: userAgent,
		Success:   success,
		Reason:    reason,
	})
}
