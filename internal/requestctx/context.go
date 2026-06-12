package requestctx

import "github.com/gin-gonic/gin"

const (
	// UserIDKey 是 gin.Context 中保存当前用户 ID 的 key。
	UserIDKey = "user_id"
	// UsernameKey 是 gin.Context 中保存当前用户名的 key。
	UsernameKey = "username"
	// RoleKey 是 gin.Context 中保存当前用户角色的 key。
	RoleKey = "role"
)

// CurrentUserID 从 gin.Context 中读取当前登录用户 ID。
func CurrentUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	userID, ok := value.(int64)
	return userID, ok
}

// CurrentUsername 从 gin.Context 中读取当前登录用户名。
func CurrentUsername(c *gin.Context) (string, bool) {
	value, exists := c.Get(UsernameKey)
	if !exists {
		return "", false
	}
	username, ok := value.(string)
	return username, ok
}

// CurrentRole 从 gin.Context 中读取当前登录用户角色。
func CurrentRole(c *gin.Context) (string, bool) {
	value, exists := c.Get(RoleKey)
	if !exists {
		return "", false
	}
	role, ok := value.(string)
	return role, ok
}
