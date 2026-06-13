package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	userpkg "compute-monitor-api/internal/user"

	"github.com/redis/go-redis/v9"
)

const refreshSessionKeyPrefix = "auth:session:"

// RefreshSession 是 Redis 中保存的登录会话。
type RefreshSession struct {
	SessionID    string    `json:"session_id"`
	UserID       int64     `json:"user_id"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	TokenVersion int       `json:"token_version"`
	IP           string    `json:"ip"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// SessionStore 定义 refresh token 会话存储能力。
type SessionStore interface {
	Create(ctx context.Context, user userpkg.User, ttl time.Duration, ip string, userAgent string) (RefreshSession, error)
	Get(ctx context.Context, sessionID string) (RefreshSession, error)
	Rotate(ctx context.Context, oldSessionID string, user userpkg.User, ttl time.Duration, ip string, userAgent string) (RefreshSession, error)
	Delete(ctx context.Context, sessionID string) error
}

// RedisSessionStore 使用 Redis 保存 refresh token 会话。
type RedisSessionStore struct {
	client *redis.Client
	now    func() time.Time
}

// NewRedisSessionStore 创建 Redis 会话存储。
func NewRedisSessionStore(client *redis.Client) *RedisSessionStore {
	return &RedisSessionStore{
		client: client,
		now:    time.Now,
	}
}

// Create 创建新的登录会话。
func (s *RedisSessionStore) Create(ctx context.Context, user userpkg.User, ttl time.Duration, ip string, userAgent string) (RefreshSession, error) {
	sessionID, err := newSessionID()
	if err != nil {
		return RefreshSession{}, err
	}
	session := s.newSession(sessionID, user, ttl, ip, userAgent)
	if err := s.save(ctx, session, ttl); err != nil {
		return RefreshSession{}, err
	}
	return session, nil
}

// Get 查询登录会话。
func (s *RedisSessionStore) Get(ctx context.Context, sessionID string) (RefreshSession, error) {
	data, err := s.client.Get(ctx, refreshSessionKey(sessionID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return RefreshSession{}, ErrSessionNotFound
		}
		return RefreshSession{}, err
	}

	var session RefreshSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return RefreshSession{}, err
	}
	return session, nil
}

// Rotate 删除旧会话并创建新会话，避免 refresh token 被重复使用。
func (s *RedisSessionStore) Rotate(ctx context.Context, oldSessionID string, user userpkg.User, ttl time.Duration, ip string, userAgent string) (RefreshSession, error) {
	if _, err := s.client.GetDel(ctx, refreshSessionKey(oldSessionID)).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			return RefreshSession{}, ErrSessionNotFound
		}
		return RefreshSession{}, err
	}

	sessionID, err := newSessionID()
	if err != nil {
		return RefreshSession{}, err
	}
	session := s.newSession(sessionID, user, ttl, ip, userAgent)
	payload, err := json.Marshal(session)
	if err != nil {
		return RefreshSession{}, err
	}
	if err := s.client.Set(ctx, refreshSessionKey(session.SessionID), payload, ttl).Err(); err != nil {
		return RefreshSession{}, err
	}
	return session, nil
}

// Delete 删除登录会话，后续可用于退出登录。
func (s *RedisSessionStore) Delete(ctx context.Context, sessionID string) error {
	return s.client.Del(ctx, refreshSessionKey(sessionID)).Err()
}

func (s *RedisSessionStore) newSession(sessionID string, user userpkg.User, ttl time.Duration, ip string, userAgent string) RefreshSession {
	now := s.now()
	return RefreshSession{
		SessionID:    sessionID,
		UserID:       user.ID,
		Username:     user.Username,
		Role:         user.Role,
		TokenVersion: user.TokenVersion,
		IP:           ip,
		UserAgent:    userAgent,
		CreatedAt:    now,
		ExpiresAt:    now.Add(ttl),
	}
}

func (s *RedisSessionStore) save(ctx context.Context, session RefreshSession, ttl time.Duration) error {
	payload, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, refreshSessionKey(session.SessionID), payload, ttl).Err()
}

func newSessionID() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func refreshSessionKey(sessionID string) string {
	return fmt.Sprintf("%s%s", refreshSessionKeyPrefix, sessionID)
}
