package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Body{
		Code:    0,
		Message: "created",
		Data:    data,
	})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Body{
		Code:    http.StatusBadRequest,
		Message: message,
	})
}

func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Body{
		Code:    http.StatusInternalServerError,
		Message: message,
	})
}
