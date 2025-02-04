package services

import (
	"context"
	"fmt"

	"github.com/wittyjudge/blog-service-api/internal/auth"
	"github.com/wittyjudge/blog-service-api/internal/domains"
	"github.com/wittyjudge/blog-service-api/internal/validator"
	"go.uber.org/zap"
)

type UserService struct {
	ctx       context.Context
	logger    *zap.Logger
	validator *validator.Validator

	repo domains.UserRepository
}

func NewUserService(ctx context.Context, repo domains.UserRepository, logger *zap.Logger, validator *validator.Validator) *UserService {
	u := &UserService{
		ctx:       ctx,
		repo:      repo,
		logger:    logger,
		validator: validator,
	}

	return u
}

func (s *UserService) Create(user *domains.User) error {
	if s.repo.CheckIfExistsByEmail(user.Email) {
		return fmt.Errorf("user with %s email already exist", user.Email)
	}

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
