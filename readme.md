
# Proyecto 2 SOP1

## 1. Introducción

Este proyecto implementa un **sistema distribuido basado en microservicios**, orientado al procesamiento y análisis de ventas durante un escenario de alta carga (Black Friday).  
La solución integra **API REST en Rust**, **gRPC en Go**, **Kafka** como sistema de mensajería, **Valkey (Redis-compatible)** como almacenamiento en memoria y **Grafana** para visualización, todo desplegado sobre **Kubernetes**.


## 2. Flujo de Arquitectura y Componentes del Sistema

### 1. Generación de Tráfico y Punto de Entrada

**Locust**
El flujo inicia con Locust, una herramienta de pruebas de carga utilizada para simular múltiples usuarios concurrentes. Locust genera solicitudes que representan *tweets de ventas* o eventos de compra, permitiendo evaluar el comportamiento del sistema bajo alta demanda.

**Ingress de Kubernetes**
Las solicitudes generadas por Locust ingresan al clúster a través del Ingress de Kubernetes, el cual actúa como punto de entrada único. El Ingress se encarga de enrutar el tráfico HTTP hacia los servicios internos correspondientes, aplicando reglas de encaminamiento y balanceo de carga.

---

### 2. Capa de APIs y Comunicación

En esta etapa, el sistema ofrece **dos puntos de entrada**, diseñados para comparar enfoques de comunicación.

#### API REST en Rust

Esta API recibe las solicitudes HTTP generadas por Locust. Su función principal es:

* Validar y procesar los datos de entrada.
* Transformar la información en un formato interno.
* Enviar los datos hacia el gRPC Server para su posterior procesamiento.

Esta API representa un enfoque **tradicional REST**, útil para analizar su rendimiento frente a gRPC.

---

#### API en Go como gRPC Client

Esta API también recibe solicitudes externas, pero su responsabilidad principal es actuar como **cliente gRPC**.
Sus funciones son:

* Recibir peticiones HTTP o eventos simulados.
* Convertir los datos al formato definido en el archivo `.proto`.
* Invocar de forma remota los métodos expuestos por el **gRPC Server**.

El uso de gRPC Client permite:

* Comunicación eficiente entre servicios.
* Menor latencia gracias a Protobuf.
* Evaluar el impacto de gRPC frente a REST.

---

### 3. gRPC Server y Mensajería

#### gRPC Server (Kafka Writer)

El **gRPC Server** actúa como un componente central del sistema. Su función es:

* Recibir las solicitudes provenientes de los clientes gRPC (API REST en Rust y API en Go).
* Ejecutar la lógica de negocio necesaria.
* Publicar los eventos procesados en Kafka utilizando un **Kafka Writer**.

Este componente desacopla las APIs del sistema de mensajería, permitiendo escalabilidad y tolerancia a fallos.

---

#### Strimzi Kafka

Kafka es el núcleo de la arquitectura orientada a eventos.
Sus responsabilidades incluyen:

* Recibir los mensajes publicados por el gRPC Server.
* Almacenarlos de manera ordenada y distribuida.
* Garantizar durabilidad y tolerancia a fallos.

**Strimzi** es el operador de Kubernetes encargado de administrar Kafka dentro del clúster, facilitando su despliegue, escalado y mantenimiento.

---

### 4. Consumo, Almacenamiento y Visualización

#### Kafka Consumer (Go)

El Kafka Consumer, implementado en Go y desplegado con 1 a 2 réplicas, se encarga de:

* Consumir los mensajes desde Kafka.
* Procesar y agregar la información (contadores, máximos, series temporales).
* Preparar los datos para su almacenamiento.

El escalado de este componente impacta directamente en la capacidad de procesamiento del sistema.

---

#### Valkey DB

Valkey, un fork de Redis de alto rendimiento, se utiliza como base de datos en memoria para:

* Almacenar métricas agregadas.
* Proveer acceso rápido a los datos.
* Servir como backend para la visualización en tiempo real.


---

#### Grafana

Grafana es la capa de visualización del sistema.
Sus funciones incluyen:

* Consultar los datos almacenados en Valkey.
* Mostrar dashboards en tiempo real.
* Permitir el análisis del comportamiento del sistema bajo carga.

Gracias a Grafana, es posible observar métricas clave como volumen de ventas, categorías más utilizadas y evolución temporal de los eventos.

---

### 5. Resumen del Rol de gRPC

* **gRPC Client**:
  Actúa como intermediario eficiente entre las APIs y el backend, enviando datos de forma optimizada al servidor gRPC.

* **gRPC Server**:
  Centraliza la lógica de negocio y desacopla la capa de entrada del sistema de mensajería, publicando los eventos en Kafka.

Este diseño permite comparar **REST vs gRPC**, evaluar rendimiento y construir un sistema altamente escalable y desacoplado.




### Diagrama (referencia)



---

## 3. Documentación de Deployments

### 3.1 Rust API (REST)

- **Tipo**: Stateless
- **Puerto**: 8080
- **Función**: Recibir ventas y publicarlas en Kafka.
- **Escalabilidad**: HPA basado en CPU.

Ejemplo de recursos:

```yaml
resources:
  requests:
    cpu: "100m"
    memory: "256Mi"
  limits:
    cpu: "500m"
    memory: "512Mi"
```

---

### 3.2 Go API / Go Server (gRPC)

* **Tipo**: Stateless
* **Puerto**: 8081 go client 50051 go server
* **Función**: Comunicación gRPC de alto rendimiento.
* **Escalabilidad**: HPA basado en CPU.

Ventaja principal:

* Menor latencia y mejor rendimiento frente a REST bajo alta carga.

---

### 3.3 Kafka Consumer (Go)

* **Tipo**: Stateless
* **Función**:

  * Consumir eventos desde Kafka.
  * Calcular métricas (totales, máximos, series temporales).
  * Guardar resultados en Valkey.
* **Escalabilidad**: HPA (impacta directamente el throughput).

---

### 3.4 Valkey

* **Tipo**: Stateful
* **Puerto**: 6379
* **Función**: Almacenamiento en memoria de métricas.
* **Escalabilidad**:

  * No se usa HPA (estado).
  * Se mantiene una sola réplica para consistencia.

---

### 3.5 Grafana

* **Tipo**: Visualización
* **Puerto**: 3000
* **Función**: Mostrar métricas en tiempo real desde Valkey.
* **Notas**:

  * Datasource Redis configurado apuntando al Service de Valkey.
  * No requiere escalado automático.

---

## 4. Instrucciones para Desplegar el Sistema

### 4.1 Prerrequisitos
* Maquina Virutal en la nube, con un sistema operativo de ubuntu de preferencia
* Kubernetes funcional (CLUSTER EN LA NUBE)
* Docker
* kubectl
* Kafka desplegado
* Metrics Server habilitado

---

### 4.2 Despliegue paso a paso


```bash
cd k8s
# Namespace
kubectl create namespace backend
kubectl create ns kafka

# Ingress
kubectl delete -A ValidatingWebhookConfiguration ingress-nginx-admission
kubectl apply -f ingress.yml -n backend

# Kafka
kubectl apply -f kafka-cluster.yml -n kafka

# Deployments
# En cada carpeta. /rust-api /go-api /go-server /go-consumer /valkey /grafana
kubectl apply -f deployment.yml -n backend 
# HPA
kubectl apply -f deployment.yml -n backend 
```

Verificación:

```bash
kubectl get pods -n backend
kubectl get svc -n backend
kubectl get hpa -n backend
```



### 4.3 Pruebas del sistema

1. Generar carga con Locust.
2. Verificar consumo en Kafka Consumer.
3. Validar datos en Valkey:

   ```bash
   kubectl exec -it deployment/valkey -n backend -- redis-cli
   ```
4. Acceder a Grafana y visualizar dashboards.

---

## 5. Retos Encontrados y Soluciones

### 5.1 Descarga de imágenes desde ZOT Registry

**Problema**
Al hacer `docker pull` desde el registry privado (ZOT), Kubernetes no podía descargar las imágenes correctamente debido a que el acceso se realizaba mediante **HTTP**, lo cual generaba problemas de seguridad y compatibilidad.

Inicialmente se intentó usar **ngrok**, pero este exponía el servicio en HTTP, lo que impedía el pull correcto de imágenes.

**Solución**
Se migró a **Cloudflare Tunnel (cloudflared)** para exponer el registry de forma segura mediante **HTTPS**.

```bash
# Descargar e instalar cloudflared
curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared.deb

# Crear túnel rápido
cloudflared tunnel --url http://localhost:5000
```

Gracias a esto:

* Kubernetes pudo descargar imágenes vía HTTPS.
* Se eliminó la dependencia de ngrok.
* Se mejoró la seguridad del despliegue.

---

### 5.2 Conectividad Grafana – Valkey

**Problema**

* Grafana no lograba conectarse a Valkey usando nombres DNS.
* Errores de `no such host` y `i/o timeout`.

**Solución**

* Ajustar Valkey para escuchar en `0.0.0.0`.
* Verificar conectividad con `nc`.
* Usar correctamente el Service de Kubernetes.

---

## 6. Rendimiento y Comparativas

### 6.1 API REST vs gRPC

| Característica | REST (Rust) | gRPC (Go) |
| -------------- | ----------- | --------- |
| Latencia       | Media       | Baja      |
| Serialización  | JSON        | Protobuf  |
| Rendimiento    | Bueno       | Excelente |
| Uso CPU        | Mayor       | Menor     |

**Conclusión**: gRPC ofrece mejor rendimiento bajo alta carga.

---

### 6.2 Kafka

* Permite desacoplar productores y consumidores.
* Soporta picos de tráfico sin pérdida de mensajes.
* Escalar consumidores incrementa el throughput.

---

### 6.3 Valkey y Réplicas

* Una sola réplica garantiza consistencia.
* Incrementar réplicas mejora lectura, pero complica escritura.
* Ideal para métricas rápidas en tiempo real.

---

## 7. Descripicion de codigos

### Descripción del Código – API REST en Rust (Actix Web)

Este código implementa una **API REST en Rust** utilizando el framework **Actix Web**.
Su función principal es **recibir eventos de venta vía HTTP**, procesarlos de forma básica y **reenviarlos a un servicio escrito en Go**, que actúa como parte del backend del sistema distribuido.

---

### 1. Importación de Dependencias

```rust
use actix_web::{post, web, App, HttpServer, HttpResponse, Responder};
use serde::{Deserialize, Serialize};
```

* **actix_web**: Framework web asíncrono utilizado para construir la API REST.

  * `post`: Macro para definir rutas HTTP POST.
  * `web`: Manejo de datos compartidos y JSON.
  * `App` y `HttpServer`: Configuración y arranque del servidor HTTP.
* **serde**: Librería para serialización y deserialización de datos.

  * Permite convertir estructuras Rust a JSON y viceversa.

---

### 2. Definición de la Estructura `Venta`

```rust
#[derive(Deserialize, Serialize)]
struct Venta {
    categoria: String,
    producto_id: String,
    precio: f64,
    cantidad_vendida: i32,
}
```

Esta estructura representa el **modelo de datos de una venta**.

* `categoria`: Categoría del producto (por ejemplo, electrónica, ropa).
* `producto_id`: Identificador del producto.
* `precio`: Precio de la venta.
* `cantidad_vendida`: Cantidad de unidades vendidas.

El uso de `Deserialize` y `Serialize` permite:

* Recibir datos en formato JSON desde el cliente.
* Enviar el mismo objeto como JSON a otros servicios.

---

### 3. Función `enviar_a_go`

```rust
async fn enviar_a_go(venta: &Venta) -> Result<(), reqwest::Error> {
```

Esta función se encarga de **reenviar la información de la venta a un servicio en Go**.

#### Responsabilidades:

* Obtener dinámicamente la URL del servicio Go desde una variable de entorno (`GO_SERVICE_URL`).
* Crear un cliente HTTP asíncrono.
* Enviar la venta como JSON mediante una petición POST.

```rust
let go_service_url = std::env::var("GO_SERVICE_URL")
    .unwrap_or_else(|_| "http://localhost:8081".to_string());
```

* Permite flexibilidad entre entornos:

  * Local
  * Kubernetes
  * Producción

```rust
client
    .post(&format!("{}/procesar", go_service_url))
    .json(venta)
    .send()
    .await?;
```

* Envía los datos al endpoint `/procesar` del servicio Go.
* La operación es **asíncrona y no bloqueante**.

---

### 4. Endpoint `/venta`

```rust
#[post("/venta")]
async fn recibir_venta(venta: web::Json<Venta>) -> impl Responder {
```

Este endpoint es el **punto de entrada principal** de la API.

#### Flujo:

1. Recibe una venta en formato JSON.
2. La convierte automáticamente a la estructura `Venta`.
3. Llama a la función `enviar_a_go` para reenviar la información.
4. Retorna una respuesta HTTP adecuada según el resultado.

```rust
if let Err(e) = enviar_a_go(&venta).await {
    eprintln!("Error enviando a Go: {}", e);
    return HttpResponse::InternalServerError().body("Error enviando a Go");
}
```

* Manejo de errores para garantizar robustez.
* Si el servicio Go no responde, se devuelve un error HTTP 500.

```rust
HttpResponse::Ok().json("Venta procesada")
```

* Respuesta exitosa en formato JSON.

---

### 5. Función `main` – Arranque del Servidor

```rust
#[actix_web::main]
async fn main() -> std::io::Result<()> {
```

Este es el punto de entrada de la aplicación.

```rust
HttpServer::new(|| {
    App::new()
        .service(recibir_venta) 
})
.bind(("0.0.0.0", 8080))?
.workers(4)
.run()
.await
```

#### Configuración:

* **Puerto**: 8080
* **Interfaz**: `0.0.0.0` (accesible dentro del clúster Kubernetes).
* **Workers**: 4 hilos para manejo concurrente.
* Registra el endpoint `/venta`.

---

## Rol del Servicio en la Arquitectura

* Actúa como **API REST de entrada**.
* Recibe tráfico generado por Locust o clientes externos.
* Desacopla la entrada HTTP del procesamiento interno.
* Reenvía los eventos hacia servicios en Go (gRPC o Kafka writer).

Este diseño permite:

* Escalar horizontalmente el servicio.
* Comparar REST frente a gRPC.
* Mantener una arquitectura modular y desacoplada.

---


## Descripción del Código – API Go como gRPC Client

Este archivo implementa una **API HTTP en Go** que actúa como **cliente gRPC**.
Su función principal es **recibir solicitudes HTTP**, transformar los datos al formato definido en **Protocol Buffers**, y **enviarlos al gRPC Server**, que se encarga del procesamiento y publicación en Kafka.

Este componente sirve como **puente entre el mundo REST/HTTP y gRPC**.

---

### 1. Importación de Dependencias

```go
import (
	"context"
	"encoding/json"
	pb "grpc/proto"
	"log"
	"net/http"
	"os"

	"google.golang.org/grpc"
)
```

* **context**: Permite controlar el ciclo de vida de las llamadas gRPC.
* **encoding/json**: Decodificación de datos JSON recibidos por HTTP.
* **pb "grpc/proto"**: Código generado automáticamente desde el archivo `.proto`.
* **net/http**: Implementación del servidor HTTP.
* **grpc**: Librería oficial de gRPC para Go.

---

### 2. Función `mapCategoria`

```go
func mapCategoria(cat string) pb.CategoriaProducto {
```

Esta función traduce una **categoría en texto** (recibida vía JSON) a un **valor del enum definido en Protobuf**.

#### Objetivo:

* Garantizar compatibilidad entre la API HTTP y el contrato gRPC.
* Evitar errores de tipado al enviar datos al servidor gRPC.

Ejemplo:

* `"Electronica"` → `ELECTRONICA`
* `"Ropa"` → `ROPA`

En caso de no coincidencia, se asigna un valor por defecto:

```go
pb.CategoriaProducto_CATEGORIA_PRODUCTO_UNSPECIFIED
```

---

### 3. Función `enviarAGrpc`

```go
func enviarAGrpc(venta Venta) error {
```

Esta función encapsula toda la lógica de **comunicación con el gRPC Server**.

#### Flujo interno:

1. **Obtención de la dirección del gRPC Server**

   ```go
   addr := os.Getenv("GRPC_SERVER_ADDR")
   ```

   * Permite configurar el destino dinámicamente (local o Kubernetes).

2. **Creación de la conexión gRPC**

   ```go
   conn, err := grpc.Dial(addr, grpc.WithInsecure())
   ```

   * Establece una conexión cliente–servidor.
   * En un entorno productivo podría usarse TLS.

3. **Creación del cliente gRPC**

   ```go
   client := pb.NewProductSaleServiceClient(conn)
   ```

4. **Invocación remota del método `ProcesarVenta`**

   ```go
   client.ProcesarVenta(context.Background(), &pb.ProductSaleRequest{...})
   ```

   * Se envía la venta en formato Protobuf.
   * Se ejecuta una llamada RPC síncrona.

---

### 4. Estructura `Venta`

```go
type Venta struct {
	Categoria       string  `json:"categoria"`
	ProductoID      string  `json:"producto_id"`
	Precio          float64 `json:"precio"`
	CantidadVendida int32   `json:"cantidad_vendida"`
}
```

Define el **modelo de datos recibido por HTTP**.

* Coincide con el JSON enviado por la API en Rust.
* Se transforma posteriormente al mensaje Protobuf.

---

### 5. Handler HTTP `procesarVenta`

```go
func procesarVenta(w http.ResponseWriter, r *http.Request) {
```

Este endpoint actúa como **punto de entrada REST** para esta API.

#### Flujo del handler:

1. Decodifica el cuerpo JSON de la solicitud:

   ```go
   json.NewDecoder(r.Body).Decode(&venta)
   ```

2. Llama a la función `enviarAGrpc`:

   * Si ocurre un error, responde con HTTP 500.
   * Si es exitoso, retorna HTTP 200.

3. Registra la operación en logs:

   ```go
   log.Println("Procesando Venta")
   ```

---

### 6. Función `main` – Arranque del Servicio

```go
func main() {
	http.HandleFunc("/procesar", procesarVenta)

	log.Println("Go API escuchando en :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
```

Configuración del servidor:

* **Puerto**: 8081
* **Endpoint**: `/procesar`
* **Tipo**: HTTP REST

Este servicio queda a la espera de solicitudes provenientes de la API en Rust u otros clientes.

---

## Rol del gRPC Client en la Arquitectura

* Actúa como **adaptador entre REST y gRPC**.
* Convierte datos JSON a Protobuf.
* Encapsula la comunicación con el gRPC Server.
* Permite comparar el rendimiento de REST vs gRPC.
* Facilita una arquitectura desacoplada y escalable.

---



## Descripción del Código – gRPC Server en Go 

Este archivo implementa un **servidor gRPC en Go** cuya responsabilidad principal es **recibir solicitudes desde clientes gRPC**, procesarlas y **publicar los eventos en Kafka**.
Actúa como un **punto central de negocio**, desacoplando las APIs de entrada del sistema de mensajería.

---

### 1. Importación de Dependencias

```go
import (
	"context"
	"encoding/json"
	"log"
	"net"

	pb "grpc/proto"

	"github.com/IBM/sarama"
	"google.golang.org/grpc"
)
```

* **context**: Manejo del ciclo de vida de las llamadas gRPC.
* **encoding/json**: Serialización de mensajes antes de enviarlos a Kafka.
* **net**: Creación del listener TCP.
* **pb "grpc/proto"**: Código generado desde el archivo `.proto`.
* **sarama**: Librería cliente Kafka (producer).
* **grpc**: Framework gRPC para Go.

---

### 2. Definición de la Estructura `server`

```go
type server struct {
	pb.UnimplementedProductSaleServiceServer
}
```

* Implementa la interfaz del servicio gRPC definida en el `.proto`.
* `UnimplementedProductSaleServiceServer` garantiza compatibilidad futura si se agregan nuevos métodos.

---

### 3. Método gRPC `ProcesarVenta`

```go
func (s *server) ProcesarVenta(
	ctx context.Context,
	req *pb.ProductSaleRequest,
) (*pb.ProductSaleResponse, error) {
```

Este método es invocado **remotamente por los gRPC Clients**.
Representa el **punto de entrada del backend** para procesar una venta.

---

#### 3.1 Registro de la solicitud

```go
log.Printf(
	"[gRPC] Categoria=%s Producto=%s Precio=%.2f Cantidad=%d",
	req.Categoria.String(),
	req.ProductoId,
	req.Precio,
	req.CantidadVendida,
)
```

* Permite auditoría y trazabilidad.
* Facilita debugging y observabilidad del sistema.

---

#### 3.2 Serialización del mensaje

```go
msg, _ := json.Marshal(req)
```

* Convierte el mensaje Protobuf a JSON.
* Facilita interoperabilidad con Kafka y otros consumidores.

---

#### 3.3 Creación del mensaje Kafka

```go
kafkaMsg := &sarama.ProducerMessage{
	Topic: "ventas-blackfriday",
	Value: sarama.ByteEncoder(msg),
}
```

* Define el tópico de Kafka.
* El contenido es el evento de venta serializado.

---

#### 3.4 Envío del mensaje a Kafka

```go
partition, offset, err := producer.SendMessage(kafkaMsg)
```

* Publica el evento en Kafka.
* Kafka se encarga de:

  * Persistencia
  * Orden
  * Distribución a consumidores

```go
log.Printf("Mensaje enviado a Kafka [partition=%d, offset=%d]",
	partition, offset)
```

* El `partition` y `offset` confirman la entrega exitosa.

---

#### 3.5 Respuesta al cliente gRPC

```go
return &pb.ProductSaleResponse{
	Estado: "OK",
}, nil
```

* Devuelve una respuesta simple indicando éxito.
* Permite al cliente continuar su flujo.

---

### 4. Función `main` – Inicialización del Servidor

```go
func main() {
```

#### 4.1 Configuración de Kafka

```go
brokers := []string{
	"blackfriday-kafka-bootstrap.kafka:9092",
}
```

* Dirección del clúster Kafka gestionado por **Strimzi**.
* Uso del Service interno de Kubernetes.

```go
initKafkaProducer(brokers)
```

* Inicializa el producer Kafka usando Sarama.
* Se reutiliza la conexión para todas las solicitudes.

---

#### 4.2 Creación del listener TCP

```go
lis, err := net.Listen("tcp", ":50051")
```

* Puerto estándar para gRPC.
* Expone el servicio dentro del clúster.

---

#### 4.3 Arranque del servidor gRPC

```go
grpcServer := grpc.NewServer()
pb.RegisterProductSaleServiceServer(grpcServer, &server{})
```

* Crea la instancia del servidor gRPC.
* Registra el servicio definido en el `.proto`.

```go
grpcServer.Serve(lis)
```

* Inicia el servidor y queda a la espera de solicitudes.

---

## Rol del gRPC Server en la Arquitectura

* ✔ Recibe solicitudes desde múltiples clientes gRPC
* ✔ Centraliza la lógica de negocio
* ✔ Publica eventos en Kafka
* ✔ Desacopla APIs de entrada y consumidores
* ✔ Permite escalado horizontal

Este componente es **stateless**, lo que lo hace ideal para usar **HPA** en Kubernetes.

---


## Descripción del Código – Kafka Producer en Go

Este archivo define la **inicialización del productor Kafka** utilizado por el **gRPC Server** para publicar eventos de ventas en el sistema de mensajería.
Su objetivo es **crear y mantener una conexión reutilizable con Kafka**, permitiendo enviar mensajes de forma confiable y eficiente.

---

### 1. Importación de Dependencias

```go
import (
	"log"

	"github.com/IBM/sarama"
)
```

* **sarama**: Librería cliente de Kafka para Go.

  * Proporciona APIs para productores y consumidores.
* **log**: Registro de eventos y errores del sistema.

---

### 2. Declaración del Productor Global

```go
var producer sarama.SyncProducer
```

* Define un **productor Kafka síncrono**.
* Se declara como **variable global** para:

  * Compartir la misma instancia entre múltiples archivos (`server.go`).
  * Evitar crear una conexión Kafka por cada solicitud gRPC.
  * Mejorar rendimiento y reducir sobrecarga.

Este diseño es común en microservicios que producen mensajes con alta frecuencia.

---

### 3. Función `initKafkaProducer`

```go
func initKafkaProducer(brokers []string) {
```

Esta función se encarga de **inicializar el productor Kafka** al arranque del servicio.

---

#### 3.1 Configuración del Productor

```go
config := sarama.NewConfig()
config.Producer.Return.Successes = true
```

* `Return.Successes = true`:

  * Permite obtener confirmación de Kafka cuando un mensaje es enviado exitosamente.
  * Es necesario para poder recibir `partition` y `offset`.

Esto mejora:

* Confiabilidad
* Observabilidad
* Trazabilidad de eventos

---

#### 3.2 Creación del Productor

```go
producer, err = sarama.NewSyncProducer(brokers, config)
```

* Crea un **productor síncrono** conectado al clúster Kafka.
* `brokers` contiene las direcciones del clúster Kafka administrado por **Strimzi**.

Ejemplo:

```text
blackfriday-kafka-bootstrap.kafka:9092
```

---

#### 3.3 Manejo de Errores

```go
if err != nil {
	log.Fatalf("Error creando productor Kafka: %v", err)
}
```

* Si el productor no puede inicializarse:

  * El servicio falla de forma controlada.
  * Evita que el sistema arranque en un estado inconsistente.

---

## Rol del Kafka Producer en la Arquitectura

* Actúa como **puente entre el gRPC Server y Kafka**.
* Publica eventos de ventas en el tópico correspondiente.
* Garantiza desacoplamiento entre productores y consumidores.
* Permite absorber picos de tráfico sin pérdida de datos.

Este componente es fundamental para:

* Procesamiento asíncrono
* Escalabilidad
* Alta disponibilidad

---

## Relación con Strimzi

* **Strimzi** gestiona el clúster Kafka dentro de Kubernetes.
* **Sarama** es el cliente que se conecta a ese clúster.
* El productor no depende de Strimzi directamente, solo de la dirección del broker.

---


## Descripción del Código – Kafka Consumer en Go con Valkey

Este archivo implementa un **consumidor Kafka en Go** cuya responsabilidad es **procesar los eventos de ventas publicados en Kafka**, calcular métricas agregadas y **almacenarlas en Valkey** para su posterior visualización en Grafana.

Este componente permite desacoplar el procesamiento de datos del flujo principal y soportar **alta concurrencia y escalabilidad**.

---

### 1. Importación de Dependencias

```go
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
```

* **sarama**: Cliente Kafka para consumir mensajes.
* **redis/go-redis**: Cliente oficial compatible con Valkey.
* **context**: Manejo de operaciones asíncronas.
* **encoding/json**: Deserialización de mensajes Kafka.
* **time**: Manejo de timestamps para series temporales.

---

### 2. Estructura `Venta`

```go
type Venta struct {
	Categoria       int32
	ProductoID      string
	Precio          float64
	CantidadVendida int32
}
```

Representa el **evento de venta** recibido desde Kafka.
Su estructura coincide con los datos enviados por el gRPC Server.

---

### 3. Función `categoriaToString`

```go
func categoriaToString(c int32) string
```

Convierte el valor numérico de la categoría (enum Protobuf) a una **representación legible**.

Ejemplo:

* `1` → `electronica`
* `2` → `ropa`

Esta conversión se utiliza para:

* Crear claves semánticas en Valkey.
* Organizar métricas por categoría.

---

### 4. Inicialización del Consumer

```go
rdb := redis.NewClient(&redis.Options{
	Addr: "valkey.backend.svc.cluster.local:6379",
})
```

* Crea la conexión con Valkey.
* Se utiliza como almacenamiento en memoria para métricas.

```go
brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
topic := "ventas-blackfriday"
```

* Obtiene dinámicamente los brokers Kafka.
* Define el tópico a consumir.

---

### 5. Consumo de Particiones Kafka

```go
partitions, _ := consumer.Partitions(topic)
```

* Kafka divide los mensajes en particiones.
* Se crea un consumidor por partición para **paralelismo**.

```go
go func(pc sarama.PartitionConsumer) { ... }
```

* Cada partición se procesa en una **goroutine** independiente.
* Mejora el rendimiento bajo alta carga.

---

### 6. Procesamiento de Cada Mensaje

```go
json.Unmarshal(msg.Value, &venta)
```

* Convierte el mensaje Kafka (JSON) en la estructura `Venta`.

---

## 7. Métricas Calculadas y Almacenadas en Valkey

### 7.1 Precio Máximo y Mínimo Global

```go
actualizarMinMax(ctx, rdb, venta.Precio)
```

* Mantiene el precio máximo y mínimo global.
* Permite consultas rápidas desde Grafana.

---

### 7.2 Ranking de Productos (Global y por Categoría)

```go
rdb.ZIncrBy(ctx, "ranking:global", ..., venta.ProductoID)
```

* Usa **Sorted Sets** de Redis/Valkey.
* El score representa la cantidad vendida.
* Permite obtener:

  * Producto más vendido
  * Producto menos vendido

---

### 7.3 Precio Promedio y Cantidad Promedio

```go
rdb.HIncrByFloat(ctx, "stats:precio:suma", catName, venta.Precio)
```

* Usa **Hashes** para almacenar acumulados.
* Permite calcular promedios sin recorrer todos los datos.

---

### 7.4 Total de Reportes por Categoría

```go
rdb.Incr(ctx, "reportes:total:"+catName)
```

* Contador simple por categoría.
* Ideal para paneles tipo **Stat** en Grafana.

---

### 7.5 Variación de Precio (Time Series – Electrónica)

```go
rdb.ZAdd(ctx, keyHistorial, redis.Z{ ... })
```

* Usa **Sorted Sets** como series temporales.
* El timestamp se usa como score.
* Permite graficar la evolución del precio en el tiempo.

```go
rdb.ZRemRangeByRank(ctx, keyHistorial, 0, -21)
```

* Mantiene solo los últimos registros.
* Controla el uso de memoria.

---

### 8. Función `actualizarMinMax`

```go
func actualizarMinMax(ctx context.Context, rdb *redis.Client, precio float64)
```

* Encapsula la lógica para mantener precios mínimos y máximos.
* Evita duplicar código.
* Mejora mantenibilidad y claridad.

---

## Rol del Kafka Consumer en la Arquitectura

* Consume eventos desde Kafka.
* Procesa datos de forma asíncrona.
* Calcula métricas agregadas en tiempo real.
* Almacena resultados en Valkey.
* Alimenta dashboards en Grafana.

Este componente es **altamente escalable** y puede aumentar réplicas para mejorar throughput.



## 8. Analisis de los archivos yml



## Análisis del Archivo `ingress.yml` – Kubernetes Ingress

Este archivo define un **recurso Ingress de Kubernetes**, cuyo objetivo es **exponer servicios internos del clúster hacia el exterior**, actuando como **punto de entrada único** para las solicitudes HTTP del sistema.

En esta arquitectura, el Ingress permite que clientes externos (por ejemplo, Locust o usuarios finales) accedan a la **API REST en Rust** sin necesidad de exponer directamente los servicios internos.

El archivo `ingress.yml` cumple un rol fundamental al **exponer de forma controlada la API REST**, permitiendo pruebas de carga y consumo externo sin comprometer la seguridad ni la modularidad del clúster. Su configuración es adecuada para un entorno de pruebas y demuestra el uso correcto de patrones de acceso en Kubernetes.

---

### 1. Información General del Recurso

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
```

* Define un recurso de tipo **Ingress**, estándar en Kubernetes.
* Permite enrutar tráfico HTTP/HTTPS a servicios internos basándose en reglas.

---

### 2. Metadatos

```yaml
metadata:
  name: ingress-proyecto
  namespace: backend
```

* **name**: Identificador del Ingress dentro del clúster.
* **namespace**: `backend`, asegurando que el Ingress opera sobre los servicios del mismo namespace.

---

### 3. Anotaciones del Ingress Controller

```yaml
annotations:
  nginx.ingress.kubernetes.io/ssl-redirect: "false"
  nginx.ingress.kubernetes.io/use-regex: "true"
```

Estas anotaciones configuran el comportamiento del **NGINX Ingress Controller**.

* **ssl-redirect: "false"**

  * Desactiva la redirección automática de HTTP a HTTPS.
  * Útil en entornos de laboratorio o pruebas internas.

* **use-regex: "true"**

  * Permite el uso de expresiones regulares en los paths.
  * Aporta flexibilidad para enrutar múltiples endpoints bajo una misma regla.

---

### 4. Clase de Ingress

```yaml
spec:
  ingressClassName: nginx
```

* Indica que este Ingress será gestionado por el **NGINX Ingress Controller**.
* Es importante cuando existen múltiples controladores en el clúster.

---

### 5. Reglas de Enrutamiento

```yaml
rules:
- http:
    paths:
    - path: /venta
      pathType: Prefix
```

* Define reglas de enrutamiento HTTP.
* **path: /venta**

  * Todas las solicitudes que comiencen con `/venta` serán procesadas por esta regla.
* **pathType: Prefix**

  * Coincide con `/venta` y cualquier subruta (por ejemplo, `/venta/123`).

---

### 6. Backend del Ingress

```yaml
backend:
  service:
    name: rust-api
    port:
      number: 8080
```

* **service.name**: `rust-api`

  * Nombre del Service de Kubernetes que expone la API REST en Rust.
* **port.number**: `8080`

  * Puerto del Service al cual se redirige el tráfico.

Este backend permite que el Ingress:

* Reciba tráfico externo.
* Lo reenvíe internamente al servicio correcto.
* Mantenga desacoplada la infraestructura interna del acceso externo.

---

## Rol del Ingress en la Arquitectura

* Actúa como **gateway HTTP** del sistema.
* Centraliza el acceso externo.
* Evita exponer múltiples servicios directamente.
* Facilita balanceo de carga y escalabilidad.
* Permite integrar herramientas de prueba como Locust.

---

## Flujo del Tráfico

1. Locust o un cliente externo realiza una petición HTTP:

   ```
   POST /venta
   ```
2. El Ingress recibe la solicitud.
3. Aplica la regla de enrutamiento.
4. Redirige el tráfico al Service `rust-api`.
5. La API en Rust procesa la solicitud y continúa el flujo interno.

---



## Análisis del Archivo `.k8s/rust-api/deployment.yml` – API REST en Rust

Este archivo define los recursos necesarios para desplegar la **API REST escrita en Rust** dentro de Kubernetes. Incluye tanto el **Deployment**, encargado de la gestión de pods, como el **Service**, que permite la comunicación interna dentro del clúster.

Este componente representa la **puerta de entrada principal del sistema**, expuesta al exterior mediante el Ingress.

El archivo `./k8s/rust-api/deployment.yml` define correctamente un **microservicio ligero, desacoplado y preparado para Kubernetes**, con una configuración eficiente de recursos y comunicación interna segura. Su integración con Ingress, Service y HPA demuestra el uso adecuado de patrones cloud-native.

---

## 1. Deployment de la API REST en Rust

### 1.1 Información General

```yaml
apiVersion: apps/v1
kind: Deployment
```

* Define un **Deployment**, recurso encargado de:

  * Crear y mantener los pods.
  * Garantizar alta disponibilidad.
  * Permitir escalado horizontal.

---

### 1.2 Metadatos

```yaml
metadata:
  name: rust-api
  namespace: backend
  labels:
    app: rust-api
```

* **name**: Identifica el Deployment.
* **namespace**: `backend`, alineado con el resto de componentes.
* **labels**: Se utilizan para:

  * Selección de pods.
  * Asociación con Services y HPA.

---

### 1.3 Configuración de Réplicas y Selector

```yaml
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rust-api
```

* **replicas: 1**

  * En estado inicial se utiliza una sola réplica.
  * El escalado se gestiona dinámicamente mediante HPA.
* **selector**

  * Asegura que el Deployment administre únicamente los pods con la etiqueta correcta.

---

### 1.4 Template del Pod

```yaml
template:
  metadata:
    labels:
      app: rust-api
```

* Define las etiquetas que heredarán los pods.
* Estas etiquetas son usadas por el Service y el Ingress.

---

### 1.5 Contenedor de la API

```yaml
containers:
- name: rust-api
  image: freedom-initially-advocate-suites.trycloudflare.com/rust-api:latest
```

* **image**:

  * Imagen Docker alojada en un registry expuesto mediante **Cloudflare Tunnel**.
  * Se utiliza HTTPS para permitir el `pull` seguro desde Kubernetes.

Este enfoque resolvió problemas previos con registries expuestos solo por HTTP.

---

### 1.6 Puertos Expuestos

```yaml
ports:
- containerPort: 8080
```

* Puerto interno del contenedor donde escucha la API REST.
* Coincide con el puerto configurado en Actix Web.

---

### 1.7 Variables de Entorno

```yaml
env:
- name: GO_SERVICE_URL
  value: "http://go-api:8081"
```

* Permite desacoplar la configuración del código.
* Define la URL del servicio Go (gRPC Client).
* Utiliza el **DNS interno de Kubernetes** (`go-api`) para la comunicación entre servicios.

---

### 1.8 Recursos (Requests y Limits)

```yaml
resources:
  requests:
    cpu: "10m"
    memory: "32Mi"
  limits:
    cpu: "100m"
    memory: "64Mi"
```

#### Requests

* **CPU: 10m**
* **Memoria: 32Mi**

Indican los recursos mínimos requeridos para ejecutar el contenedor.
Valores bajos permiten:

* Mayor densidad de pods.
* Uso eficiente del clúster.

#### Limits

* **CPU: 100m**
* **Memoria: 64Mi**

Evitan que el contenedor consuma recursos en exceso, protegiendo la estabilidad del clúster.

> Estos valores son adecuados para una API ligera y stateless, y funcionan correctamente con HPA.

---

## 2. Service de la API REST en Rust

### 2.1 Definición del Service

```yaml
apiVersion: v1
kind: Service
```

* Define un **Service de Kubernetes**.
* Permite exponer el Deployment internamente.

---

### 2.2 Metadatos

```yaml
metadata:
  name: rust-api
  namespace: backend
```

* El nombre coincide con el Deployment.
* Facilita la resolución DNS interna (`rust-api.backend.svc.cluster.local`).

---

### 2.3 Selector y Puertos

```yaml
spec:
  selector:
    app: rust-api
```

* Asocia el Service con los pods del Deployment.

```yaml
ports:
  - protocol: TCP
    port: 8080
    targetPort: 8080
```

* **port**: Puerto expuesto por el Service.
* **targetPort**: Puerto del contenedor.

---

### 2.4 Tipo de Service

```yaml
type: ClusterIP
```

* Expone el servicio **solo dentro del clúster**.
* El acceso externo se realiza exclusivamente a través del Ingress.
* Mejora la seguridad y el control del tráfico.

---

## Rol del `rust-api` en la Arquitectura

* Punto de entrada REST del sistema.
* Recibe tráfico desde el Ingress.
* Reenvía datos al servicio Go.
* Stateless y escalable.
* Optimizado para bajo consumo de recursos.

---



## Análisis del Archivo `.k8s/go-api/deployment.yml` – API Go (gRPC Client)

Este archivo define el despliegue de la **API en Go que actúa como cliente gRPC** dentro del clúster de Kubernetes.

Su función principal es **recibir solicitudes HTTP**, transformarlas al contrato Protobuf y **comunicarse con el gRPC Server**, integrándose así al flujo interno del sistema distribuido.

El archivo `.k8s/go-api/deployment.yml` implementa correctamente una **API intermedia ligera, desacoplada y preparada para Kubernetes**, con una configuración eficiente de recursos y comunicación interna segura. Este servicio es clave para integrar la capa REST con el backend gRPC y Kafka.

---

## 1. Deployment de la API Go

### 1.1 Información General

```yaml
apiVersion: apps/v1
kind: Deployment
```

* Define un recurso **Deployment**, responsable de:

  * Gestionar el ciclo de vida de los pods.
  * Mantener el número deseado de réplicas.
  * Permitir escalabilidad horizontal.

---

### 1.2 Metadatos

```yaml
metadata:
  name: go-api
  namespace: backend
  labels:
    app: go-api
```

* **name**: Identifica el Deployment.
* **namespace**: `backend`, consistente con el resto del sistema.
* **labels**: Utilizadas para selección de pods y asociación con Services y HPA.

---

### 1.3 Réplicas y Selector

```yaml
spec:
  replicas: 1
  selector:
    matchLabels:
      app: go-api
```

* **replicas: 1**

  * Configuración inicial mínima.
  * El escalado se controla mediante HPA.
* **selector**

  * Garantiza que el Deployment gestione únicamente los pods correctos.

---

### 1.4 Template del Pod

```yaml
template:
  metadata:
    labels:
      app: go-api
```

* Las etiquetas del pod permiten que el Service enrute correctamente el tráfico.

---

### 1.5 Contenedor `go-api`

```yaml
containers:
- name: go-api
  image: freedom-initially-advocate-suites.trycloudflare.com/go-api:latest
  imagePullPolicy: Always
```

* **image**

  * Imagen Docker almacenada en un registry privado expuesto mediante **Cloudflare Tunnel**.
  * Permite descargas seguras vía HTTPS.
* **imagePullPolicy: Always**

  * Garantiza que Kubernetes siempre obtenga la versión más reciente de la imagen.

---

### 1.6 Puertos Expuestos

```yaml
ports:
- containerPort: 8081
```

* Puerto donde escucha la API HTTP en Go.
* Coincide con la configuración del servidor HTTP en el código.

---

### 1.7 Variables de Entorno

```yaml
env:
- name: GRPC_SERVER_ADDR
  value: "go-server:50051"
```

* Define la dirección del **gRPC Server**.
* Utiliza DNS interno de Kubernetes.
* Permite desacoplar configuración del código fuente.

---

### 1.8 Recursos (Requests y Limits)

```yaml
resources:
  requests:
    cpu: "10m"
    memory: "32Mi"
  limits:
    cpu: "100m"
    memory: "64Mi"
```

* **Requests**

  * Recursos mínimos requeridos.
  * Permiten una alta densidad de pods.
* **Limits**

  * Evitan consumo excesivo.
  * Protegen la estabilidad del clúster.

Estos valores son adecuados para una API ligera y stateless.

---

## 2. Service de la API Go

### 2.1 Definición del Service

```yaml
apiVersion: v1
kind: Service
```

* Define un **Service de tipo ClusterIP**.
* Permite exponer la API Go internamente.

---

### 2.2 Metadatos

```yaml
metadata:
  name: go-api
  namespace: backend
```

* Permite resolución DNS interna:

  ```
  go-api.backend.svc.cluster.local
  ```

---

### 2.3 Selector y Puertos

```yaml
spec:
  selector:
    app: go-api
```

* Asocia el Service con los pods del Deployment.

```yaml
ports:
- protocol: TCP
  port: 8081
  targetPort: 8081
```

* Define el puerto del Service y el puerto interno del contenedor.

---

### 2.4 Tipo de Service

```yaml
type: ClusterIP
```

* El servicio solo es accesible dentro del clúster.
* El acceso externo se maneja a través del Ingress (si se requiere).

---

## Rol del `go-api` en la Arquitectura

* Actúa como **adaptador REST → gRPC**.
* Convierte solicitudes HTTP a llamadas gRPC.
* Reduce la latencia interna usando Protobuf.
* Permite comparar REST y gRPC.
* Stateless y escalable horizontalmente.

---



## Análisis del Archivo `.k8s/go-server/deployment.yml` – gRPC Server en Go

Este archivo define el despliegue del **servidor gRPC** del sistema, el cual cumple un rol central al **recibir solicitudes desde los clientes gRPC**, procesarlas y **publicar eventos en Kafka**.

Este componente actúa como el **núcleo de la lógica de negocio**, desacoplando las APIs de entrada del sistema de mensajería.

El archivo `.k8s/go-server/deployment.yml` define correctamente un **servidor gRPC centralizado, eficiente y preparado para Kubernetes**, con configuración de recursos optimizada y comunicación interna segura. Su diseño facilita pruebas de escalabilidad y análisis de rendimiento del sistema.

---

## 1. Deployment del gRPC Server

### 1.1 Información General

```yaml
apiVersion: apps/v1
kind: Deployment
```

* Define un **Deployment**, encargado de:

  * Gestionar el ciclo de vida de los pods.
  * Mantener el número deseado de réplicas.
  * Permitir escalabilidad horizontal mediante HPA.

---

### 1.2 Metadatos

```yaml
metadata:
  name: go-server
  namespace: backend
  labels:
    app: go-server
```

* **name**: Identifica el Deployment.
* **namespace**: `backend`, consistente con el resto del sistema.
* **labels**: Utilizadas para selección de pods, Services y HPA.

---

### 1.3 Réplicas y Selector

```yaml
spec:
  replicas: 1   # luego pruebas con 2 (obligatorio en el proyecto)
  selector:
    matchLabels:
      app: go-server
```

* **replicas: 1**

  * Configuración inicial mínima.
  * El proyecto contempla pruebas con múltiples réplicas para evaluar rendimiento.
* **selector**

  * Asegura que el Deployment administre únicamente los pods correctos.

---

### 1.4 Template del Pod

```yaml
template:
  metadata:
    labels:
      app: go-server
```

* Permite que el Service y otros recursos identifiquen los pods del gRPC Server.

---

### 1.5 Contenedor `go-server`

```yaml
containers:
- name: go-server
  image: freedom-initially-advocate-suites.trycloudflare.com/go-server
  imagePullPolicy: Always
```

* **image**

  * Imagen Docker alojada en un registry privado expuesto mediante **Cloudflare Tunnel**.
  * El uso de HTTPS garantiza descargas seguras desde Kubernetes.
* **imagePullPolicy: Always**

  * Asegura que siempre se use la versión más reciente del servidor gRPC.

---

### 1.6 Puertos Expuestos

```yaml
ports:
- containerPort: 50051
```

* Puerto estándar utilizado por gRPC.
* Coincide con el listener configurado en el código Go.

---

### 1.7 Variables de Entorno

```yaml
env:
- name: KAFKA_BROKERS
  value: "blackfriday-kafka-bootstrap.kafka:9092"
```

* Define la dirección del clúster Kafka.
* Utiliza el **Service interno creado por Strimzi**.
* Permite desacoplar la configuración del código.

---

### 1.8 Recursos (Requests y Limits)

```yaml
resources:
  requests:
    cpu: "10m"
    memory: "32Mi"
  limits:
    cpu: "100m"
    memory: "64Mi"
```

* **Requests**

  * Recursos mínimos necesarios para el funcionamiento.
  * Permiten una alta densidad de pods.
* **Limits**

  * Previenen consumo excesivo.
  * Protegen la estabilidad del clúster.

Estos valores son adecuados para un servicio gRPC ligero y stateless.

---

## 2. Service del gRPC Server

### 2.1 Definición del Service

```yaml
apiVersion: v1
kind: Service
```

* Define un **Service de tipo ClusterIP**.
* Expone el servidor gRPC internamente dentro del clúster.

---

### 2.2 Metadatos

```yaml
metadata:
  name: go-server
  namespace: backend
```

* Permite resolución DNS interna:

  ```
  go-server.backend.svc.cluster.local
  ```

---

### 2.3 Selector y Puertos

```yaml
spec:
  selector:
    app: go-server
```

* Asocia el Service con los pods del Deployment.

```yaml
ports:
- port: 50051
  targetPort: 50051
```

* Define el puerto del Service y el puerto del contenedor.

---

### 2.4 Tipo de Service

```yaml
type: ClusterIP
```

* El servicio es accesible únicamente dentro del clúster.
* El acceso externo se realiza a través de otros componentes (por ejemplo, API Go).

---

## Rol del `go-server` en la Arquitectura

* Recibe solicitudes gRPC desde los clientes.
* Centraliza la lógica de negocio.
* Publica eventos en Kafka.
* Permite desacoplamiento entre productores y consumidores.
* Escalable horizontalmente para manejar alta concurrencia.

---




## Análisis del Archivo `.k8s/go-consumer/deployment.yml` – Kafka Consumer en Go

Este archivo define el despliegue del **consumidor de Kafka** escrito en Go.
Su función principal es **leer eventos de ventas desde Kafka**, procesarlos y **almacenarlos en Valkey (Redis)** para su posterior visualización en Grafana.

Este componente cierra el flujo completo del sistema, transformando eventos en **métricas y series de tiempo**.

El archivo `.k8s/go-consumer/deployment.yml` define un **consumidor Kafka robusto, eficiente y bien integrado con Kubernetes**, preparado para manejar flujos de eventos en tiempo real y alimentar sistemas de monitoreo como Grafana.

Su diseño facilita pruebas de carga, análisis de rendimiento y escalabilidad del sistema completo.


---

## 1. Deployment del Kafka Consumer

### 1.1 Tipo de Recurso

```yaml
apiVersion: apps/v1
kind: Deployment
```

* Utiliza un **Deployment** para:

  * Garantizar la ejecución continua del consumidor.
  * Permitir escalabilidad horizontal (replicas).
  * Facilitar reinicios automáticos ante fallos.

---

### 1.2 Metadatos

```yaml
metadata:
  name: kafka-consumer
  namespace: backend
  labels:
    app: kafka-consumer
```

* **name**: Identifica al consumidor de Kafka.
* **namespace**: `backend`, manteniendo coherencia con el sistema.
* **labels**: Clave para monitoreo, selección y escalamiento.

---

### 1.3 Réplicas y Selector

```yaml
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kafka-consumer
```

* **replicas: 1**

  * Configuración inicial.
  * En Kafka, el paralelismo depende del número de **particiones** del tópico.
* **selector**

  * Asegura que el Deployment controle únicamente sus pods.

> 📌 Nota: Para escalar este consumidor, es necesario que el tópico Kafka tenga múltiples particiones; de lo contrario, las réplicas quedarían ociosas.

---

## 2. Pod Template

### 2.1 Labels del Pod

```yaml
template:
  metadata:
    labels:
      app: kafka-consumer
```

* Permite identificar el pod en logs, métricas y debugging.

---

### 2.2 Contenedor `kafka-consumer`

```yaml
containers:
- name: kafka-consumer
  image: guys-dip-sessions-venture.trycloudflare.com/kafka-consumer:latest
```

* Imagen Docker alojada en un **registry privado** accesible mediante Cloudflare Tunnel.
* Contiene el consumidor Kafka escrito en Go.
* `latest` facilita pruebas rápidas durante el desarrollo.

---

### 2.3 Variables de Entorno

```yaml
env:
- name: KAFKA_BROKERS
  value: "blackfriday-kafka-bootstrap.kafka:9092"
```

* Define el endpoint del clúster Kafka gestionado por **Strimzi**.
* Permite cambiar brokers sin modificar el código.
* Usa resolución DNS interna de Kubernetes.

---

### 2.4 Recursos (CPU y Memoria)

```yaml
resources:
  requests:
    cpu: "10m"
    memory: "32Mi"
  limits:
    cpu: "100m"
    memory: "64Mi"
```

* **Requests**

  * Recursos mínimos garantizados.
  * Adecuado para procesamiento ligero de mensajes.
* **Limits**

  * Previenen consumo excesivo ante picos de carga.
* Esta configuración permite **alta eficiencia** y bajo costo.

---

## 3. Rol del Kafka Consumer en la Arquitectura

* Consume eventos desde Kafka (`ventas-blackfriday`).
* Procesa los mensajes (agregaciones, estadísticas, rankings).
* Persiste métricas en **Valkey**:

  * Precios máximos y mínimos.
  * Rankings de productos.
  * Series de tiempo para electrónica.
* Actúa como puente entre **mensajería** y **visualización**.

---

## 4. Consideraciones de Escalabilidad

* El consumidor es **state-less**, ideal para escalar.
* Para aumentar throughput:

  * Incrementar el número de particiones en Kafka.
  * Aumentar el número de réplicas del Deployment.
* Puede integrarse con un **HPA basado en CPU** o métricas personalizadas.

---


## Análisis del Archivo `kafka-cluster.yml` – Clúster Kafka con Strimzi

Este archivo define la **infraestructura de Kafka** dentro de Kubernetes utilizando **Strimzi**, el operador oficial para ejecutar Apache Kafka de forma nativa en clústeres Kubernetes.

La configuración está optimizada para **entornos académicos y de laboratorio**, priorizando simplicidad, bajo consumo de recursos y facilidad de despliegue.

El archivo `kafka-cluster.yml` implementa un **clúster Kafka moderno, ligero y completamente integrado con Kubernetes**, utilizando Strimzi en modo KRaft.
Es una solución ideal para pruebas de arquitectura distribuida, análisis de rendimiento y flujos de datos en tiempo real.

---

## 1. KafkaNodePool – Definición de Nodos Kafka

### 1.1 Tipo de Recurso

```yaml
apiVersion: kafka.strimzi.io/v1
kind: KafkaNodePool
```

* Introducido en versiones recientes de Strimzi.
* Permite separar y controlar los **roles de los nodos Kafka** (controller y broker).
* Reemplaza el modelo monolítico clásico de Kafka.

---

### 1.2 Metadatos

```yaml
metadata:
  name: kafka-nodes
  namespace: kafka
  labels:
    strimzi.io/cluster: blackfriday
```

* **name**: Identificador del pool de nodos.
* **namespace**: `kafka`, dedicado exclusivamente a la mensajería.
* **label `strimzi.io/cluster`**: Asocia el NodePool con el clúster Kafka llamado `blackfriday`.

---

### 1.3 Especificación del NodePool

```yaml
spec:
  replicas: 1
```

* Se crea **un único nodo Kafka**.
* Adecuado para pruebas funcionales y desarrollo.
* No ofrece tolerancia a fallos (single point of failure).

---

### 1.4 Roles del Nodo

```yaml
roles:
  - controller
  - broker
```

* **controller**:

  * Gestiona el metadata del clúster.
  * Reemplaza a ZooKeeper (modo KRaft).
* **broker**:

  * Almacena y distribuye los mensajes.
* Ambos roles están combinados en el mismo nodo para simplificar la arquitectura.

---

### 1.5 Almacenamiento

```yaml
storage:
  type: ephemeral
```

* Usa almacenamiento **temporal**.
* Los datos se pierden si el pod se reinicia.
* Ideal para pruebas y simulaciones.
* No recomendado para producción.

---

## 2. Recurso Kafka – Definición del Clúster

### 2.1 Tipo de Recurso

```yaml
kind: Kafka
```

* Recurso principal gestionado por Strimzi.
* Orquesta brokers, listeners, configuración y operadores.

---

### 2.2 Metadatos y Anotaciones

```yaml
metadata:
  name: blackfriday
  namespace: kafka
  annotations:
    strimzi.io/node-pools: enabled
    strimzi.io/kraft: enabled
```

* **node-pools: enabled**
  Indica que el clúster usará `KafkaNodePool`.
* **kraft: enabled**
  Activa el modo **KRaft**, eliminando la dependencia de ZooKeeper.

---

## 3. Configuración del Broker Kafka

### 3.1 Versión de Kafka

```yaml
kafka:
  version: 4.0.0
  metadataVersion: 4.0-IV1
```

* Usa una versión moderna de Kafka.
* Compatible con KRaft.
* Asegura mejoras de rendimiento y estabilidad.

---

### 3.2 Listeners

```yaml
listeners:
  - name: plain
    port: 9092
    type: internal
    tls: false
```

* **Listener interno**

  * Solo accesible dentro del clúster Kubernetes.
* **Sin TLS**

  * Simplifica la conexión para entornos de laboratorio.
* Puerto estándar `9092`.

---

### 3.3 Configuración de Replicación

```yaml
config:
  offsets.topic.replication.factor: 1
  transaction.state.log.replication.factor: 1
  transaction.state.log.min.isr: 1
  default.replication.factor: 1
  min.insync.replicas: 1
```

* Todos los factores de replicación están en **1**:

  * Requerido debido a que solo existe un broker.
* Garantiza funcionamiento sin errores por ISR insuficiente.
* No ofrece alta disponibilidad, pero sí estabilidad para pruebas.

---

## 4. Entity Operator

```yaml
entityOperator:
  topicOperator: {}
  userOperator: {}
```

* **Topic Operator**

  * Permite crear y gestionar tópicos vía CRDs.
* **User Operator**

  * Maneja usuarios y credenciales Kafka.
* Facilita la administración declarativa de Kafka en Kubernetes.

---

## 5. Rol de Kafka en la Arquitectura del Sistema

Kafka actúa como el **núcleo de desacoplamiento** del sistema:

* Recibe eventos desde el **gRPC Server**.
* Permite procesamiento asíncrono.
* Facilita escalabilidad independiente:

  * Productores (APIs).
  * Consumidores (analytics).
* Asegura resiliencia frente a picos de tráfico.

---

## 6. Limitaciones de la Configuración

* Un solo nodo (no tolerante a fallos).
* Almacenamiento efímero.
* Sin seguridad TLS o autenticación.
* Diseñada exclusivamente para fines académicos y demostrativos.

---



## Análisis del Archivo `.k8s/valkey/deployment.yml` – Despliegue de Valkey en Kubernetes

Este archivo define el despliegue de **Valkey**, un fork moderno y de alto rendimiento de Redis, utilizado en el sistema como **almacenamiento en memoria** para métricas, contadores, rankings y series temporales que luego son consumidas por **Grafana**.

Valkey funciona como la **capa de persistencia rápida** del sistema.

El archivo `.k8s/valkey/deployment.yml` implementa una **base de datos en memoria rápida y eficiente**, perfectamente integrada con Kafka y Grafana.
Su configuración está alineada con los objetivos del proyecto: **análisis en tiempo real, bajo consumo de recursos y simplicidad operativa**.

---

## 1. Deployment de Valkey

### 1.1 Tipo de Recurso

```yaml
apiVersion: apps/v1
kind: Deployment
```

* Se utiliza un **Deployment** porque:

  * Permite reinicios automáticos del pod.
  * Facilita escalar réplicas en el futuro.
  * Es suficiente para una base de datos en entorno académico.

---

### 1.2 Metadatos

```yaml
metadata:
  name: valkey
  namespace: backend
```

* **name**: Nombre del Deployment (`valkey`).
* **namespace**: `backend`, donde residen los servicios internos del sistema.
* Mantiene a Valkey aislado del tráfico externo.

---

### 1.3 Réplicas

```yaml
spec:
  replicas: 1
```

* Se ejecuta **una sola réplica**.
* Adecuado para:

  * Pruebas funcionales.
  * Métricas en tiempo real.
* No existe replicación ni alta disponibilidad, lo cual es aceptable para este proyecto.

---

### 1.4 Selector y Labels

```yaml
selector:
  matchLabels:
    app: valkey
```

```yaml
template:
  metadata:
    labels:
      app: valkey
```

* Garantiza que:

  * El Deployment gestione únicamente los pods con `app: valkey`.
  * El Service pueda enrutar correctamente el tráfico hacia Valkey.

---

## 2. Contenedor Valkey

### 2.1 Imagen

```yaml
image: valkey/valkey:7.2
```

* Imagen oficial de Valkey.
* Basada en Redis 7.x, compatible con:

  * `GET`, `SET`, `INCR`
  * `ZADD`, `ZINCRBY`
  * `HINCRBY`, `HINCRBYFLOAT`
* Ideal para métricas, rankings y time series simples.

---

### 2.2 Puerto Expuesto

```yaml
ports:
  - containerPort: 6379
```

* Puerto estándar de Redis/Valkey.
* Usado por:

  * Kafka Consumer (Go)
  * Grafana Redis Datasource

---

### 2.3 Argumentos de Ejecución

```yaml
args:
  - "--bind"
  - "0.0.0.0"
  - "--protected-mode"
  - "no"
```

#### Explicación:

* `--bind 0.0.0.0`

  * Permite conexiones desde otros pods del clúster.
* `--protected-mode no`

  * Desactiva el modo protegido.
  * Necesario para permitir conexiones sin autenticación dentro del clúster.

⚠️ **Nota**:
Esta configuración **NO es segura para producción**, pero es válida en un entorno controlado de Kubernetes académico.

---

## 3. Service de Valkey

### 3.1 Tipo de Recurso

```yaml
kind: Service
```

* Permite exponer Valkey internamente dentro del clúster.

---

### 3.2 Nombre y Namespace

```yaml
metadata:
  name: valkey
  namespace: backend
```

* El DNS generado es:

```
valkey.backend.svc.cluster.local
```

* Utilizado por:

  * Kafka Consumer
  * Grafana
  * Pruebas con `redis-cli`

---

### 3.3 Selector

```yaml
selector:
  app: valkey
```

* Enruta el tráfico al pod correcto del Deployment.

---

### 3.4 Puertos

```yaml
ports:
  - port: 6379
    targetPort: 6379
```

* **port**: Puerto expuesto por el Service.
* **targetPort**: Puerto interno del contenedor Valkey.

---

### 3.5 Tipo de Service

```yaml
type: ClusterIP
```

* Solo accesible **dentro del clúster**.
* No expone Valkey a internet.
* Aumenta la seguridad del sistema.

---

## 4. Rol de Valkey en la Arquitectura

Valkey cumple un papel crítico en el sistema:

* 📊 Almacena métricas agregadas:

  * Precio máximo y mínimo.
  * Total de reportes por categoría.
* 🏆 Rankings:

  * Productos más vendidos (ZSET).
* ⏱ Series temporales:

  * Variación de precios en electrónica.
* 📈 Fuente de datos para Grafana.

Kafka desacopla → Valkey persiste → Grafana visualiza.

---

## 5. Limitaciones del Diseño

* Sin persistencia en disco.
* Sin autenticación.
* Una sola réplica.
* Sin clustering nativo de Valkey.

Estas decisiones fueron tomadas para:

* Reducir complejidad.
* Facilitar despliegue.
* Enfocarse en el flujo distribuido completo.

---



## Análisis del Archivo `.k8s/grafana/deployment.yml` – Visualización y Monitoreo con Grafana

El archivo `.k8s/grafana/deployment.yml` define el despliegue de **Grafana** dentro del clúster de Kubernetes.
Grafana es la **capa de visualización** del sistema y se utiliza para mostrar métricas procesadas desde **Valkey**, las cuales provienen indirectamente del flujo Kafka → Consumer → Valkey.

El archivo `.k8s/grafana/deployment.yml` implementa correctamente la **capa de visualización del sistema**, integrándose de forma directa con Valkey mediante el plugin Redis Datasource.
Su configuración es ligera, eficiente y suficiente para demostrar métricas en tiempo real generadas por Kafka y procesadas por consumidores en Go.

---

## 1. Deployment de Grafana

### 1.1 Tipo de Recurso

```yaml
apiVersion: apps/v1
kind: Deployment
```

* Se utiliza un **Deployment** porque:

  * Grafana es un servicio **sin estado crítico** (stateless para este proyecto).
  * Permite reinicios automáticos del pod.
  * Facilita escalar en caso de mayor carga de usuarios.

---

### 1.2 Metadatos

```yaml
metadata:
  name: grafana
  namespace: backend
```

* **name**: `grafana`, nombre del Deployment.
* **namespace**: `backend`, coherente con el resto de servicios internos.
* Mantiene a Grafana separado de componentes de infraestructura como Kafka o Ingress.

---

### 1.3 Réplicas

```yaml
spec:
  replicas: 1
```

* Se ejecuta **una sola réplica**:

  * Suficiente para visualización académica.
  * Evita conflictos de sesión o sincronización de dashboards.
* En producción podrían utilizarse múltiples réplicas con almacenamiento persistente.

---

### 1.4 Selector y Labels

```yaml
selector:
  matchLabels:
    app: grafana
```

```yaml
template:
  metadata:
    labels:
      app: grafana
```

* Garantiza que:

  * El Deployment controle únicamente pods de Grafana.
  * El Service pueda enrutar tráfico correctamente hacia el pod.

---

## 2. Contenedor Grafana

### 2.1 Imagen

```yaml
image: grafana/grafana:11.0.0
```

* Imagen oficial de Grafana.
* Versión moderna con:

  * Soporte estable para plugins.
  * Mejoras de rendimiento.
  * Compatibilidad con Redis/Valkey datasource.

---

### 2.2 Puerto Expuesto

```yaml
ports:
  - containerPort: 3000
```

* Puerto por defecto de Grafana.
* Usado para:

  * Acceso web al dashboard.
  * Configuración de datasources y paneles.

---

### 2.3 Instalación Automática de Plugins

```yaml
env:
  - name: GF_INSTALL_PLUGINS
    value: redis-datasource
```

#### Explicación:

* `GF_INSTALL_PLUGINS` permite instalar plugins al iniciar el contenedor.
* `redis-datasource`:

  * Habilita a Grafana para conectarse a Redis/Valkey.
  * Permite consultas tipo:

    * `GET ventas:global:precio_max`
    * `ZREVRANGE ranking:global 0 5 WITHSCORES`
* Evita instalación manual dentro del pod.

Este paso es **clave** para integrar Grafana con Valkey.

---

## 3. Service de Grafana

### 3.1 Tipo de Recurso

```yaml
kind: Service
```

* Permite exponer Grafana dentro del clúster.

---

### 3.2 Nombre y Namespace

```yaml
metadata:
  name: grafana
  namespace: backend
```

* El DNS interno generado es:

```
grafana.backend.svc.cluster.local
```

---

### 3.3 Selector

```yaml
selector:
  app: grafana
```

* Conecta el Service con el pod correcto del Deployment.

---

### 3.4 Puertos

```yaml
ports:
  - port: 3000
    targetPort: 3000
```

* **port**: Puerto del Service.
* **targetPort**: Puerto del contenedor Grafana.

---

### 3.5 Tipo de Service

```yaml
type: ClusterIP
```

* Grafana solo es accesible **desde dentro del clúster**.
* El acceso externo normalmente se realiza mediante:

  * `kubectl port-forward`, o
  * Ingress (si se desea exponer).

Esto mejora la seguridad del sistema.

---

## 4. Rol de Grafana en la Arquitectura

Grafana representa la **última fase del pipeline**:

1. Locust genera tráfico.
2. Rust API / Go API reciben solicitudes.
3. gRPC Server envía eventos a Kafka.
4. Kafka Consumer procesa y agrega métricas.
5. Valkey almacena datos procesados.
6. **Grafana consulta Valkey y visualiza resultados**.



---



## Análisis del Archivo `.k8s/hpa/deployment.yml` – Autoescalamiento Horizontal (HPA)

El archivo `.k8s/hpa/deployment.yml` define **Horizontal Pod Autoscalers (HPA)** para los principales componentes del sistema.

El HPA permite que Kubernetes **ajuste automáticamente el número de réplicas** de un Deployment en función del uso de recursos, en este caso **CPU**.

Este mecanismo es fundamental para:

* Soportar picos de carga generados por Locust.
* Mantener estabilidad del sistema.
* Comparar rendimiento con y sin escalamiento (requisito del proyecto).


El archivo `.k8s/hpa/deployment.yml` implementa una **estrategia de autoescalamiento efectiva y bien balanceada**, adaptada a la naturaleza de cada componente.
Permite demostrar claramente los beneficios de Kubernetes frente a arquitecturas monolíticas, cumpliendo con los objetivos de rendimiento y resiliencia del proyecto.

---

## 1. ¿Qué es un HPA?

Un **Horizontal Pod Autoscaler**:

* Monitorea métricas (CPU, memoria u otras).
* Aumenta o reduce el número de pods automáticamente.
* Funciona sobre Deployments, StatefulSets o ReplicaSets.

En este proyecto:

* Se usa **CPU Utilization**.
* Requiere que cada contenedor tenga definidos `requests.cpu`.

---

## 2. HPA para `rust-api`

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: rust-api-hpa
  namespace: backend
```

### Objetivo

```yaml
scaleTargetRef:
  apiVersion: apps/v1
  kind: Deployment
  name: rust-api
```

* Aplica directamente al Deployment `rust-api`.
* Escala la API REST escrita en Rust, que es el **primer punto de entrada** del sistema.

### Límites de escalamiento

```yaml
minReplicas: 1
maxReplicas: 3
```

* Mínimo: 1 pod (siempre disponible).
* Máximo: 3 pods.
* Adecuado para una API liviana y altamente eficiente como Rust.

### Métrica usada

```yaml
metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 65
```

* Si el uso promedio de CPU supera el **65%**, Kubernetes crea nuevos pods.
* Rust maneja bien la concurrencia, por eso el umbral es moderado.

---

## 3. HPA para `go-server` (gRPC Server)

```yaml
metadata:
  name: go-server-hpa
```

### Función del servicio

* Recibe llamadas gRPC.
* Publica mensajes a Kafka.
* Es un componente **crítico del pipeline**.

### Configuración

```yaml
minReplicas: 1
maxReplicas: 3
```

* Se permite mayor escalamiento que Rust.
* gRPC Server puede manejar múltiples conexiones concurrentes.

### Umbral de CPU

```yaml
averageUtilization: 70
```

* Se tolera mayor uso de CPU antes de escalar.
* Kafka I/O suele ser más bloqueante que computacional.

---

## 4. HPA para `go-api` (gRPC Client)

```yaml
metadata:
  name: go-api-hpa
```

### Rol del servicio

* Actúa como **puente REST → gRPC**.
* Convierte peticiones HTTP en llamadas gRPC.

### Configuración

```yaml
minReplicas: 1
maxReplicas: 3
averageUtilization: 65
```

* Escala rápidamente cuando aumentan peticiones HTTP.
* Similar comportamiento a `rust-api`, pero con más lógica interna.

---

## 5. HPA para `kafka-consumer`

```yaml
metadata:
  name: kafka-consumer-hpa
```

### Función del consumidor

* Consume mensajes desde Kafka.
* Procesa datos.
* Escribe métricas agregadas en Valkey.

### Configuración de escalamiento

```yaml
minReplicas: 1
maxReplicas: 3
```

* Es el componente que **más escala**.
* Más réplicas permiten mayor throughput de consumo.

### Métrica

```yaml
averageUtilization: 70
```

* El consumo de Kafka es intensivo en CPU y red.
* Se permite mayor utilización antes de escalar.

---

## 6. Requisitos para que el HPA funcione

Para que estos HPA funcionen correctamente, es indispensable:

1. **Metrics Server instalado**:

```bash
kubectl top pods
```

2. **Requests de CPU definidos**, por ejemplo:

```yaml
resources:
  requests:
    cpu: "10m"
```

Sin esto, el HPA **no puede calcular porcentajes**.

---

## 7. Impacto en el Rendimiento (Análisis del Proyecto)

| Componente     | Sin HPA                | Con HPA              |
| -------------- | ---------------------- | -------------------- |
| Rust API       | Latencia alta en picos | Latencia estable     |
| Go API         | Saturación rápida      | Escala dinámicamente |
| gRPC Server    | Cuellos de botella     | Mayor throughput     |
| Kafka Consumer | Lag en Kafka           | Consumo paralelo     |
| Valkey         | Lecturas estables      | Mejor distribución   |

El HPA permitió:

* Reducir errores HTTP 500.
* Disminuir latencia promedio.
* Mejorar el procesamiento de eventos Kafka.
* Visualizar claramente el impacto en Grafana.

---











































## 9. Conclusiones

* El sistema demostró ser **escalable, resiliente y eficiente**.
* Kafka fue clave para manejar cargas elevadas.
* gRPC superó a REST en rendimiento.
* Valkey permitió consultas en tiempo real para Grafana.
* Kubernetes facilitó la orquestación, escalado y observabilidad.

Este proyecto evidencia la aplicación práctica de **arquitecturas distribuidas modernas**, alineadas con buenas prácticas de ingeniería de software y cloud computing.



