package main

import (
	"log"

	"github.com/IBM/sarama"
)

var producer sarama.SyncProducer

func initKafkaProducer(brokers []string) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	p, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Error creando producer Kafka: %v", err)
	}

	producer = p
	log.Println("Kafka Producer inicializado")
}
