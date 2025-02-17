package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/wittyjudge/blog-service-api/internal/auth"
	"github.com/wittyjudge/blog-service-api/pkg/cache"
)

type AuthService struct {
	ctx               context.Context
	jwtManager        *auth.JWTManager
	jwtBlocklistCache *cache.Cache[string, struct{}]
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

func NewAuthService(ctx context.Context, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		ctx:               ctx,
		jwtManager:        jwtManager,
		jwtBlocklistCache: cache.New[string, struct{}](10_000, "jwt_blocklist_cache"),
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
	s.jwtBlocklistCache.Set(token, struct{}{}, ttl)

	return nil
}

func (s *AuthService) IsBlocked(token string) (bool, error) {
	_, ok := s.jwtBlocklistCache.Get(token)
	return ok, nil
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
