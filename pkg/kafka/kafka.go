package kafka

import (
	"errors"
	"strings"

	"github.com/IBM/sarama"

	"floody-buddy/config"
	"floody-buddy/pkg/logger"
)

var err error
var client sarama.Client

func PublishMessage(message string) error {
	log := logger.New()
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true

	log.Debug().Msg("Publishing message to Kafka")

	brokers := strings.Split(config.AppConfig.Brokers, ",")

	if len(brokers) == 0 {
		err = errors.New("no brokers found")
		log.Error().Err(err).Msg("Unable to parse brokers from env")
		return err
	}

	client, err = sarama.NewClient(brokers, kafkaConfig)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create client")
		return err
	}

	producer, err := sarama.NewSyncProducerFromClient(client)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create producer")
		return err
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: config.AppConfig.Topic,
		Value: sarama.StringEncoder(message),
	}

	_, _, err = producer.SendMessage(msg)
	return err
}
