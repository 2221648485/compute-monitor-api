package user

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUsernameExists    = errors.New("username exists")
	ErrCannotDisableSelf = errors.New("cannot disable self")
	ErrInvalidUserStatus = errors.New("invalid user status")
	ErrInvalidUserRole   = errors.New("invalid user role")
	ErrPasswordTooWeak   = errors.New("password too weak")
)

// IsUserError 判断是否是用户模块已经定义过的业务错误。
func IsUserError(err error) bool {
	return errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrUsernameExists) ||
		errors.Is(err, ErrCannotDisableSelf) ||
		errors.Is(err, ErrInvalidUserStatus) ||
		errors.Is(err, ErrInvalidUserRole) ||
		errors.Is(err, ErrPasswordTooWeak)
}

// ErrorMessage 把内部业务错误转换成可以返回给前端的提示。
func ErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return "user not found"
	case errors.Is(err, ErrUsernameExists):
		return "username already exists"
	case errors.Is(err, ErrCannotDisableSelf):
		return "cannot disable current user"
	case errors.Is(err, ErrInvalidUserStatus):
		return "invalid user status"
	case errors.Is(err, ErrInvalidUserRole):
		return "invalid user role"
	case errors.Is(err, ErrPasswordTooWeak):
		return "password is too weak"
	default:
		return "user operation failed"
	}
}
