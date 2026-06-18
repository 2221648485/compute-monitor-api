package auth

import (
	"errors"
	"log"
	"net/http"

	"compute-monitor-api/internal/requestctx"
	"compute-monitor-api/internal/response"

	"github.com/gin-gonic/gin"
)

// Handler 负责认证相关 HTTP 入参解析和响应输出。
type Handler struct {
	service *Service
}

// NewHandler 创建认证 handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Login 用户登录。
//
// @Summary 用户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.Login(c.Request.Context(), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		writeAuthError(c, err)
		return
	}

	response.OK(c, result)
}

// RefreshToken 使用 refresh token 换取新的 token。
//
// @Summary 刷新 token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新 token 请求"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.RefreshToken(c.Request.Context(), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		writeAuthError(c, err)
		return
	}

	response.OK(c, result)
}

// GetCurrentUser 查询当前登录用户。
//
// @Summary 查询当前登录用户
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/auth/me [get]
func (h *Handler) GetCurrentUser(c *gin.Context) {
	userID, ok := requestctx.CurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Body{Code: http.StatusUnauthorized, Message: ErrorMessage(ErrTokenInvalid)})
		return
	}

	result, err := h.service.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	response.OK(c, result)
}

// ChangePassword 修改当前登录用户密码。
//
// @Summary 修改当前登录用户密码
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Failure 500 {object} response.Body
// @Router /api/admin/auth/password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, ok := requestctx.CurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Body{Code: http.StatusUnauthorized, Message: ErrorMessage(ErrTokenInvalid)})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), userID, req); err != nil {
		writeAuthError(c, err)
		return
	}

	response.OK(c, IDResponse{ID: userID})
}

func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredential), errors.Is(err, ErrUserDisabled), errors.Is(err, ErrTokenInvalid), errors.Is(err, ErrTokenExpired):
		c.JSON(http.StatusUnauthorized, response.Body{Code: http.StatusUnauthorized, Message: ErrorMessage(err)})
	case IsAuthError(err):
		response.BadRequest(c, ErrorMessage(err))
	default:
		log.Printf("auth request failed: error=%v", err)
		response.InternalServerError(c, "authentication request failed")
	}
}
