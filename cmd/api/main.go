package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/wittyjudge/blog-service-api/internal/app"
	"github.com/wittyjudge/blog-service-api/internal/config"
)

var configPath = flag.String("config", "./config/api/config.yml", "path to api config file")

func main() {
	flag.Parse()

	config, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to load config file: %w", err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if err := app.StartAPI(ctx, config); err != nil {
		log.Fatal(err)
	}
}
