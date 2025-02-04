package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wittyjudge/blog-service-api/internal/config"
)

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

type JWTMaker struct {
	config config.JWT
}

func NewJWTMaker(config config.JWT) *JWTMaker {
	return &JWTMaker{
		config: config,
	}
}

func (m *JWTMaker) CreateToken(tokenType JWTTokenType, userID int) (string, *UserClaims, error) {
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

func (m *JWTMaker) VerifyToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(m.config.SecretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse the token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
