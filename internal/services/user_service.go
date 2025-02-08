package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wittyjudge/blog-service-api/internal/auth"
	"github.com/wittyjudge/blog-service-api/internal/domains"
	"go.uber.org/zap"
)

var ErrInvalidEmailOrPassword = errors.New("invalid email or password")

type UserService struct {
	ctx    context.Context
	logger *zap.Logger

	repo        domains.UserRepository
	redisClient *redis.Client
}

func NewUserService(ctx context.Context, logger *zap.Logger, repo domains.UserRepository, redisClient *redis.Client) *UserService {
	u := &UserService{
		ctx:    ctx,
		logger: logger,

		repo:        repo,
		redisClient: redisClient,
	}

	return u
}

func (s *UserService) GetByEmail(email string) (*domains.User, error) {
	key := fmt.Sprintf("user:%s", email)
	cachedUser, err := s.redisClient.Get(s.ctx, key).Result()
	if err == nil {
		user := &domains.User{}
		if err := json.Unmarshal([]byte(cachedUser), user); err != nil {
			return nil, err
		}

		return user, nil
	}

	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	err = s.redisClient.Set(s.ctx, key, userJson, 15*time.Minute).Err()
	if err != nil {
		return nil, err
	}

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
