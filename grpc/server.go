package main

import (
	"context"
	"encoding/json"
	"log"
	"net"

	pb "grpc/proto"

	"github.com/IBM/sarama"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedProductSaleServiceServer
}

func (s *server) ProcesarVenta(
	ctx context.Context,
	req *pb.ProductSaleRequest,
) (*pb.ProductSaleResponse, error) {

	log.Printf(
		"[gRPC] Categoria=%s Producto=%s Precio=%.2f Cantidad=%d",
		req.Categoria.String(),
		req.ProductoId,
		req.Precio,
		req.CantidadVendida,
	)

	msg, _ := json.Marshal(req)

	kafkaMsg := &sarama.ProducerMessage{
		Topic: "ventas-blackfriday",
		Value: sarama.ByteEncoder(msg),
	}

	partition, offset, err := producer.SendMessage(kafkaMsg)
	if err != nil {
		log.Println("Error enviando a Kafka:", err)
	} else {
		log.Printf("Mensaje enviado a Kafka [partition=%d, offset=%d]",
			partition, offset)
	}

	return &pb.ProductSaleResponse{
		Estado: "OK",
	}, nil
}

func main() {
	brokers := []string{
		"blackfriday-kafka-bootstrap.kafka:9092",
	}

	initKafkaProducer(brokers)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProductSaleServiceServer(grpcServer, &server{})

	log.Println("gRPC Server escuchando en :50051")
	grpcServer.Serve(lis)
}
