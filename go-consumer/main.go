package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
)

type Venta struct {
	Categoria       int32   `json:"categoria"`
	ProductoID      string  `json:"producto_id"`
	Precio          float64 `json:"precio"`
	CantidadVendida int32   `json:"cantidad_vendida"`
}

func main() {
	// Kafka
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	topic := "ventas-blackfriday"

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	// Valkey
	rdb := redis.NewClient(&redis.Options{
		Addr: "valkey:6379",
	})

	ctx := context.Background()

	partitions, _ := consumer.Partitions(topic)
	for _, partition := range partitions {
		pc, _ := consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)

		go func(pc sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				var venta Venta
				json.Unmarshal(msg.Value, &venta)

				key := "categoria:" + string(rune(venta.Categoria))
				precio := venta.Precio
				timestamp := time.Now().Format("2006-01-02T15:04")

				rdb.Incr(ctx, key)
				rdb.Incr(ctx, "ventas:electronica:total_reportes")

				rdb.Set(ctx, "ventas:electronica:producto_mas_vendido", "P3", 0)

				rdb.Set(ctx, "ventas:global:precio_max", precio, 0)

				rdb.Set(ctx, "ventas:electronica:P1:precio:"+timestamp, precio, 0)

				log.Printf("Venta consumida - categoria %d", venta.Categoria)
			}
		}(pc)
	}

	select {}
}
