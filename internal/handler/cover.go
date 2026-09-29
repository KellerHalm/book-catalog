package handler

import (
	"book-catalog/internal/apperror"
	"book-catalog/internal/service"

	"github.com/gin-gonic/gin"
)

type CoverHandler struct {
	covers *service.BookCoverService
}

func NewCoverHandler(covers *service.BookCoverService) *CoverHandler {
	return &CoverHandler{covers: covers}
}

func (h *CoverHandler) Upload(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	header, err := c.FormFile("cover")
	if err != nil {
		c.Error(apperror.BadRequest("multipart form field 'cover' with an image is required"))
		return
	}

	book, err := h.covers.Upload(id, header)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{"book": book})
}

func (h *CoverHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.covers.Delete(id); err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{"message": "cover removed"})
}
