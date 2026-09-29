package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"book-catalog/internal/config"
	"book-catalog/internal/handler"
	"book-catalog/internal/middleware"
	"book-catalog/internal/repository"
	"book-catalog/internal/service"
	"book-catalog/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(cfg *config.Config, db *gorm.DB) *gin.Engine {
	utils.Init(cfg.JWTSecret, cfg.TokenTTL())

	bookRepo := repository.NewBookRepository(db)
	userBookRepo := repository.NewUserBookRepository(db)
	userRepo := service.NewUserRepository(db)

	bookService := service.NewBookService(bookRepo)
	userBookService := service.NewUserBookService(userBookRepo, bookRepo)
	authService := service.NewAuthService(userRepo)
	coverService := service.NewBookCoverService(bookService, service.NewCoverStore(cfg.UploadDir(), cfg.MaxUploadSize()))

	bookHandler := handler.NewBookHandler(bookService)
	userBookHandler := handler.NewUserBookHandler(userBookService)
	authHandler := handler.NewAuthHandler(authService)
	coverHandler := handler.NewCoverHandler(coverService)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS(cfg.Origins), middleware.ErrorHandler())

	r.MaxMultipartMemory = cfg.MaxUploadSize()
	r.Static("/uploads", cfg.UploadDir())
	r.Static("/static", cfg.StaticDir())
	r.GET("/", serveIndex(cfg.StaticDir()))

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "route not found"}})
			return
		}
		c.File(indexPath(cfg.StaticDir()))
	})

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			code := http.StatusOK
			status := "ok"
			if sqlDB, err := db.DB(); err != nil || sqlDB.Ping() != nil {
				code, status = http.StatusServiceUnavailable, "degraded"
			}
			c.JSON(code, gin.H{"status": status})
		})

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthRequired(), authHandler.Me)
		}

		api.GET("/books", bookHandler.List)
		api.GET("/books/genres", bookHandler.Genres)
		api.GET("/books/:id", bookHandler.GetByID)

		private := api.Group("")
		private.Use(middleware.AuthRequired())
		{
			private.POST("/books", bookHandler.Create)
			private.PUT("/books/:id", bookHandler.Update)
			private.DELETE("/books/:id", bookHandler.Delete)
			private.DELETE("/books/:id/cover", coverHandler.Delete)

			private.GET("/my/books", userBookHandler.List)
			private.GET("/my/stats", userBookHandler.Stats)
			private.POST("/my/books/:bookId", userBookHandler.Add)
			private.PATCH("/my/books/:id", userBookHandler.UpdateStatus)
			private.DELETE("/my/books/:id", userBookHandler.Delete)
		}

		upload := api.Group("")
		upload.Use(middleware.MaxBodySize(cfg.MaxUploadSize()+1024*1024), middleware.AuthRequired())
		{
			upload.POST("/books/:id/cover", coverHandler.Upload)
		}
	}

	return r
}

func indexPath(dir string) string {
	return filepath.Join(dir, "index.html")
}

func serveIndex(dir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := os.Stat(indexPath(dir)); err != nil {
			c.JSON(http.StatusOK, gin.H{"service": "book-catalog", "status": "ok"})
			return
		}
		c.File(indexPath(dir))
	}
}
