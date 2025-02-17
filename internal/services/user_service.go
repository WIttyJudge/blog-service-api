package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wittyjudge/blog-service-api/internal/auth"
	"github.com/wittyjudge/blog-service-api/internal/domains"
	"github.com/wittyjudge/blog-service-api/pkg/cache"
	"go.uber.org/zap"
)

var ErrInvalidEmailOrPassword = errors.New("invalid email or password")

type UserService struct {
	ctx              context.Context
	logger           *zap.Logger
	repo             domains.UserRepository
	userByEmailCache *cache.Cache[string, *domains.User]
}

func NewUserService(ctx context.Context, logger *zap.Logger, repo domains.UserRepository) *UserService {
	return &UserService{
		ctx:              ctx,
		logger:           logger,
		repo:             repo,
		userByEmailCache: cache.New[string, *domains.User](1000, "user_by_email_cache"),
	}
}

func (s *UserService) GetByEmail(email string) (*domains.User, error) {
	if cached, ok := s.userByEmailCache.Get(email); ok {
		return cached, nil
	}

	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	s.userByEmailCache.Set(email, user, 10*time.Minute)

	return user, nil
}

func (s *UserService) Create(user *domains.User) error {
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	if err := s.repo.Create(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *UserService) VerifyCredentials(email string, password string) error {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return ErrInvalidEmailOrPassword
	}

	if !auth.CheckPasswordHash(user.Password, password) {
		return ErrInvalidEmailOrPassword
	}

	return nil
}

func (s *UserService) CheckIfExistsByEmail(email string) bool {
	return s.repo.CheckIfExistsByEmail(email)
}
