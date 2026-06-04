package student

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var ErrStudentNotFound = errors.New("student not found")

// Service contains student business logic.
type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, req CreateStudentRequest) (Student, error) {
	student := Student{
		Name:  strings.TrimSpace(req.Name),
		Age:   req.Age,
		Email: strings.TrimSpace(req.Email),
	}

	return s.repository.Create(ctx, student)
}

func (s *Service) GetByID(ctx context.Context, id int64) (Student, error) {
	student, err := s.repository.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Student{}, ErrStudentNotFound
	}
	if err != nil {
		return Student{}, err
	}

	return student, nil
}
