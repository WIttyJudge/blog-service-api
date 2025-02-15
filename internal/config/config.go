package config

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Environment string    `yaml:"environment" env:"ENVIRONMENT" env-default:"development"`
	Databases   Databases `yaml:"databases"`
	API         API       `yaml:"api"`
}

type Databases struct {
	Postgres Postgres `yaml:"postgres"`
	Redis    Redis    `yaml:"redis"`
}

type Postgres struct {
	Host         string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port         string `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	Password     string `yaml:"password" env:"POSTGRES_PASSWORD"`
	Username     string `yaml:"username" env:"POSTGRES_USERNAME"`
	Database     string `yaml:"database" env:"POSTGRES_DATABASE"`
	SSLMode      string `yaml:"ssl_mode" env:"POSTGRES_SSLMODE" env-default:"disable"`
	PoolMinConns int    `yaml:"pool_min_conns" env:"POSTGRES_POOL_MIN_CONNS" env-default:"1"`
	PoolMaxConns int    `yaml:"pool_max_conns" env:"POSTGRES_POOL_MAX_CONNS" env-default:"5"`
}

type Redis struct {
	Host         string `yaml:"host" env:"REDIS_HOST" env-default:"localhost"`
	Port         string `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
	Password     string `yaml:"password" env:"REDIS_PASSWORD"`
	Database     int    `yaml:"database" env:"REDIS_DATABASE"`
	PoolMaxConns int    `yaml:"pool_max_conns" env:"REDIS_POOL_MAX_CONNS" env-default:"5"`
}

type API struct {
	Host string `yaml:"host" env:"API_HOST"`
	Port string `yaml:"port" env:"API_PORT"`
	JWT  JWT    `yaml:"jwt"`
}

type JWT struct {
	SecretKey       string        `yaml:"secret_key" env:"API_JWT_SECRET_KEY"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" env:"API_JWT_ACCESS_TOKEN_TTL"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env:"API_JWT_REFRESH_TOKEN_TTL"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (a *API) HostPort() string {
	return net.JoinHostPort(a.Host, a.Port)
}

func (a *Redis) HostPort() string {
	return net.JoinHostPort(a.Host, a.Port)
}

func (p *Postgres) ConnectionDSN() string {
	return fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s pool_max_conns=%d pool_min_conns=%d",
		p.Username,
		p.Password,
		p.Host,
		p.Port,
		p.Database,
		p.SSLMode,
		p.PoolMaxConns,
		p.PoolMinConns,
	)
}

// ConnectionURL create for go-migration package to run migration.
// It only accepts URL format.
// https://github.com/golang-migrate/migrate/tree/master/database/postgres
func (p *Postgres) ConnectionURL() string {
	hostPort := net.JoinHostPort(p.Host, p.Port)

	u := &url.URL{
		Scheme: "postgres",
		Host:   hostPort,
		Path:   p.Database,
	}

	if p.Username != "" || p.Password != "" {
		u.User = url.UserPassword(p.Username, p.Password)
	}

	q := u.Query()

	if v := p.SSLMode; v != "" {
		q.Add("sslmode", v)
	}

	u.RawQuery = q.Encode()

	return u.String()
}
