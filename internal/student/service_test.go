package student

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"
)

type fakeRepository struct {
	createFunc   func(ctx context.Context, student Student) (Student, error)
	findByIDFunc func(ctx context.Context, id int64) (Student, error)
}

func (r fakeRepository) Create(ctx context.Context, student Student) (Student, error) {
	return r.createFunc(ctx, student)
}

func (r fakeRepository) FindByID(ctx context.Context, id int64) (Student, error) {
	return r.findByIDFunc(ctx, id)
}

func TestServiceGetByIDMapsGORMRecordNotFound(t *testing.T) {
	service := NewService(fakeRepository{
		findByIDFunc: func(ctx context.Context, id int64) (Student, error) {
			return Student{}, gorm.ErrRecordNotFound
		},
	})

	_, err := service.GetByID(context.Background(), 1)
	if !errors.Is(err, ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound, got %v", err)
	}
}

func TestServiceCreateTrimsInput(t *testing.T) {
	service := NewService(fakeRepository{
		createFunc: func(ctx context.Context, student Student) (Student, error) {
			if student.Name != "Zhang San" {
				t.Fatalf("expected trimmed name, got %q", student.Name)
			}
			if student.Email != "zhangsan@example.com" {
				t.Fatalf("expected trimmed email, got %q", student.Email)
			}
			return student, nil
		},
	})

	_, err := service.Create(context.Background(), CreateStudentRequest{
		Name:  "  Zhang San  ",
		Age:   20,
		Email: "  zhangsan@example.com  ",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
