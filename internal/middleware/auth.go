package middleware

import (
	"net/http"
	"strings"

	"compute-monitor-api/internal/auth"
	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

const (
	// ContextUserIDKey 是 gin.Context 中保存当前用户 ID 的 key。
	ContextUserIDKey = "user_id"
	// ContextUsernameKey 是 gin.Context 中保存当前用户名的 key。
	ContextUsernameKey = "username"
	// ContextRoleKey 是 gin.Context 中保存当前用户角色的 key。
	ContextRoleKey = "role"
)

// Auth 校验 Authorization: Bearer <token>，并把用户身份写入上下文。
func Auth(tokenManager auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken := bearerToken(c.GetHeader("Authorization"))
		if rawToken == "" {
			c.JSON(http.StatusUnauthorized, response.Body{
				Code:    http.StatusUnauthorized,
				Message: "missing authorization token",
			})
			c.Abort()
			return
		}

		claims, err := tokenManager.Parse(c.Request.Context(), rawToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.Body{
				Code:    http.StatusUnauthorized,
				Message: auth.ErrorMessage(err),
			})
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Set(ContextRoleKey, claims.Role)
		c.Next()
	}
}

// RequireRole 只允许指定角色访问当前接口。
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		role, _ := c.Get(ContextRoleKey)
		roleValue, _ := role.(string)
		if _, ok := allowed[roleValue]; !ok {
			c.JSON(http.StatusForbidden, response.Body{
				Code:    http.StatusForbidden,
				Message: auth.ErrorMessage(auth.ErrPermissionDenied),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CurrentUserID 从 gin.Context 中读取当前用户 ID。
func CurrentUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, false
	}
	userID, ok := value.(int64)
	return userID, ok
}

// CurrentUsername 从 gin.Context 中读取当前用户名。
func CurrentUsername(c *gin.Context) (string, bool) {
	value, exists := c.Get(ContextUsernameKey)
	if !exists {
		return "", false
	}
	username, ok := value.(string)
	return username, ok
}

// CurrentRole 从 gin.Context 中读取当前用户角色。
func CurrentRole(c *gin.Context) (string, bool) {
	value, exists := c.Get(ContextRoleKey)
	if !exists {
		return "", false
	}
	role, ok := value.(string)
	return role, ok
}

func bearerToken(authorization string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(authorization, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(authorization, prefix))
}
