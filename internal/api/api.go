package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/wittyjudge/blog-service-api/internal/auth"
	"github.com/wittyjudge/blog-service-api/internal/config"
	"github.com/wittyjudge/blog-service-api/internal/domains"
	"github.com/wittyjudge/blog-service-api/internal/repositories"
	"github.com/wittyjudge/blog-service-api/internal/services"
	"github.com/wittyjudge/blog-service-api/internal/validator"
	"go.uber.org/zap"
)

type API struct {
	ctx         context.Context
	config      *config.Config
	logger      *zap.Logger
	validator   *validator.Validator
	redisClient *redis.Client

	authService    *services.AuthService
	userService    domains.UserService
	articleService domains.ArticleService

	httpSrv *http.Server
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

func NewAPI(ctx context.Context, config *config.Config, logger *zap.Logger, pgPool *pgxpool.Pool, redisClient *redis.Client) *API {
	jwtManager := auth.NewJWTManager(config.API.JWT)

	userRepo := repositories.NewUserPostgres(ctx, pgPool)
	articleRepo := repositories.NewArticlePostgres(ctx, pgPool)

	authService := services.NewAuthService(ctx, jwtManager, redisClient)
	userService := services.NewUserService(ctx, logger, userRepo, redisClient)
	articleService := services.NewArticleService(ctx, logger, articleRepo, redisClient)

	api := &API{
		ctx:       ctx,
		config:    config,
		logger:    logger,
		validator: validator.New(),

		authService:    authService,
		userService:    userService,
		articleService: articleService,
	}

	api.httpSrv = &http.Server{
		Addr:    config.API.HostPort(),
		Handler: api.routers(),
	}

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

	r.Use(a.PopulateRequestID)
	r.Use(a.Logging)
	r.Use(a.Recovery)

	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthz", a.healthzHandler())

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", a.loginUserHandler())
			r.Post("/register", a.registerUserHandler())

			r.Group(func(r chi.Router) {
				r.Use(a.JWTRefreshToken)
				r.Post("/refresh", a.refreshAccessTokenHandler())
				r.Post("/logout", a.logoutUserHandler())
			})
		})

		r.Route("/articles", func(r chi.Router) {
			r.With(a.Pagination).Get("/", a.getArticlesHandler())
			r.With(a.JWTAccessToken).Post("/", a.createArticleHandler())

			r.Route("/{slug}", func(r chi.Router) {
				r.Get("/", a.getArticleHandler())
				r.Group(func(r chi.Router) {
					r.Use(a.JWTAccessToken)
					r.Put("/", a.updateArticleHandler())
					r.Delete("/", a.deleteArticleHandler())
				})
			})
		})
	})

	return r
}

func (a *API) successResponse(w http.ResponseWriter, data any, code int) {
	resp := SuccessResponse{Data: data}
	a.respond(w, resp, code)
}

func (a *API) errorResponse(w http.ResponseWriter, message any, code int) {
	if message, ok := message.(string); ok {
		resp := ErrorResponse{ErrorFormat{Code: code, Message: message}}
		a.respond(w, resp, code)
		return
	}

	messages := message.([]string)

	if len(messages) == 1 {
		resp := ErrorResponse{ErrorFormat{Code: code, Message: messages[0]}}
		a.respond(w, resp, code)
		return
	}

	resp := ErrorResponse{ErrorFormat{Code: code, Messages: messages}}
	a.respond(w, resp, code)
}

func (a *API) respond(w http.ResponseWriter, resp any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	a.toJSON(w, resp)
}

func (a *API) fromJSON(r io.Reader, s any) error {
	return json.NewDecoder(r).Decode(s)
}

func (a *API) toJSON(w io.Writer, s any) {
	_ = json.NewEncoder(w).Encode(s)
	return
}
