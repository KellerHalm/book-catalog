package service

import (
	"errors"
	"strings"

	"book-catalog/internal/apperror"
	"book-catalog/internal/models"
	"book-catalog/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

var errUserNotFound = errors.New("user not found")

type AuthService struct {
	users UserRepository
}

func NewAuthService(users UserRepository) *AuthService {
	return &AuthService{users: users}
}

type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (i RegisterInput) Validate() *apperror.Error {
	if !usernamePattern.MatchString(strings.TrimSpace(i.Username)) {
		return apperror.BadRequest("username must be 3-32 characters, letters, digits or underscore")
	}
	if !emailPattern.MatchString(strings.TrimSpace(i.Email)) {
		return apperror.BadRequest("email is invalid")
	}
	if len(i.Password) < 8 {
		return apperror.BadRequest("password must be at least 8 characters")
	}
	return nil
}

func (s *AuthService) Register(input RegisterInput) (string, *models.User, error) {
	if appErr := input.Validate(); appErr != nil {
		return "", nil, appErr
	}

	username := normalizeUsername(input.Username)
	email := normalizeEmail(input.Email)

	if _, err := s.users.FindByLogin(username); err == nil {
		return "", nil, apperror.Conflict("username is already taken")
	} else if !errors.Is(err, errUserNotFound) {
		return "", nil, apperror.Internal(err)
	}

	if _, err := s.users.FindByLogin(email); err == nil {
		return "", nil, apperror.Conflict("email is already registered")
	} else if !errors.Is(err, errUserNotFound) {
		return "", nil, apperror.Internal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, apperror.Internal(err)
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := s.users.Create(user); err != nil {
		return "", nil, apperror.Internal(err)
	}

	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return "", nil, apperror.Internal(err)
	}
	return token, user, nil
}

type LoginInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (s *AuthService) Login(input LoginInput) (string, *models.User, error) {
	login := strings.ToLower(strings.TrimSpace(input.Login))
	if login == "" || input.Password == "" {
		return "", nil, apperror.BadRequest("login and password are required")
	}

	user, err := s.users.FindByLogin(login)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return "", nil, apperror.Unauthorized("invalid credentials")
		}
		return "", nil, apperror.Internal(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return "", nil, apperror.Unauthorized("invalid credentials")
	}

	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return "", nil, apperror.Internal(err)
	}
	return token, user, nil
}

func (s *AuthService) Me(userID uint) (*models.User, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Internal(err)
	}
	return user, nil
}
