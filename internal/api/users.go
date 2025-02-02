package api

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/wittyjudge/blog-service-api/internal/auth"
	"github.com/wittyjudge/blog-service-api/internal/domains"
)

func (a *API) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := &domains.User{}

		// Read Body. If no body, error
		// Validate user
		// Check if user already exist
		// Save to database
		// Return user

		if err := a.fromJSON(r.Body, user); err != nil {
			a.errorResponse(w, fmt.Sprintf("failed to decode body: %v", err), http.StatusBadRequest)
			return
		}

		v := validator.New()
		if err := v.Struct(user); err != nil {
			errors := err.(validator.ValidationErrors)
			a.errorResponse(w, fmt.Sprintf("%v", errors), http.StatusBadRequest)
			return
		}

		// hash password
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
	}
}

func (a *API) GetUserByUsername() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (a *API) GetUserByEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
