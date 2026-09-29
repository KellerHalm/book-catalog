package handler

import (
	"strconv"

	"book-catalog/internal/apperror"
	"book-catalog/internal/middleware"
	"book-catalog/internal/repository"
	"book-catalog/internal/service"

	"github.com/gin-gonic/gin"
)

type UserBookHandler struct {
	userBooks *service.UserBookService
}

func NewUserBookHandler(userBooks *service.UserBookService) *UserBookHandler {
	return &UserBookHandler{userBooks: userBooks}
}

func (h *UserBookHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "12"))

	filter := repository.UserBookFilter{
		Status:  c.DefaultQuery("status", "all"),
		Page:    page,
		PerPage: perPage,
	}

	items, meta, err := h.userBooks.List(middleware.CurrentUserID(c), filter)
	if err != nil {
		c.Error(err)
		return
	}

	stats, err := h.userBooks.Stats(middleware.CurrentUserID(c))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{
		"items": items,
		"meta":  meta,
		"stats": stats,
	})
}

func (h *UserBookHandler) Add(c *gin.Context) {
	bookID, err := strconv.ParseUint(c.Param("bookId"), 10, 64)
	if err != nil || bookID == 0 {
		c.Error(apperror.BadRequest("bookId must be a positive number"))
		return
	}

	item, err := h.userBooks.Add(middleware.CurrentUserID(c), uint(bookID))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(201, gin.H{"item": item})
}

func (h *UserBookHandler) UpdateStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("request body must be valid JSON"))
		return
	}

	item, err := h.userBooks.UpdateStatus(middleware.CurrentUserID(c), id, input.Status)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{"item": item})
}

func (h *UserBookHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.userBooks.Delete(middleware.CurrentUserID(c), id); err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{"message": "book removed from your list"})
}

func (h *UserBookHandler) Stats(c *gin.Context) {
	stats, err := h.userBooks.Stats(middleware.CurrentUserID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, gin.H{"stats": stats})
}
