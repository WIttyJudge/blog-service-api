package config

import (
	"fmt"
	"net"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Environment string    `yaml:"environment" env:"ENVIRONMENT" env-default:"development"`
	Databases   Databases `yaml:"databases"`
	API         API       `yaml:"api"`
}

type Databases struct {
	PostgreSQL PostgreSQL `yaml:"postgresql"`
	Redis      Redis      `yaml:"redis"`
}

type PostgreSQL struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	Password     string `yaml:"password"`
	Username     string `yaml:"username"`
	Database     string `yaml:"database"`
	SSLMode      string `yaml:"ssl_mode" env-default:"disable"`
	PoolMaxConns int    `yaml:"pool_max_conns" env-default:"5"`
}

type Redis struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	Password     string `yaml:"password"`
	Database     int    `yaml:"database"`
	PoolMaxConns int    `yaml:"pool_max_conns" env-default:"5"`
}

type API struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
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

func (p *PostgreSQL) ConnectionURL() string {
	return fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s pool_max_conns=%d",
		p.Username,
		p.Password,
		p.Host,
		p.Port,
		p.Database,
		p.SSLMode,
		p.PoolMaxConns,
	)
}
