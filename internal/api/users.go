package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/wittyjudge/blog-service-api/internal/domains"
)

type RegisterUserPayload struct {
	FirstName string `json:"first_name" validate:"required,max=50"`
	LastName  string `json:"last_name" validate:"required,max=50"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=3,max=100"`
}

type RegisterUserResp struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (a *API) registerUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := &RegisterUserPayload{}

		if err := a.fromJSON(r.Body, payload); err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to decode body: %v", err), http.StatusBadRequest)
			return
		}

		if err := a.validator.Struct(payload); err != nil {
			errors := a.validator.ValidationErrorsToSlice(err)
			a.errorResponse(w, errors, http.StatusBadRequest)
			return
		}

		if a.userService.CheckIfExistsByEmail(payload.Email) {
			a.errorResponse(w, fmt.Sprintf("user with %s email already exist", payload.Email), http.StatusBadRequest)
			return
		}

		user := &domains.User{
			FirstName: payload.FirstName,
			LastName:  payload.LastName,
			Email:     payload.Email,
			Password:  payload.Password,
		}
		if err := a.userService.Create(user); err != nil {
			a.errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := &RegisterUserResp{ID: user.ID, Email: user.Email}
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

		tokens, err := a.authService.CreateAccessRefreshToken(user.ID)
		if err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to create token: %v", err), http.StatusInternalServerError)
			return
		}

		a.successResponse(w, tokens, http.StatusOK)
	}
}

func (a *API) refreshAccessTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := JWTUserClaimsCtx(r.Context())

		accessToken, err := a.authService.CreateAccessToken(claims.UserID)
		if err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to create token: %v", err), http.StatusInternalServerError)
			return
		}

		a.successResponse(w, accessToken, http.StatusOK)
	}
}

func (a *API) logoutUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := JWTUserClaimsCtx(r.Context())
		tokenStr, err := a.authService.TokenFromRequest(r)
		if err != nil {
			a.errorResponse(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ttl := time.Until(claims.ExpiresAt.Time)

		if err := a.authService.BlockToken(tokenStr, ttl); err != nil {
			a.errorResponse(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
