package main

import (
	client "floody-buddy/cmd"
	"floody-buddy/internal/constants"
	"floody-buddy/pkg/logger"
)

func main() {
	log := logger.New()
	log.Info().Msgf("Starting Floody Buddy v%s", constants.VERSION)

	client.Start()

	select {}
}
