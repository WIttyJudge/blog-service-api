package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wittyjudge/blog-service-api/internal/auth"
	"github.com/wittyjudge/blog-service-api/internal/config"
	"github.com/wittyjudge/blog-service-api/internal/domains"
	"github.com/wittyjudge/blog-service-api/internal/repositories"
	"github.com/wittyjudge/blog-service-api/internal/validator"
	"go.uber.org/zap"
)

type API struct {
	ctx       context.Context
	config    *config.Config
	logger    *zap.Logger
	validator *validator.Validator
	jwtMaker  *auth.JWTMaker

	userRepo    domains.UserRepository
	sessionRepo domains.SessionRepository
	httpSrv     *http.Server
}

type SuccessResponse struct {
	Data any `json:"data"`
}

type ErrorResponse struct {
	Error ErrorFormat `json:"error"`
}

type ErrorFormat struct {
	Code     int      `json:"code"`
	Message  string   `json:"message,omitempty"`
	Messages []string `json:"messages,omitempty"`
}

func NewAPI(ctx context.Context, config *config.Config, logger *zap.Logger, pgPool *pgxpool.Pool) *API {
	userRepo := repositories.NewPostgresUser(ctx, pgPool)
	sessionRepo := repositories.NewPostgresSession(ctx, pgPool)
	// userService := services.NewUserService(ctx, userRepo, logger, validator)

	api := &API{
		ctx:       ctx,
		config:    config,
		logger:    logger,
		validator: validator.New(),
		jwtMaker:  auth.NewJWTMaker(config.API.JWT),

		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}

	httpSrv := &http.Server{
		Addr:    config.API.HostPort(),
		Handler: api.routers(),
	}

	api.httpSrv = httpSrv
	return api
}

func (a *API) Start() error {
	a.logger.Info("Starting HTTP server", zap.String("addr", a.config.API.HostPort()))

	err := a.httpSrv.ListenAndServe()
	if err != http.ErrServerClosed && err != nil {
		return err
	}

	return nil
}

func (a *API) Stop() error {
	a.logger.Info("Stopping API")
	err := a.httpSrv.Shutdown(a.ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return nil
}

func (a *API) routers() *chi.Mux {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthcheck", a.healthCheckHandler())

		r.Get("/secure", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Super secure endpoint"))
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", a.LoginUser())
			r.Post("/register", a.CreateUser())
			r.Post("/refresh", a.RefreshAccessToken())
			r.Post("/logout", a.LogoutUser())
		})
	})

	return r
}

func (a *API) successResponse(w http.ResponseWriter, data any, code int) error {
	resp := SuccessResponse{Data: data}
	return a.respond(w, resp, code)
}

func (a *API) errorResponse(w http.ResponseWriter, message any, code int) error {
	if message, ok := message.(string); ok {
		resp := ErrorResponse{ErrorFormat{Code: code, Message: message}}
		return a.respond(w, resp, code)
	}

	messages := message.([]string)

	if len(messages) == 1 {
		resp := ErrorResponse{ErrorFormat{Code: code, Message: messages[0]}}
		return a.respond(w, resp, code)
	}

	resp := ErrorResponse{ErrorFormat{Code: code, Messages: messages}}
	return a.respond(w, resp, code)
}

func (a *API) respond(w http.ResponseWriter, resp any, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	return a.toJSON(w, resp)
}

func (a *API) fromJSON(r io.Reader, s any) error {
	return json.NewDecoder(r).Decode(s)
}

func (a *API) toJSON(w io.Writer, s any) error {
	return json.NewEncoder(w).Encode(s)
}
