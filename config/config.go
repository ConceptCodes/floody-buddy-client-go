package config

import (
	"floody-buddy/pkg/logger"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Domains    string `env:"DOMAINS" envDefault:"0.0.0.0"`
	Timeout    int    `env:"HTTP_TIMEOUT" envDefault:"15"`
	MaxRetries int    `env:"MAX_RETRIES" envDefault:"5"`
	Topic      string `env:"KAFKA_TOPIC" envDefault:"test"`
	Brokers    string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	Workers    int    `env:"WORKERS" envDefault:"10"`
}

var AppConfig = Config{}

func init() {
	log := logger.New()
	log.Debug().Msg("Loading env vars")

	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error while loading env vars")
	}
	log.Debug().Msg("Env vars loaded")

	err = env.Parse(&AppConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Error while parsing env vars")
	}
	log.Debug().Msg("Env vars parsed")
}
