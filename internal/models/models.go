package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`

	Books []UserBook `gorm:"foreignKey:UserID" json:"-"`
}

type Book struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:255;not null;index" json:"title"`
	Author      string    `gorm:"size:255;not null;index" json:"author"`
	Genre       string    `gorm:"size:64;not null;index" json:"genre"`
	Year        int       `gorm:"not null;default:0" json:"year"`
	Pages       int       `gorm:"not null;default:0" json:"pages"`
	Rating      float64   `gorm:"not null;default:0" json:"rating"`
	Description string    `gorm:"type:text" json:"description"`
	CoverURL    string    `gorm:"size:255;not null;default:''" json:"cover_url"`
	SearchIndex string    `gorm:"column:search_index;size:1024;not null;default:'';index" json:"-"`
	GenreIndex  string    `gorm:"column:genre_index;size:64;not null;default:'';index" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}

func (b *Book) BeforeSave(*gorm.DB) error {
	b.SearchIndex = strings.ToLower(strings.Join([]string{b.Title, b.Author, b.Genre}, " "))
	b.GenreIndex = strings.ToLower(b.Genre)
	return nil
}

const (
	StatusWantToRead = "want_to_read"
	StatusReading    = "reading"
	StatusFinished   = "finished"
)

var AllStatuses = []string{StatusWantToRead, StatusReading, StatusFinished}

func IsValidStatus(status string) bool {
	for _, s := range AllStatuses {
		if s == status {
			return true
		}
	}
	return false
}

type UserBook struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_book;index" json:"user_id"`
	BookID    uint      `gorm:"not null;uniqueIndex:idx_user_book;index" json:"book_id"`
	Status    string    `gorm:"size:32;not null;default:'want_to_read';index" json:"status"`
	AddedAt   time.Time `gorm:"autoCreateTime" json:"added_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Book Book `gorm:"foreignKey:BookID" json:"book"`
}

type PageMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type Stats struct {
	WantToRead int `json:"want_to_read"`
	Reading    int `json:"reading"`
	Finished   int `json:"finished"`
	Total      int `json:"total"`
}
