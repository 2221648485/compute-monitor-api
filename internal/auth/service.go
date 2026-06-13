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
	refreshToken, refreshExpiresIn, err := s.tokenManager.GenerateRefresh(ctx, user)
	if err != nil {
		return LoginResponse{}, err
	}
	_ = s.repository.UpdateLastLoginAt(ctx, user.ID)
	s.writeLoginLog(ctx, user.ID, user.Username, ip, userAgent, true, "")

	return LoginResponse{
		AccessToken:      token,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
		User:             userpkg.ToResponse(user),
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

// RefreshToken 使用 refresh token 换取新的 access token 和 refresh token。
func (s *Service) RefreshToken(ctx context.Context, req RefreshTokenRequest) (RefreshTokenResponse, error) {
	claims, err := s.tokenManager.ParseRefresh(ctx, strings.TrimSpace(req.RefreshToken))
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	user, err := s.repository.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, userpkg.ErrUserNotFound) {
			return RefreshTokenResponse{}, ErrInvalidCredential
		}
		return RefreshTokenResponse{}, err
	}
	if !user.IsEnabled() {
		return RefreshTokenResponse{}, ErrUserDisabled
	}
	if !sameTokenIdentity(user, claims) {
		return RefreshTokenResponse{}, ErrTokenInvalid
	}

	token, expiresIn, err := s.tokenManager.Generate(ctx, user)
	if err != nil {
		return RefreshTokenResponse{}, err
	}
	refreshToken, refreshExpiresIn, err := s.tokenManager.GenerateRefresh(ctx, user)
	if err != nil {
		return RefreshTokenResponse{}, err
	}
	return RefreshTokenResponse{
		AccessToken:      token,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	}, nil
}

// ValidateAccessClaims 校验 access token 中的用户状态和 token 版本是否仍然有效。
func (s *Service) ValidateAccessClaims(ctx context.Context, claims TokenClaims) error {
	user, err := s.repository.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, userpkg.ErrUserNotFound) {
			return ErrTokenInvalid
		}
		return err
	}
	if !user.IsEnabled() {
		return ErrUserDisabled
	}
	if !sameTokenIdentity(user, claims) {
		return ErrTokenInvalid
	}
	return nil
}

func sameTokenIdentity(user userpkg.User, claims TokenClaims) bool {
	return user.ID == claims.UserID &&
		user.Username == claims.Username &&
		user.Role == claims.Role &&
		user.TokenVersion == claims.TokenVersion
}
