package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bot/internal/app"
	"bot/internal/config"
)

func main() {
	logger := log.New(os.Stdout, "[bot] ", log.LstdFlags)

	cfg, err := config.Load(".env")
	if err != nil {
		logger.Fatalf("config load error: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.New(cfg, logger).Run(ctx); err != nil {
		logger.Fatalf("run error: %v", err)
	}
}
