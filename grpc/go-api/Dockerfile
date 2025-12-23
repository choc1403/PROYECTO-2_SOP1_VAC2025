# ---------- BUILD STAGE ----------
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copiar proto (para replace grpc/proto)
COPY proto ./proto

# Copiar go.mod y go.sum
COPY go-api/go.mod go-api/go.sum ./go-api/
WORKDIR /app/go-api

RUN go mod download

# Copiar código fuente
COPY go-api/ .

# Compilar binario
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o go-api

# ---------- RUNTIME STAGE ----------
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/go-api/go-api .

EXPOSE 8081
CMD ["./go-api"]
