package service

import (
	"errors"
	"math"
	"regexp"
	"strings"

	"book-catalog/internal/apperror"
	"book-catalog/internal/models"
	"book-catalog/internal/repository"
)

type BookService struct {
	books *repository.BookRepository
}

func NewBookService(books *repository.BookRepository) *BookService {
	return &BookService{books: books}
}

var (
	usernamePattern = regexp.MustCompile(`^[\p{L}\p{N}_]{3,32}$`)
	emailPattern    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[a-zA-Z]{2,}$`)
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func (s *BookService) List(filter repository.BookFilter) ([]models.Book, models.PageMeta, error) {
	books, total, err := s.books.List(filter)
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

	return books, models.PageMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: totalPages,
	}, nil
}

func (s *BookService) GetByID(id uint) (*models.Book, error) {
	book, err := s.books.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.NotFound("book not found")
		}
		return nil, apperror.Internal(err)
	}
	return book, nil
}

func (s *BookService) Genres() ([]string, error) {
	genres, err := s.books.Genres()
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if genres == nil {
		genres = []string{}
	}
	return genres, nil
}

func (s *BookService) Create(input BookInput) (*models.Book, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	book := input.toModel()
	if err := s.books.Create(&book); err != nil {
		return nil, apperror.Internal(err)
	}
	return &book, nil
}

func (s *BookService) Update(id uint, input BookInput) (*models.Book, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	book, err := s.books.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.NotFound("book not found")
		}
		return nil, apperror.Internal(err)
	}

	input.applyTo(book)
	if err := s.books.Update(book); err != nil {
		return nil, apperror.Internal(err)
	}
	return book, nil
}

func (s *BookService) Delete(id uint) error {
	if err := s.books.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperror.NotFound("book not found")
		}
		return apperror.Internal(err)
	}
	return nil
}

func (s *BookService) save(book *models.Book) error {
	if err := s.books.Update(book); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

type BookInput struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Genre       string  `json:"genre"`
	Year        int     `json:"year"`
	Pages       int     `json:"pages"`
	Rating      float64 `json:"rating"`
	Description string  `json:"description"`
}

func (i BookInput) Validate() *apperror.Error {
	if strings.TrimSpace(i.Title) == "" {
		return apperror.BadRequest("title is required")
	}
	if strings.TrimSpace(i.Author) == "" {
		return apperror.BadRequest("author is required")
	}
	if strings.TrimSpace(i.Genre) == "" {
		return apperror.BadRequest("genre is required")
	}
	if i.Year < 0 || i.Year > 2100 {
		return apperror.BadRequest("year must be between 0 and 2100")
	}
	if i.Pages < 0 || i.Pages > 100000 {
		return apperror.BadRequest("pages must be between 0 and 100000")
	}
	if i.Rating < 0 || i.Rating > 5 {
		return apperror.BadRequest("rating must be between 0 and 5")
	}
	return nil
}

func (i BookInput) toModel() models.Book {
	book := models.Book{}
	i.applyTo(&book)
	return book
}

func (i BookInput) applyTo(book *models.Book) {
	book.Title = strings.TrimSpace(i.Title)
	book.Author = strings.TrimSpace(i.Author)
	book.Genre = strings.TrimSpace(i.Genre)
	book.Year = i.Year
	book.Pages = i.Pages
	book.Rating = i.Rating
	book.Description = strings.TrimSpace(i.Description)
}
