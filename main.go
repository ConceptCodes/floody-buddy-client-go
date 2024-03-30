package main

import (
	"floody-buddy/cmd"
	"floody-buddy/pkg/logger"
)

func main() {
	log := logger.New()
	log.Info().Msg("Starting Floody Buddy")

	client.Run()

	select {}
}
