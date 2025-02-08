package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/wittyjudge/blog-service-api/internal/auth"
	"github.com/wittyjudge/blog-service-api/internal/domains"
)

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (a *API) registerUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := &domains.User{}

		if err := a.fromJSON(r.Body, user); err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to decode body: %v", err), http.StatusBadRequest)
			return
		}

		if err := a.validator.Struct(user); err != nil {
			errors := a.validator.ValidationErrorsToSlice(err)
			a.errorResponse(w, errors, http.StatusBadRequest)
			return
		}

		if a.userService.CheckIfExistsByEmail(user.Email) {
			a.errorResponse(w, fmt.Sprintf("user with %s email already exist", user.Email), http.StatusBadRequest)
			return
		}

		if err := a.userService.Create(user); err != nil {
			a.errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := map[string]any{
			"id":    user.ID,
			"email": user.Email,
		}
		a.successResponse(w, resp, http.StatusCreated)
	}
}

func (a *API) loginUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := &LoginUserPayload{}

		if err := a.fromJSON(r.Body, payload); err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to decode body: %v", err), http.StatusBadRequest)
			return
		}

		if err := a.validator.Struct(payload); err != nil {
			errors := a.validator.ValidationErrorsToSlice(err)
			a.errorResponse(w, errors, http.StatusBadRequest)
			return
		}

		if err := a.userService.VerifyCredentials(payload.Email, payload.Password); err != nil {
			a.errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, err := a.userService.GetByEmail(payload.Email)
		if err != nil {
			a.errorResponse(w, "user not found", http.StatusNotFound)
			return
		}

		accessToken, accessClaims, err := a.jwtManager.CreateToken(auth.AccessTokenType, user.ID)
		if err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to create token: %v", err), http.StatusInternalServerError)
			return
		}

		refreshToken, refreshClaims, err := a.jwtManager.CreateToken(auth.RefreshTokenType, user.ID)
		if err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to create token: %v", err), http.StatusInternalServerError)
			return
		}

		resp := map[string]string{
			"access_token":             accessToken,
			"access_token_expires_at":  accessClaims.ExpiresAt.Time.String(),
			"refresh_token":            refreshToken,
			"refresh_token_expires_at": refreshClaims.ExpiresAt.Time.String(),
		}

		a.successResponse(w, resp, http.StatusOK)
	}
}

func (a *API) refreshAccessTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := JWTUserClaimsCtx(r.Context())

		accessToken, accessClaims, err := a.jwtManager.CreateToken(auth.AccessTokenType, claims.UserID)
		if err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to create token: %v", err), http.StatusInternalServerError)
			return
		}

		resp := map[string]string{
			"access_token":            accessToken,
			"access_token_expires_at": accessClaims.ExpiresAt.Time.String(),
		}
		a.successResponse(w, resp, http.StatusOK)
	}
}

func (a *API) logoutUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := JWTUserClaimsCtx(r.Context())
		tokenStr, err := a.jwtManager.TokenFromRequest(r)
		if err != nil {
			a.errorResponse(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ttl := time.Until(claims.ExpiresAt.Time)

		key := fmt.Sprintf("jwt-blocklist:%s", tokenStr)
		_, err = a.redisClient.Set(a.ctx, key, true, ttl).Result()
		if err != nil {
			a.errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
