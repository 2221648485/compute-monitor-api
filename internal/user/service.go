package user

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

const (
	defaultPage = 1
	defaultSize = 20
	maxSize     = 100
)

// PasswordHasher 是用户服务需要的密码哈希能力，由 auth.BcryptPasswordHasher 实现。
type PasswordHasher interface {
	Hash(password string) (string, error)
}

// Service 负责用户管理业务规则，handler 不直接操作数据库。
type Service struct {
	repository     Repository
	passwordHasher PasswordHasher
}

// NewService 创建用户服务。
func NewService(repository Repository, passwordHasher PasswordHasher) *Service {
	return &Service{
		repository:     repository,
		passwordHasher: passwordHasher,
	}
}

// Create 创建后台用户。
func (s *Service) Create(ctx context.Context, req CreateUserRequest) (User, error) {
	if err := validateRole(req.Role); err != nil {
		return User{}, err
	}
	if err := validateStatus(req.Status); err != nil {
		return User{}, err
	}
	if err := validatePassword(req.Password); err != nil {
		return User{}, err
	}

	username := strings.TrimSpace(req.Username)
	if _, err := s.repository.FindByUsername(ctx, username); err == nil {
		return User{}, ErrUsernameExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) && !errors.Is(err, ErrUserNotFound) {
		return User{}, err
	}

	passwordHash, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return User{}, err
	}

	user := User{
		Username:     username,
		PasswordHash: passwordHash,
		DisplayName:  strings.TrimSpace(req.DisplayName),
		Email:        strings.TrimSpace(req.Email),
		Phone:        strings.TrimSpace(req.Phone),
		Role:         strings.TrimSpace(req.Role),
		Status:       req.Status,
	}
	return s.repository.Create(ctx, user)
}

// GetByID 查询用户详情。
func (s *Service) GetByID(ctx context.Context, id int64) (User, error) {
	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return User{}, normalizeRecordNotFound(err)
	}
	return user, nil
}

// List 分页查询用户列表。
func (s *Service) List(ctx context.Context, req UserListRequest) (UserListResponse, error) {
	page, size := normalizePagination(req.Page, req.Size)
	query := ListQuery{
		Keyword: strings.TrimSpace(req.Keyword),
		Role:    strings.TrimSpace(req.Role),
		Status:  req.Status,
		Offset:  (page - 1) * size,
		Limit:   size,
	}

	items, err := s.repository.List(ctx, query)
	if err != nil {
		return UserListResponse{}, err
	}
	total, err := s.repository.Count(ctx, query)
	if err != nil {
		return UserListResponse{}, err
	}

	return UserListResponse{
		Items: ToResponses(items),
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

// Update 修改用户基础信息。
func (s *Service) Update(ctx context.Context, id int64, req UpdateUserRequest) (User, error) {
	if err := validateRole(req.Role); err != nil {
		return User{}, err
	}
	current, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return User{}, normalizeRecordNotFound(err)
	}

	current.DisplayName = strings.TrimSpace(req.DisplayName)
	current.Email = strings.TrimSpace(req.Email)
	current.Phone = strings.TrimSpace(req.Phone)
	current.Role = strings.TrimSpace(req.Role)
	updated, err := s.repository.Update(ctx, current)
	if err != nil {
		return User{}, normalizeRecordNotFound(err)
	}
	return updated, nil
}

// UpdateStatus 修改用户启用状态。
func (s *Service) UpdateStatus(ctx context.Context, id int64, currentUserID int64, req UpdateUserStatusRequest) error {
	if err := validateStatus(req.Status); err != nil {
		return err
	}
	if id == currentUserID && req.Status == UserStatusDisabled {
		return ErrCannotDisableSelf
	}
	if err := s.repository.UpdateStatus(ctx, id, req.Status); err != nil {
		return normalizeRecordNotFound(err)
	}
	return nil
}

// ResetPassword 由管理员重置用户密码。
func (s *Service) ResetPassword(ctx context.Context, id int64, req ResetPasswordRequest) error {
	if err := validatePassword(req.Password); err != nil {
		return err
	}
	passwordHash, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return err
	}
	if err := s.repository.UpdatePassword(ctx, id, passwordHash); err != nil {
		return normalizeRecordNotFound(err)
	}
	return nil
}

func validateRole(role string) error {
	switch strings.TrimSpace(role) {
	case RoleAdmin, RoleOperator, RoleViewer:
		return nil
	default:
		return ErrInvalidUserRole
	}
}

func validateStatus(status int) error {
	switch status {
	case UserStatusEnabled, UserStatusDisabled:
		return nil
	default:
		return ErrInvalidUserStatus
	}
}

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return ErrPasswordTooWeak
	}
	return nil
}

func normalizePagination(page int, size int) (int, int) {
	if page <= 0 {
		page = defaultPage
	}
	if size <= 0 {
		size = defaultSize
	}
	if size > maxSize {
		size = maxSize
	}
	return page, size
}
