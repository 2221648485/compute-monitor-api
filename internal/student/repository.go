package student

import (
	"context"
	"database/sql"
)

type Repository interface {
	Create(ctx context.Context, student Student) (Student, error)
	FindByID(ctx context.Context, id int64) (Student, error)
}

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Create(ctx context.Context, student Student) (Student, error) {
	const statement = `
		INSERT INTO students (name, age, email, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
	`

	result, err := r.db.ExecContext(ctx, statement, student.Name, student.Age, student.Email)
	if err != nil {
		return Student{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Student{}, err
	}

	student.ID = id
	return student, nil
}

func (r *MySQLRepository) FindByID(ctx context.Context, id int64) (Student, error) {
	const statement = `
		SELECT id, name, age, email, created_at, updated_at
		FROM students
		WHERE id = ?
		LIMIT 1
	`

	var student Student
	err := r.db.QueryRowContext(ctx, statement, id).Scan(
		&student.ID,
		&student.Name,
		&student.Age,
		&student.Email,
		&student.CreatedAt,
		&student.UpdatedAt,
	)
	if err != nil {
		return Student{}, err
	}

	return student, nil
}
