package api

import (
	"context"
	"net/http"

	"github.com/wittyjudge/blog-service-api/internal/auth"
)

type (
	contextJWTUserClaims struct{}
	contextJWTToken      string
)

func (a *API) JWTAccessTokenMiddleware(next http.Handler) http.Handler {
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

func (a *API) JWTRefreshTokenMiddleware(next http.Handler) http.Handler {
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

func (a *API) getJWTClaimsFromAuth(r *http.Request) (*auth.UserClaims, error) {
	tokenStr, err := a.jwtManager.TokenFromRequest(r)
	if err != nil {
		return nil, err
	}

	exists, err := a.redisClient.Exists(a.ctx, tokenStr).Result()
	if err != nil {
		return nil, err
	}

	if exists == 1 {
		return nil, auth.ErrTokenIsInvalid
	}

	return a.jwtManager.VerifyToken(tokenStr)
}

func JWTUserClaimsCtx(ctx context.Context) *auth.UserClaims {
	return ctx.Value(contextJWTUserClaims{}).(*auth.UserClaims)
}

func withJWTUserClaims(ctx context.Context, userClaims *auth.UserClaims) context.Context {
	return context.WithValue(ctx, contextJWTUserClaims{}, userClaims)
}
