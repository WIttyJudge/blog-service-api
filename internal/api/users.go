package api

import (
	"fmt"
	"net/http"

	"github.com/wittyjudge/blog-service-api/internal/auth"
	"github.com/wittyjudge/blog-service-api/internal/domains"
)

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (a *API) CreateUser() http.HandlerFunc {
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

		if a.userRepo.CheckIfExistsByEmail(user.Email) {
			a.errorResponse(w, fmt.Sprintf("user with %s email already exist", user.Email), http.StatusBadRequest)
			return
		}

		hashedPassword, err := auth.HashPassword(user.Password)
		if err != nil {
			a.errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user.Password = hashedPassword

		if err := a.userRepo.Create(user); err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to create user: %v", err), http.StatusInternalServerError)
			return
		}

		resp := map[string]any{
			"id":    user.ID,
			"email": user.Email,
		}
		a.successResponse(w, resp, http.StatusCreated)
	}
}

func (a *API) LoginUser() http.HandlerFunc {
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

		user, err := a.userRepo.GetByEmail(payload.Email)
		if err != nil {
			a.errorResponse(w, "invalid email or password", http.StatusBadRequest)
			return
		}

		if !auth.CheckPasswordHash(user.Password, payload.Password) {
			a.errorResponse(w, "invalid email or password", http.StatusBadRequest)
			return
		}

		accessToken, accessClaims, err := a.jwtMaker.CreateToken(auth.AccessTokenType, user.ID)
		if err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to create token: %v", err), http.StatusInternalServerError)
			return
		}

		refreshToken, refreshClaims, err := a.jwtMaker.CreateToken(auth.RefreshTokenType, user.ID)
		if err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to create token: %v", err), http.StatusInternalServerError)
			return
		}

		session := &domains.Session{
			UserID:       user.ID,
			RefreshToken: refreshToken,
			ExpiresAt:    refreshClaims.ExpiresAt.Time,
		}
		if err := a.sessionRepo.CreateOrUpdate(session); err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to create token in db: %v", err), http.StatusInternalServerError)
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

func (a *API) RefreshAccessToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

func (a *API) LogoutUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
