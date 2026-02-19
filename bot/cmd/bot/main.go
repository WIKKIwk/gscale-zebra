package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"bot/internal/app"
	"bot/internal/config"
	"core/workflowlog"
)

func main() {
	logs, err := workflowlog.New("bot")
	if err != nil {
		log.Fatalf("workflow logger init error: %v", err)
	}
	defer logs.Close()

	logger := logs.Logger("main")
	runLogger := logs.Logger("worker.run")
	batchLogger := logs.Logger("worker.batch")
	callbackLogger := logs.Logger("worker.callback")
	cleanupLogger := logs.Logger("worker.cleanup")
	logger.Printf("workflow logs dir: %s", logs.Dir())

	cfg, err := config.Load(".env")
	if err != nil {
		logger.Fatalf("config load error: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.New(cfg, logger, runLogger, batchLogger, callbackLogger, cleanupLogger).Run(ctx); err != nil {
		logger.Fatalf("run error: %v", err)
	}
}
