package student

import (
	"errors"
	"net/http"
	"strconv"

	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

// Handler handles student HTTP requests.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	student, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "failed to create student")
		return
	}

	response.Created(c, ToResponse(student))
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "invalid student id")
		return
	}

	student, err := h.service.GetByID(c.Request.Context(), id)
	if errors.Is(err, ErrStudentNotFound) {
		c.JSON(http.StatusNotFound, response.Body{
			Code:    http.StatusNotFound,
			Message: "student not found",
		})
		return
	}
	if err != nil {
		response.InternalServerError(c, "failed to get student")
		return
	}

	response.OK(c, ToResponse(student))
}
