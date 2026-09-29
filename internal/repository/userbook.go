package repository

import (
	"errors"
	"strings"

	"book-catalog/internal/models"

	"gorm.io/gorm"
)

type UserBookFilter struct {
	Status  string
	Page    int
	PerPage int
}

type UserBookRepository struct {
	db *gorm.DB
}

func NewUserBookRepository(db *gorm.DB) *UserBookRepository {
	return &UserBookRepository{db: db}
}

func (r *UserBookRepository) buildQuery(userID uint, filter UserBookFilter) *gorm.DB {
	q := r.db.Model(&models.UserBook{}).Where("user_id = ?", userID)

	if s := strings.TrimSpace(filter.Status); s != "" && s != "all" {
		q = q.Where("status = ?", s)
	}

	return q
}

func (r *UserBookRepository) List(userID uint, filter UserBookFilter) ([]models.UserBook, int64, error) {
	var total int64
	if err := r.buildQuery(userID, filter).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 12
	}

	var items []models.UserBook
	err := r.buildQuery(userID, filter).
		Preload("Book").
		Order("user_books.updated_at DESC, user_books.id DESC").
		Limit(filter.PerPage).
		Offset((filter.Page - 1) * filter.PerPage).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *UserBookRepository) FindByID(userID, id uint) (*models.UserBook, error) {
	var item models.UserBook
	err := r.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Book").
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *UserBookRepository) Create(item *models.UserBook) error {
	return r.db.Create(item).Error
}

func (r *UserBookRepository) Update(item *models.UserBook) error {
	return r.db.Save(item).Error
}

func (r *UserBookRepository) Delete(userID, id uint) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.UserBook{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserBookRepository) Stats(userID uint) (models.Stats, error) {
	stats := models.Stats{}

	rows, err := r.db.Model(&models.UserBook{}).
		Select("status, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("status").
		Rows()
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return stats, err
		}
		switch status {
		case models.StatusWantToRead:
			stats.WantToRead = int(count)
		case models.StatusReading:
			stats.Reading = int(count)
		case models.StatusFinished:
			stats.Finished = int(count)
		}
		stats.Total += int(count)
	}

	return stats, rows.Err()
}
