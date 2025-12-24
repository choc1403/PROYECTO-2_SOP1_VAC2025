package main

import (
	"context"
	"log"
	"net"

	pb "grpc/proto"

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

	// 👉 aquí luego Kafka

	return &pb.ProductSaleResponse{
		Estado: "OK",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProductSaleServiceServer(grpcServer, &server{})

	log.Println("gRPC Server escuchando en :50051")
	grpcServer.Serve(lis)
}
