package student

type CreateStudentRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=64"`
	Age   int    `json:"age" binding:"required,gte=1,lte=150"`
	Email string `json:"email" binding:"required,email,max=128"`
}

type StudentResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

func ToResponse(student Student) StudentResponse {
	return StudentResponse{
		ID:    student.ID,
		Name:  student.Name,
		Age:   student.Age,
		Email: student.Email,
	}
}
