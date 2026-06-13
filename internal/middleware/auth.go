package middleware

import (
	"net/http"
	"strings"

	"compute-monitor-api/internal/auth"
	"compute-monitor-api/internal/requestctx"
	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

const (
	// ContextUserIDKey 是 gin.Context 中保存当前用户 ID 的 key。
	ContextUserIDKey = requestctx.UserIDKey
	// ContextUsernameKey 是 gin.Context 中保存当前用户名的 key。
	ContextUsernameKey = requestctx.UsernameKey
	// ContextRoleKey 是 gin.Context 中保存当前用户角色的 key。
	ContextRoleKey = requestctx.RoleKey
)

// Auth 校验 Authorization: Bearer <token>，并把用户身份写入上下文。
func Auth(tokenManager auth.TokenManager, validators ...auth.TokenValidator) gin.HandlerFunc {
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
		for _, validator := range validators {
			if validator == nil {
				continue
			}
			if err := validator.ValidateAccessClaims(c.Request.Context(), claims); err != nil {
				c.JSON(http.StatusUnauthorized, response.Body{
					Code:    http.StatusUnauthorized,
					Message: auth.ErrorMessage(err),
				})
				c.Abort()
				return
			}
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
	return requestctx.CurrentUserID(c)
}

// CurrentUsername 从 gin.Context 中读取当前用户名。
func CurrentUsername(c *gin.Context) (string, bool) {
	return requestctx.CurrentUsername(c)
}

// CurrentRole 从 gin.Context 中读取当前用户角色。
func CurrentRole(c *gin.Context) (string, bool) {
	return requestctx.CurrentRole(c)
}

func bearerToken(authorization string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(authorization, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(authorization, prefix))
}
