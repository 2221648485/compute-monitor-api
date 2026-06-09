package auth

import "errors"

var (
	ErrInvalidCredential = errors.New("invalid username or password")
	ErrUserDisabled      = errors.New("user disabled")
	ErrTokenInvalid      = errors.New("token invalid")
	ErrTokenExpired      = errors.New("token expired")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrPasswordMismatch  = errors.New("password mismatch")
	ErrPasswordTooWeak   = errors.New("password too weak")
)

// IsAuthError 判断 err 是否为认证模块已知业务错误。
func IsAuthError(err error) bool {
	return errors.Is(err, ErrInvalidCredential) ||
		errors.Is(err, ErrUserDisabled) ||
		errors.Is(err, ErrTokenInvalid) ||
		errors.Is(err, ErrTokenExpired) ||
		errors.Is(err, ErrPermissionDenied) ||
		errors.Is(err, ErrPasswordMismatch) ||
		errors.Is(err, ErrPasswordTooWeak)
}

// ErrorMessage 把认证模块内部错误转换成可以返回给前端的提示。
func ErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrInvalidCredential):
		return "invalid username or password"
	case errors.Is(err, ErrUserDisabled):
		return "user disabled"
	case errors.Is(err, ErrTokenInvalid):
		return "invalid token"
	case errors.Is(err, ErrTokenExpired):
		return "token expired"
	case errors.Is(err, ErrPermissionDenied):
		return "permission denied"
	case errors.Is(err, ErrPasswordMismatch):
		return "password confirmation does not match"
	case errors.Is(err, ErrPasswordTooWeak):
		return "password is too weak"
	default:
		return "authentication failed"
	}
}
