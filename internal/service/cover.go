package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"book-catalog/internal/apperror"
	"book-catalog/internal/models"
)

type CoverStore struct {
	uploadDir string
	maxBytes  int64
}

func NewCoverStore(uploadDir string, maxBytes int64) *CoverStore {
	return &CoverStore{uploadDir: uploadDir, maxBytes: maxBytes}
}

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

const sniffLen = 512

func (s *CoverStore) Save(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", apperror.BadRequest("image file is required")
	}
	if fileHeader.Size <= 0 {
		return "", apperror.BadRequest("uploaded file is empty")
	}
	if fileHeader.Size > s.maxBytes {
		return "", apperror.BadRequest(fmt.Sprintf("file is too large, maximum is %d MB", s.maxBytes/1024/1024))
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", apperror.BadRequest("uploaded file cannot be read")
	}
	defer file.Close()

	head := make([]byte, sniffLen)
	n, err := file.Read(head)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", apperror.BadRequest("uploaded file cannot be read")
	}
	head = head[:n]

	ext, ok := allowedImageTypes[http.DetectContentType(head)]
	if !ok {
		return "", apperror.BadRequest("file must be an image in jpeg, png, webp or gif format")
	}

	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return "", apperror.Internal(err)
	}

	name := newFileName(ext)
	fullPath := filepath.Join(s.uploadDir, name)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", apperror.Internal(err)
	}
	defer dst.Close()

	if _, err := dst.Write(head); err != nil {
		return "", apperror.Internal(err)
	}
	if _, err := io.Copy(dst, file); err != nil {
		return "", apperror.Internal(err)
	}

	return "/uploads/" + name, nil
}

func (s *CoverStore) Remove(url string) {
	if url == "" {
		return
	}
	name := filepath.Base(url)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return
	}
	_ = os.Remove(filepath.Join(s.uploadDir, name))
}

func newFileName(ext string) string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("cover-%d%s", os.Getpid(), ext)
	}
	return "cover-" + hex.EncodeToString(buf) + ext
}

type BookCoverService struct {
	books  *BookService
	covers *CoverStore
}

func NewBookCoverService(books *BookService, covers *CoverStore) *BookCoverService {
	return &BookCoverService{books: books, covers: covers}
}

func (s *BookCoverService) Upload(id uint, header *multipart.FileHeader) (*models.Book, error) {
	book, err := s.books.GetByID(id)
	if err != nil {
		return nil, err
	}

	url, err := s.covers.Save(header)
	if err != nil {
		return nil, err
	}

	previous := book.CoverURL
	book.CoverURL = url
	if err := s.books.save(book); err != nil {
		s.covers.Remove(url)
		return nil, err
	}
	s.covers.Remove(previous)

	return book, nil
}

func (s *BookCoverService) Delete(id uint) error {
	book, err := s.books.GetByID(id)
	if err != nil {
		return err
	}

	previous := book.CoverURL
	book.CoverURL = ""
	if err := s.books.save(book); err != nil {
		return err
	}
	s.covers.Remove(previous)

	return nil
}
