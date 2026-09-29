package handler

import (
	"book-catalog/internal/apperror"
	"book-catalog/internal/middleware"
	"book-catalog/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input service.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("request body must be valid JSON"))
		return
	}

	token, user, err := h.auth.Register(input)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(201, gin.H{"token": token, "user": user})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input service.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("request body must be valid JSON"))
		return
	}

	token, user, err := h.auth.Login(input)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{"token": token, "user": user})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	if userID == 0 {
		c.Error(apperror.Unauthorized("invalid or expired token"))
		return
	}

	user, err := h.auth.Me(userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{"user": user})
}
