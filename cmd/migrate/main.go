package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/wittyjudge/blog-service-api/internal/config"
)

var (
	pathFlag   = flag.String("path", "migrations/", "path to migrations folder")
	configPath = flag.String("config", "./config/api/config.yml", "path to api config file")
)

func main() {
	flag.Parse()

	config, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal(fmt.Errorf("failed ot load config file: %w", err))
	}

	err = runMigration(config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done!")
}

func runMigration(config *config.Config) error {
	dir := fmt.Sprintf("file://%s", *pathFlag)
	m, err := migrate.New(dir, config.Databases.Postgres.ConnectionURL())
	if err != nil {
		return fmt.Errorf("failed to initialize migration: %w", err)
	}

	m.Log = newLogger()

	if err := m.Up(); err != nil {
		return fmt.Errorf("failed to run migration: %w", err)
	}

	srcErr, dbErr := m.Close()
	if srcErr != nil {
		return fmt.Errorf("migrate source error: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("migrate database error: %w", dbErr)
	}

	return nil
}

type logger struct {
	logger *log.Logger
}

func newLogger() *logger {
	return &logger{
		logger: log.New(os.Stdout, "migrate", log.LstdFlags),
	}
}

func (l *logger) Printf(arg string, vars ...interface{}) {
	l.logger.Printf(arg, vars...)
}

func (l *logger) Verbose() bool {
	return true
}
