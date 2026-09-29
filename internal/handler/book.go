package handler

import (
	"strconv"

	"book-catalog/internal/apperror"
	"book-catalog/internal/repository"
	"book-catalog/internal/service"

	"github.com/gin-gonic/gin"
)

type BookHandler struct {
	books *service.BookService
}

func NewBookHandler(books *service.BookService) *BookHandler {
	return &BookHandler{books: books}
}

func (h *BookHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "12"))

	filter := repository.BookFilter{
		Search:  c.Query("search"),
		Genre:   c.Query("genre"),
		Sort:    c.DefaultQuery("sort", ""),
		Page:    page,
		PerPage: perPage,
	}

	books, meta, err := h.books.List(filter)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{
		"items": books,
		"meta":  meta,
	})
}

func (h *BookHandler) GetByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	book, err := h.books.GetByID(id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{"book": book})
}

func (h *BookHandler) Genres(c *gin.Context) {
	genres, err := h.books.Genres()
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, gin.H{"genres": genres})
}

func (h *BookHandler) Create(c *gin.Context) {
	var input service.BookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("request body must be valid JSON"))
		return
	}

	book, err := h.books.Create(input)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(201, gin.H{"book": book})
}

func (h *BookHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input service.BookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("request body must be valid JSON"))
		return
	}

	book, err := h.books.Update(id, input)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{"book": book})
}

func (h *BookHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.books.Delete(id); err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{"message": "book deleted"})
}

func parseID(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		c.Error(apperror.BadRequest("id must be a positive number"))
		return 0, false
	}
	return uint(value), true
}
