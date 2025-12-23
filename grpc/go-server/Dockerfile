# ---------- BUILD STAGE ----------
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY proto ./proto

COPY go-server/go.mod go-server/go.sum ./go-server/
WORKDIR /app/go-server

RUN go mod download

COPY go-server/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o go-server

# ---------- RUNTIME STAGE ----------
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/go-server/go-server .

EXPOSE 50051
CMD ["./go-server"]
