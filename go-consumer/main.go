package main

import (
	"context"
	"encoding/json"
	"fmt"
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

func categoriaToString(c int32) string {
	switch c {
	case 1:
		return "electronica"
	case 2:
		return "ropa"
	case 3:
		return "hogar"
	case 4:
		return "belleza"
	default:
		return "desconocida"
	}
}

func main() {
	log.Println("Esperando para procesar...")
	ctx := context.Background()

	// Configuración de Valkey (Redis)
	rdb := redis.NewClient(&redis.Options{
		Addr: "valkey.backend.svc.cluster.local:6379", // Ajusta si es necesario
	})

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

	partitions, _ := consumer.Partitions(topic)
	for _, partition := range partitions {
		pc, _ := consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)

		go func(pc sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				var venta Venta
				if err := json.Unmarshal(msg.Value, &venta); err != nil {
					log.Printf("Error decodificando: %v", err)
					continue
				}

				catName := categoriaToString(venta.Categoria)

				// 1. PRECIO MÁXIMO Y MÍNIMO GLOBAL
				// Usamos un script de Lua o comparaciones simples
				actualizarMinMax(ctx, rdb, venta.Precio)

				// 2. PRODUCTO MÁS/MENOS VENDIDO (GLOBAL Y POR CATEGORÍA)
				// ZIncrBy suma la cantidad vendida al score del producto
				rdb.ZIncrBy(ctx, "ranking:global", float64(venta.CantidadVendida), venta.ProductoID)
				rdb.ZIncrBy(ctx, "ranking:"+catName, float64(venta.CantidadVendida), venta.ProductoID)

				// 3. DATOS PARA PRECIO PROMEDIO Y PRODUCTO PROMEDIO
				// Guardamos la suma acumulada y el conteo para calcular el promedio después
				rdb.HIncrByFloat(ctx, "stats:precio:suma", catName, venta.Precio)
				rdb.HIncrByFloat(ctx, "stats:cantidad:suma", catName, float64(venta.CantidadVendida))
				rdb.HIncrBy(ctx, "stats:conteo", catName, 1)

				// 4. TOTAL DE REPORTES POR CATEGORÍA
				rdb.Incr(ctx, "reportes:total:"+catName)

				// 5. VARIACIÓN DE PRECIO (SOLO ELECTRÓNICA - TIME SERIES)
				if catName == "electronica" {
					timestamp := time.Now().Unix()
					// Guardamos el precio con el tiempo como Score
					// Key: history:electronica:P1 (por ejemplo)
					keyHistorial := fmt.Sprintf("history:electronica:%s", venta.ProductoID)
					rdb.ZAdd(ctx, keyHistorial, redis.Z{
						Score:  float64(timestamp),
						Member: fmt.Sprintf("%f:%d", venta.Precio, timestamp),
					})
					// Opcional: Mantener solo los últimos 20 registros para la gráfica
					rdb.ZRemRangeByRank(ctx, keyHistorial, 0, -21)
				}

				log.Printf("Procesada venta de %s - Producto: %s", catName, venta.ProductoID)
			}
		}(pc)
	}

	select {}
}

func actualizarMinMax(ctx context.Context, rdb *redis.Client, precio float64) {
	// Lógica para el máximo
	valMax, errMax := rdb.Get(ctx, "ventas:global:precio_max").Float64()
	if errMax == redis.Nil || precio > valMax {
		rdb.Set(ctx, "ventas:global:precio_max", precio, 0)
	}

	// Lógica para el mínimo
	valMin, errMin := rdb.Get(ctx, "ventas:global:precio_min").Float64()
	if errMin == redis.Nil || precio < valMin {
		rdb.Set(ctx, "ventas:global:precio_min", precio, 0)
	}
}
