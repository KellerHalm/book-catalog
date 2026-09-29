package repository

import (
	"errors"
	"strings"

	"book-catalog/internal/models"

	"gorm.io/gorm"
)

type BookFilter struct {
	Search  string
	Genre   string
	Sort    string
	Page    int
	PerPage int
}

type BookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) *BookRepository {
	return &BookRepository{db: db}
}

func (r *BookRepository) buildQuery(filter BookFilter) *gorm.DB {
	q := r.db.Model(&models.Book{})

	if s := strings.TrimSpace(filter.Search); s != "" {
		q = q.Where("search_index LIKE ?", "%"+strings.ToLower(s)+"%")
	}

	if g := strings.TrimSpace(filter.Genre); g != "" && g != "all" {
		q = q.Where("genre_index = ?", strings.ToLower(g))
	}

	switch filter.Sort {
	case "title":
		q = q.Order("title ASC")
	case "author":
		q = q.Order("author ASC")
	case "year":
		q = q.Order("year DESC")
	case "rating":
		q = q.Order("rating DESC")
	default:
		q = q.Order("id ASC")
	}

	return q
}

func (r *BookRepository) List(filter BookFilter) ([]models.Book, int64, error) {
	var total int64
	if err := r.buildQuery(filter).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 12
	}

	var books []models.Book
	err := r.buildQuery(filter).
		Limit(filter.PerPage).
		Offset((filter.Page - 1) * filter.PerPage).
		Find(&books).Error
	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (r *BookRepository) FindByID(id uint) (*models.Book, error) {
	var book models.Book
	if err := r.db.First(&book, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &book, nil
}

func (r *BookRepository) Genres() ([]string, error) {
	var genres []string
	err := r.db.Model(&models.Book{}).
		Distinct().
		Order("genre ASC").
		Where("genre <> ''").
		Pluck("genre", &genres).Error
	if err != nil {
		return nil, err
	}
	return genres, nil
}

func (r *BookRepository) Create(book *models.Book) error {
	return r.db.Create(book).Error
}

func (r *BookRepository) Update(book *models.Book) error {
	return r.db.Save(book).Error
}

func (r *BookRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Book{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *BookRepository) IsInUserList(userID, bookID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserBook{}).
		Where("user_id = ? AND book_id = ?", userID, bookID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
