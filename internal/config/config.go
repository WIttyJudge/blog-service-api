package config

import (
	"fmt"
	"net"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Environment string    `yaml:"environment" env:"ENVIRONMENT" env-default:"development"`
	Databases   Databases `yaml:"databases"`
	API         API       `yaml:"api"`
}

type Databases struct {
	PostgreSQL PostgreSQL `yaml:"postgresql"`
}

type PostgreSQL struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	Password     string `yaml:"password"`
	Username     string `yaml:"username"`
	Database     string `yaml:"database"`
	SSLMode      string `yaml:"ssl_mode"`
	PoolMaxConns int    `yaml:"pool_max_conns"`
}

type API struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
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
