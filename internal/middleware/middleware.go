package middleware

import (
	"net/http"
	"strings"

	"book-catalog/internal/apperror"
	"book-catalog/internal/models"
	"book-catalog/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserKey = "currentUser"
	ContextClaims  = "currentClaims"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"message": "authorization header is required"},
			})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"message": "authorization header must use Bearer scheme"},
			})
			return
		}

		claims, err := utils.ParseToken(strings.TrimSpace(parts[1]))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"message": "invalid or expired token"},
			})
			return
		}

		c.Set(ContextClaims, claims)
		c.Set(ContextUserKey, &models.User{ID: claims.UserID, Username: claims.Username})
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) uint {
	if claims, ok := c.Get(ContextClaims); ok {
		if typed, ok := claims.(*utils.Claims); ok {
			return typed.UserID
		}
	}
	return 0
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		appErr := apperror.From(c.Errors.Last().Err)
		c.AbortWithStatusJSON(appErr.Code, gin.H{
			"error": gin.H{"message": appErr.Message},
		})
	}
}

func MaxBodySize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": gin.H{"message": "request body is too large"},
			})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[strings.TrimSpace(origin)] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Max-Age", "600")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
