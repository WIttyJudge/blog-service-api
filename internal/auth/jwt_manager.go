package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wittyjudge/blog-service-api/internal/config"
)

var ErrTokenIsInvalid = errors.New("token is invalid")

type JWTTokenType string

const (
	AccessTokenType  JWTTokenType = "access"
	RefreshTokenType JWTTokenType = "refresh"
)

type UserClaims struct {
	UserID    int          `json:"user_id"`
	TokenType JWTTokenType `json:"token_type"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	config config.JWT
}

func NewJWTManager(config config.JWT) *JWTManager {
	return &JWTManager{
		config: config,
	}
}

func (m *JWTManager) CreateToken(tokenType JWTTokenType, userID int) (string, *UserClaims, error) {
	timeNow := time.Now()
	expiresAt := timeNow.Add(m.config.AccessTokenTTL)
	if tokenType == RefreshTokenType {
		expiresAt = timeNow.Add(m.config.RefreshTokenTTL)
	}

	userClaims := &UserClaims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(timeNow),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	tokenStr, err := token.SignedString([]byte(m.config.SecretKey))
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign the token: %w", err)
	}

	return tokenStr, userClaims, nil
}

func (m *JWTManager) VerifyToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(m.config.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (m *JWTManager) TokenFromRequest(r *http.Request) (string, error) {
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
