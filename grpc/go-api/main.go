package main

import (
	"context"
	"encoding/json"
	pb "grpc/proto"
	"log"
	"net/http"
	"os"

	"google.golang.org/grpc"
)

func mapCategoria(cat string) pb.CategoriaProducto {
	switch cat {
	case "Electronica":
		return pb.CategoriaProducto_ELECTRONICA
	case "Ropa":
		return pb.CategoriaProducto_ROPA
	case "Hogar":
		return pb.CategoriaProducto_HOGAR
	case "Belleza":
		return pb.CategoriaProducto_BELLEZA
	default:
		return pb.CategoriaProducto_CATEGORIA_PRODUCTO_UNSPECIFIED
	}
}

func enviarAGrpc(venta Venta) error {
	addr := os.Getenv("GRPC_SERVER_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())

	if err != nil {
		return err
	}

	defer conn.Close()

	client := pb.NewProductSaleServiceClient(conn)

	_, err = client.ProcesarVenta(
		context.Background(),
		&pb.ProductSaleRequest{
			Categoria:       mapCategoria(venta.Categoria),
			ProductoId:      venta.ProductoID,
			Precio:          venta.Precio,
			CantidadVendida: venta.CantidadVendida,
		},
	)

	return err
}

type Venta struct {
	Categoria       string  `json:"categoria"`
	ProductoID      string  `json:"producto_id"`
	Precio          float64 `json:"precio"`
	CantidadVendida int32   `json:"cantidad_vendida"`
}

func procesarVenta(w http.ResponseWriter, r *http.Request) {
	var venta Venta

	json.NewDecoder(r.Body).Decode(&venta)

	err := enviarAGrpc(venta)
	if err != nil {
		http.Error(w, "Error gRPC", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"enviado a grpc"}`))
}

func main() {
	http.HandleFunc("/procesar", procesarVenta)

	log.Println("Go API escuchando en :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
