package user

import (
	"context"
	"errors"
	"strings"

	"compute-monitor-api/internal/page"

	"gorm.io/gorm"
)

// ListQuery 是用户仓储层的列表查询条件。
type ListQuery struct {
	Keyword string
	Role    string
	Status  *int
	Page    page.Query
}

// Repository 定义用户领域需要的数据访问能力。
type Repository interface {
	Create(ctx context.Context, user User) (User, error)
	FindByID(ctx context.Context, id int64) (User, error)
	FindByUsername(ctx context.Context, username string) (User, error)
	List(ctx context.Context, query ListQuery) ([]User, error)
	Count(ctx context.Context, query ListQuery) (int64, error)
	Update(ctx context.Context, user User) (User, error)
	UpdateStatus(ctx context.Context, id int64, status int) error
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	UpdateLastLoginAt(ctx context.Context, id int64) error
	CreateLoginLog(ctx context.Context, log LoginLog) error
}

// MySQLRepository 是基于 GORM 的 MySQL 用户仓储实现。
type MySQLRepository struct {
	db *gorm.DB
}

// NewRepository 创建用户仓储。
func NewRepository(db *gorm.DB) *MySQLRepository {
	return NewMySQLRepository(db)
}

// NewMySQLRepository 创建 MySQL 用户仓储。
func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// Create 新增用户。
func (r *MySQLRepository) Create(ctx context.Context, user User) (User, error) {
	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return User{}, err
	}
	return user, nil
}

// FindByID 根据用户 ID 查询用户。
func (r *MySQLRepository) FindByID(ctx context.Context, id int64) (User, error) {
	var user User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return User{}, err
	}
	return user, nil
}

// FindByUsername 根据用户名查询用户。
func (r *MySQLRepository) FindByUsername(ctx context.Context, username string) (User, error) {
	var user User
	if err := r.db.WithContext(ctx).First(&user, "username = ?", username).Error; err != nil {
		return User{}, err
	}
	return user, nil
}

// List 按条件分页查询用户。
func (r *MySQLRepository) List(ctx context.Context, query ListQuery) ([]User, error) {
	var users []User
	db := applyListQuery(r.db.WithContext(ctx).Model(&User{}), query)
	db = page.Apply(db, query.Page)
	if err := db.Order("id DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Count 统计满足条件的用户数量。
func (r *MySQLRepository) Count(ctx context.Context, query ListQuery) (int64, error) {
	var total int64
	db := applyListQuery(r.db.WithContext(ctx).Model(&User{}), query)
	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// Update 更新用户基础信息。
func (r *MySQLRepository) Update(ctx context.Context, user User) (User, error) {
	updates := map[string]interface{}{
		"display_name": user.DisplayName,
		"email":        user.Email,
		"phone":        user.Phone,
		"role":         user.Role,
	}
	if err := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return User{}, err
	}
	return r.FindByID(ctx, user.ID)
}

// UpdateStatus 修改用户启用状态。
func (r *MySQLRepository) UpdateStatus(ctx context.Context, id int64, status int) error {
	result := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdatePassword 修改用户密码哈希，并递增 token 版本使旧 token 失效。
func (r *MySQLRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	updates := map[string]interface{}{
		"password_hash": passwordHash,
		"token_version": gorm.Expr("token_version + 1"),
	}
	result := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateLastLoginAt 更新用户最后登录时间。
func (r *MySQLRepository) UpdateLastLoginAt(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("last_login_at", gorm.Expr("CURRENT_TIMESTAMP"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CreateLoginLog 写入登录日志。日志失败不应该阻断登录主流程。
func (r *MySQLRepository) CreateLoginLog(ctx context.Context, log LoginLog) error {
	return r.db.WithContext(ctx).Create(&log).Error
}

func applyListQuery(db *gorm.DB, query ListQuery) *gorm.DB {
	keyword := strings.TrimSpace(query.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("username LIKE ? OR display_name LIKE ? OR email LIKE ?", like, like, like)
	}
	if strings.TrimSpace(query.Role) != "" {
		db = db.Where("role = ?", strings.TrimSpace(query.Role))
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	return db
}

func normalizeRecordNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrUserNotFound
	}
	return err
}
