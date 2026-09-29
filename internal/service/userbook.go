package service

import (
	"errors"
	"math"
	"strings"

	"book-catalog/internal/apperror"
	"book-catalog/internal/models"
	"book-catalog/internal/repository"
)

type UserBookService struct {
	userBooks *repository.UserBookRepository
	books     *repository.BookRepository
}

func NewUserBookService(userBooks *repository.UserBookRepository, books *repository.BookRepository) *UserBookService {
	return &UserBookService{userBooks: userBooks, books: books}
}

func (s *UserBookService) List(userID uint, filter repository.UserBookFilter) ([]models.UserBook, models.PageMeta, error) {
	items, total, err := s.userBooks.List(userID, filter)
	if err != nil {
		return nil, models.PageMeta{}, apperror.Internal(err)
	}

	perPage := filter.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 12
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(perPage)))
	}

	return items, models.PageMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: totalPages,
	}, nil
}

func (s *UserBookService) Add(userID uint, bookID uint) (*models.UserBook, error) {
	if _, err := s.books.FindByID(bookID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.NotFound("book not found")
		}
		return nil, apperror.Internal(err)
	}

	exists, err := s.books.IsInUserList(userID, bookID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if exists {
		return nil, apperror.Conflict("book is already in your list")
	}

	item := models.UserBook{
		UserID: userID,
		BookID: bookID,
		Status: models.StatusWantToRead,
	}
	if err := s.userBooks.Create(&item); err != nil {
		return nil, apperror.Internal(err)
	}

	return s.find(userID, item.ID)
}

func (s *UserBookService) UpdateStatus(userID, id uint, status string) (*models.UserBook, error) {
	status = strings.TrimSpace(status)
	if !models.IsValidStatus(status) {
		return nil, apperror.BadRequest("status must be one of: want_to_read, reading, finished")
	}

	item, err := s.find(userID, id)
	if err != nil {
		return nil, err
	}

	item.Status = status
	if err := s.userBooks.Update(item); err != nil {
		return nil, apperror.Internal(err)
	}

	return s.find(userID, id)
}

func (s *UserBookService) Delete(userID, id uint) error {
	if err := s.userBooks.Delete(userID, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperror.NotFound("entry not found in your list")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (s *UserBookService) Stats(userID uint) (models.Stats, error) {
	stats, err := s.userBooks.Stats(userID)
	if err != nil {
		return models.Stats{}, apperror.Internal(err)
	}
	return stats, nil
}

func (s *UserBookService) find(userID, id uint) (*models.UserBook, error) {
	item, err := s.userBooks.FindByID(userID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.NotFound("entry not found in your list")
		}
		return nil, apperror.Internal(err)
	}
	return item, nil
}
