package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	userpkg "compute-monitor-api/internal/user"

	"gorm.io/gorm"
)

// Service 负责认证业务逻辑，包含登录、当前用户、修改密码和刷新 token。
type Service struct {
	repository     UserRepository
	passwordHasher PasswordHasher
	tokenManager   TokenManager
	sessionStore   SessionStore
	refreshTTL     time.Duration
}

// NewService 创建认证服务。
func NewService(repository UserRepository, passwordHasher PasswordHasher, tokenManager TokenManager, sessionStore SessionStore, refreshTTL time.Duration) *Service {
	return &Service{
		repository:     repository,
		passwordHasher: passwordHasher,
		tokenManager:   tokenManager,
		sessionStore:   sessionStore,
		refreshTTL:     refreshTTL,
	}
}

// Login 校验用户名密码，创建 Redis 会话，签发 access token 和 refresh token，并记录登录日志。
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

	session, err := s.sessionStore.Create(ctx, user, s.refreshTTL, ip, userAgent)
	if err != nil {
		return LoginResponse{}, err
	}
	token, expiresIn, err := s.tokenManager.Generate(ctx, user, session.SessionID)
	if err != nil {
		_ = s.sessionStore.Delete(ctx, session.SessionID)
		return LoginResponse{}, err
	}
	refreshToken, refreshExpiresIn, err := s.tokenManager.GenerateRefresh(ctx, user, session.SessionID)
	if err != nil {
		_ = s.sessionStore.Delete(ctx, session.SessionID)
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

// RefreshToken 使用 refresh token 换取新的 access token 和 refresh token。
func (s *Service) RefreshToken(ctx context.Context, req RefreshTokenRequest, ip string, userAgent string) (RefreshTokenResponse, error) {
	claims, err := s.tokenManager.ParseRefresh(ctx, strings.TrimSpace(req.RefreshToken))
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	session, err := s.sessionStore.Get(ctx, claims.SessionID)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return RefreshTokenResponse{}, ErrTokenInvalid
		}
		return RefreshTokenResponse{}, err
	}
	if !sameSessionIdentity(session, claims) {
		_ = s.sessionStore.Delete(ctx, claims.SessionID)
		return RefreshTokenResponse{}, ErrTokenInvalid
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
		_ = s.sessionStore.Delete(ctx, claims.SessionID)
		return RefreshTokenResponse{}, ErrTokenInvalid
	}

	newSession, err := s.sessionStore.Rotate(ctx, claims.SessionID, user, s.refreshTTL, ip, userAgent)
	if err != nil {
		return RefreshTokenResponse{}, err
	}
	token, expiresIn, err := s.tokenManager.Generate(ctx, user, newSession.SessionID)
	if err != nil {
		return RefreshTokenResponse{}, err
	}
	refreshToken, refreshExpiresIn, err := s.tokenManager.GenerateRefresh(ctx, user, newSession.SessionID)
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

// ValidateAccessClaims 校验 access token 中的用户状态、token 版本和 Redis 会话是否仍然有效。
func (s *Service) ValidateAccessClaims(ctx context.Context, claims TokenClaims) error {
	session, err := s.sessionStore.Get(ctx, claims.SessionID)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return ErrTokenInvalid
		}
		return err
	}
	if !sameSessionIdentity(session, claims) {
		return ErrTokenInvalid
	}

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

func sameTokenIdentity(user userpkg.User, claims TokenClaims) bool {
	return user.ID == claims.UserID &&
		user.Username == claims.Username &&
		user.Role == claims.Role &&
		user.TokenVersion == claims.TokenVersion
}

func sameSessionIdentity(session RefreshSession, claims TokenClaims) bool {
	return session.SessionID == claims.SessionID &&
		session.UserID == claims.UserID &&
		session.Username == claims.Username &&
		session.Role == claims.Role &&
		session.TokenVersion == claims.TokenVersion
}
