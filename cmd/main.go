package client

import (
	"context"
	"floody-buddy/config"
	"floody-buddy/internal/spammer"
	"floody-buddy/pkg/logger"
	"os"
	"os/signal"
	"syscall"
)

func Start() {
	log := logger.New()

	workers := config.AppConfig.Workers

	log.Info().Msgf("Initializing %d Workers", workers)
	log.Info().Msgf("Targeting the Following URLs: %s", config.AppConfig.Domains)

	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < workers; i++ {
		go func() { spammer.Flood(ctx, cancel) }()
	}

	// Wait for a signal to shutdown.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	// Cancel the context, which will stop the Flood operations.
	cancel()
}
