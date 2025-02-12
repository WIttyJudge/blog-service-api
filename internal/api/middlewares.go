package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/wittyjudge/blog-service-api/internal/auth"
	"go.uber.org/zap"
)

type CustomKey string

type (
	contextJWTUserClaims     struct{}
	contextPaginationOptions struct{}
)

type PaginationOptions struct {
	Cursor   int `validate:"min=0"`
	PageSize int `validate:"min=5,max=100"`
}

func NewPaginationOptions() *PaginationOptions {
	return &PaginationOptions{
		Cursor:   0,
		PageSize: 5,
	}
}

func (a *API) JWTAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userClaims, err := a.getJWTClaimsFromAuth(r)
		if err != nil {
			a.errorResponse(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if userClaims.TokenType != auth.AccessTokenType {
			a.errorResponse(w, "access token must be provided", http.StatusUnauthorized)
			return
		}

		ctx := withJWTUserClaims(r.Context(), userClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *API) JWTRefreshToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userClaims, err := a.getJWTClaimsFromAuth(r)
		if err != nil {
			a.errorResponse(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if userClaims.TokenType != auth.RefreshTokenType {
			a.errorResponse(w, "refresh token must be provided", http.StatusUnauthorized)
			return
		}

		ctx := withJWTUserClaims(r.Context(), userClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *API) Pagination(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryCursor := r.URL.Query().Get("cursor")
		queryPageSize := r.URL.Query().Get("pageSize")

		paginationOptions := NewPaginationOptions()

		if parsedCursor, err := strconv.Atoi(queryCursor); err == nil {
			paginationOptions.Cursor = parsedCursor
		}

		if parsedPageSize, err := strconv.Atoi(queryPageSize); err == nil {
			paginationOptions.PageSize = parsedPageSize
		}

		if err := a.validator.Struct(paginationOptions); err != nil {
			errors := a.validator.ValidationErrorsToSlice(err)
			a.errorResponse(w, errors, http.StatusBadRequest)
			return
		}

		ctx := withContextPaginationOptions(r.Context(), paginationOptions)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Recovery recovers from panics and other fatal errors. It keeps the server and
// service running, returning 500 to the caller while also logging the error in
// a structured format.
func (a *API) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				a.logger.Error("http handler panic", zap.Any("addr", p))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (a *API) getJWTClaimsFromAuth(r *http.Request) (*auth.UserClaims, error) {
	tokenStr, err := a.authService.TokenFromRequest(r)
	if err != nil {
		return nil, err
	}

	blocked, err := a.authService.IsBlocked(tokenStr)
	if err != nil {
		return nil, err
	}

	if blocked {
		return nil, auth.ErrTokenIsInvalid
	}

	return a.authService.VerifyToken(tokenStr)
}

func JWTUserClaimsCtx(ctx context.Context) *auth.UserClaims {
	return ctx.Value(contextJWTUserClaims{}).(*auth.UserClaims)
}

func withJWTUserClaims(ctx context.Context, userClaims *auth.UserClaims) context.Context {
	return context.WithValue(ctx, contextJWTUserClaims{}, userClaims)
}

func contextPaginationOptionsCtx(ctx context.Context) *PaginationOptions {
	return ctx.Value(contextPaginationOptions{}).(*PaginationOptions)
}

func withContextPaginationOptions(ctx context.Context, pagOpts *PaginationOptions) context.Context {
	return context.WithValue(ctx, contextPaginationOptions{}, pagOpts)
}
