package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"cloudpulse/internal/models"
	"cloudpulse/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

// Used when the email is unknown, so a wrong email takes as long as a wrong
// password. That stops attackers from guessing which emails have accounts.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("not-a-real-password"), bcrypt.DefaultCost)

type AuthService struct {
	users  *repository.UserRepo
	secret []byte
	ttl    time.Duration
}

func NewAuthService(users *repository.UserRepo, secret string, ttl time.Duration) *AuthService {
	return &AuthService{users: users, secret: []byte(secret), ttl: ttl}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (string, *models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}
	user, err := s.users.Create(ctx, name, strings.ToLower(email), string(hash))
	if err != nil {
		return "", nil, err
	}
	token, err := s.issueToken(user.ID)
	return token, user, err
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *models.User, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return "", nil, ErrInvalidCredentials
	}
	if err != nil {
		return "", nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", nil, ErrInvalidCredentials
	}
	token, err := s.issueToken(user.ID)
	return token, user, err
}

func (s *AuthService) issueToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

// ParseToken checks the signature and expiry, and returns the user id inside.
func (s *AuthService) ParseToken(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) { return s.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil || !token.Valid || claims.Subject == "" {
		return "", fmt.Errorf("invalid token")
	}
	return claims.Subject, nil
}