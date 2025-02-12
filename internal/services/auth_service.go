package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wittyjudge/blog-service-api/internal/auth"
)

type AuthService struct {
	ctx        context.Context
	jwtManager *auth.JWTManager
	cache      *redis.Client
}

type Token struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type AccessToken struct {
	Token *Token `json:"access_token"`
}

type RefreshToken struct {
	Token *Token `json:"refresh_token"`
}

type AccessRefreshToken struct {
	AccessToken
	RefreshToken
}

func NewAuthService(ctx context.Context, jwtManager *auth.JWTManager, cache *redis.Client) *AuthService {
	return &AuthService{
		ctx:        ctx,
		jwtManager: jwtManager,
		cache:      cache,
	}
}

func (s *AuthService) CreateAccessRefreshToken(userID int) (*AccessRefreshToken, error) {
	access, err := s.CreateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refresh, err := s.CreateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &AccessRefreshToken{AccessToken: *access, RefreshToken: *refresh}, nil
}

func (s *AuthService) CreateAccessToken(userID int) (*AccessToken, error) {
	token, err := s.createToken(auth.AccessTokenType, userID)
	if err != nil {
		return nil, err
	}

	return &AccessToken{Token: token}, nil
}

func (s *AuthService) CreateRefreshToken(userID int) (*RefreshToken, error) {
	token, err := s.createToken(auth.RefreshTokenType, userID)
	if err != nil {
		return nil, err
	}

	return &RefreshToken{Token: token}, nil
}

func (s *AuthService) VerifyToken(token string) (*auth.UserClaims, error) {
	return s.jwtManager.VerifyToken(token)
}

func (s *AuthService) TokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "", fmt.Errorf("invalid token format")
	}

	return tokenString, nil
}

func (s *AuthService) BlockToken(token string, ttl time.Duration) error {
	key := fmt.Sprintf("jwt-blocklist:%s", token)
	_, err := s.cache.Set(s.ctx, key, true, ttl).Result()

	return err
}

func (s *AuthService) IsBlocked(token string) (bool, error) {
	key := fmt.Sprintf("jwt-blocklist:%s", token)
	exists, err := s.cache.Exists(s.ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}

func (s *AuthService) createToken(t auth.JWTTokenType, userID int) (*Token, error) {
	token, claims, err := s.jwtManager.CreateToken(t, userID)
	if err != nil {
		return nil, err
	}

	resp := &Token{
		Token:     token,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	return resp, nil
}
