package user

import (
	"errors"
	"net/http"
	"strconv"

	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

// Handler 负责用户管理 HTTP 入参解析和响应输出。
type Handler struct {
	service *Service
}

// NewHandler 创建用户管理 handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// List 查询后台用户列表。
//
// @Summary 查询后台用户列表
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "用户名、姓名或邮箱关键字"
// @Param role query string false "角色：admin/operator/viewer"
// @Param status query int false "状态：1启用，0禁用"
// @Param page query int false "页码，默认1"
// @Param size query int false "每页数量，默认20，最大100"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/users [get]
func (h *Handler) List(c *gin.Context) {
	var req UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.List(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "failed to list users")
		return
	}

	response.OK(c, result)
}

// Create 创建后台用户。
//
// @Summary 创建后台用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUserRequest true "创建用户请求"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/users [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	created, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		writeUserError(c, err)
		return
	}

	response.Created(c, ToResponse(created))
}

// GetByID 查询后台用户详情。
//
// @Summary 查询后台用户详情
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Param userId path int true "用户ID"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 404 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/users/{userId} [get]
func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), userID)
	if err != nil {
		writeUserError(c, err)
		return
	}

	response.OK(c, ToResponse(user))
}

// Update 修改后台用户基础信息。
//
// @Summary 修改后台用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path int true "用户ID"
// @Param request body UpdateUserRequest true "修改用户请求"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 404 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/users/{userId} [put]
func (h *Handler) Update(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updated, err := h.service.Update(c.Request.Context(), userID, req)
	if err != nil {
		writeUserError(c, err)
		return
	}

	response.OK(c, ToResponse(updated))
}

// UpdateStatus 启用或禁用后台用户。
//
// @Summary 修改后台用户状态
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path int true "用户ID"
// @Param request body UpdateUserStatusRequest true "修改状态请求"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 404 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/users/{userId}/status [put]
func (h *Handler) UpdateStatus(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	currentUserID, _ := currentUserID(c)

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), userID, currentUserID, req); err != nil {
		writeUserError(c, err)
		return
	}

	response.OK(c, gin.H{"id": userID, "status": req.Status})
}

// ResetPassword 重置后台用户密码。
//
// @Summary 重置后台用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path int true "用户ID"
// @Param request body ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 404 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/users/{userId}/password [put]
func (h *Handler) ResetPassword(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), userID, req); err != nil {
		writeUserError(c, err)
		return
	}

	response.OK(c, gin.H{"id": userID})
}

func parseUserID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "invalid user id")
		return 0, false
	}
	return id, true
}

func currentUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	userID, ok := value.(int64)
	return userID, ok
}

func writeUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrUserNotFound):
		c.JSON(http.StatusNotFound, response.Body{Code: http.StatusNotFound, Message: ErrorMessage(err)})
	case IsUserError(err):
		response.BadRequest(c, ErrorMessage(err))
	default:
		response.InternalServerError(c, "failed to process user request")
	}
}
