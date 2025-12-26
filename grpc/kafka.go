package main

import (
	"log"

	"github.com/IBM/sarama"
)

var producer sarama.SyncProducer // Variable global para que server.go la vea

func initKafkaProducer(brokers []string) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	var err error
	producer, err = sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Error creando productor Kafka: %v", err)
	}
}
