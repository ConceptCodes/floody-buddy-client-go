package client

import (
	"floody-buddy/internal/spammer"
	"floody-buddy/pkg/logger"
	"floody-buddy/config"
)

func Run() {
	log := logger.New()

	workers := config.AppConfig.Workers
	
	log.Info().Msgf("Initializing %d Workers", workers)

	for i := 0; i < workers; i++ {
		go func() { spammer.Flood() }()
	}
}
