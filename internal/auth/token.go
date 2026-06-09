package auth

import (
	"context"
	"errors"
	"time"

	userpkg "compute-monitor-api/internal/user"

	"github.com/golang-jwt/jwt/v5"
)

const defaultAccessTokenTTL = 2 * time.Hour

// TokenOptions 配置 JWT 签名信息和 token 有效期。
type TokenOptions struct {
	Secret         string
	Issuer         string
	AccessTokenTTL time.Duration
}

// TokenClaims 是 token 解析后得到的用户身份信息。
type TokenClaims struct {
	UserID    int64
	Username  string
	Role      string
	ExpiresAt time.Time
}

// TokenManager 定义 JWT 签发和解析能力。
type TokenManager interface {
	Generate(ctx context.Context, user userpkg.User) (string, int64, error)
	Parse(ctx context.Context, token string) (TokenClaims, error)
}

// JWTManager 负责签发和解析 HS256 JWT access token。
type JWTManager struct {
	secret         []byte
	issuer         string
	accessTokenTTL time.Duration
	now            func() time.Time
}

// jwtClaims 是真正写入 JWT 的载荷结构。
type jwtClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewTokenManager 创建 JWT 管理器，生产环境不要使用默认 secret。
func NewTokenManager(opts TokenOptions) *JWTManager {
	if opts.Secret == "" {
		opts.Secret = "change-me"
	}
	if opts.Issuer == "" {
		opts.Issuer = "compute-monitor-api"
	}
	if opts.AccessTokenTTL <= 0 {
		opts.AccessTokenTTL = defaultAccessTokenTTL
	}
	return &JWTManager{
		secret:         []byte(opts.Secret),
		issuer:         opts.Issuer,
		accessTokenTTL: opts.AccessTokenTTL,
		now:            time.Now,
	}
}

// Generate 签发 JWT，并返回 token 字符串和过期秒数。
func (m *JWTManager) Generate(ctx context.Context, user userpkg.User) (string, int64, error) {
	_ = ctx
	now := m.now()
	expiresAt := now.Add(m.accessTokenTTL)
	claims := jwtClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", 0, err
	}

	return signed, int64(m.accessTokenTTL.Seconds()), nil
}

// Parse 校验 token 签名、签发方和过期时间。
func (m *JWTManager) Parse(ctx context.Context, rawToken string) (TokenClaims, error) {
	_ = ctx
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrTokenInvalid
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return TokenClaims{}, ErrTokenExpired
		}
		return TokenClaims{}, ErrTokenInvalid
	}
	if token == nil || !token.Valid {
		return TokenClaims{}, ErrTokenInvalid
	}

	expiresAt := time.Time{}
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}

	return TokenClaims{
		UserID:    claims.UserID,
		Username:  claims.Username,
		Role:      claims.Role,
		ExpiresAt: expiresAt,
	}, nil
}

// GenerateToken 是函数式调用风格的便捷包装。
func GenerateToken(ctx context.Context, manager TokenManager, user userpkg.User) (string, int64, error) {
	return manager.Generate(ctx, user)
}

// ParseToken 是函数式调用风格的便捷包装。
func ParseToken(ctx context.Context, manager TokenManager, token string) (TokenClaims, error) {
	return manager.Parse(ctx, token)
}
