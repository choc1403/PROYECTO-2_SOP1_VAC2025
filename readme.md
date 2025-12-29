
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
````

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







## 9. Conclusiones

* El sistema demostró ser **escalable, resiliente y eficiente**.
* Kafka fue clave para manejar cargas elevadas.
* gRPC superó a REST en rendimiento.
* Valkey permitió consultas en tiempo real para Grafana.
* Kubernetes facilitó la orquestación, escalado y observabilidad.

Este proyecto evidencia la aplicación práctica de **arquitecturas distribuidas modernas**, alineadas con buenas prácticas de ingeniería de software y cloud computing.



