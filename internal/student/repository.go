package student

import (
	"context"

	"gorm.io/gorm"
)

// Repository defines student data access behavior.
// Service depends on this interface instead of a concrete MySQL implementation.
type Repository interface {
	Create(ctx context.Context, student Student) (Student, error)
	FindByID(ctx context.Context, id int64) (Student, error)
}

// MySQLRepository is a GORM-backed Repository implementation.
type MySQLRepository struct {
	db *gorm.DB
}

func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Create(ctx context.Context, student Student) (Student, error) {
	if err := r.db.WithContext(ctx).Create(&student).Error; err != nil {
		return Student{}, err
	}

	return student, nil
}

func (r *MySQLRepository) FindByID(ctx context.Context, id int64) (Student, error) {
	var student Student
	if err := r.db.WithContext(ctx).First(&student, id).Error; err != nil {
		return Student{}, err
	}

	return student, nil
}
