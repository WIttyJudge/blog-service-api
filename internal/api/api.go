package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wittyjudge/blog-service-api/internal/config"
	"github.com/wittyjudge/blog-service-api/internal/domains"
	"github.com/wittyjudge/blog-service-api/internal/repositories"
	"go.uber.org/zap"
)

type API struct {
	ctx    context.Context
	config *config.Config
	logger *zap.Logger

	userRepo domains.UserRepository
	httpSrv  *http.Server
}

func NewAPI(ctx context.Context, config *config.Config, logger *zap.Logger, pgPool *pgxpool.Pool) *API {
	userRepo := repositories.NewPostgresUser(ctx, pgPool)

	api := &API{
		ctx:    ctx,
		config: config,
		logger: logger,

		userRepo: userRepo,
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

	r.Get("/api/v1/healthcheck", a.healthCheckHandler())

	r.Post("/api/v1/register", a.CreateUser())
	r.Get("/api/v1/users/{username}", a.GetUserByUsername())

	return r
}

func (a *API) errorResponse(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	http.Error(w, err, code)
}

func (a *API) fromJSON(r io.Reader, s any) error {
	return json.NewDecoder(r).Decode(s)
}

func (a *API) toJSON(w io.Writer, s any) error {
	return json.NewEncoder(w).Encode(s)
}
