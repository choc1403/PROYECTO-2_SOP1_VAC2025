# ============================================
# ETAPA 1: BUILD
# ============================================
FROM rust:1.83-slim AS builder

WORKDIR /app

# Dependencias necesarias para reqwest + openssl
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    pkg-config \
    libssl-dev \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copiar archivos de dependencias
COPY Cargo.toml Cargo.lock ./

# Compilar dependencias (cache)
RUN mkdir -p src && \
    echo "fn main() {}" > src/main.rs && \
    cargo build --release 2>&1 || true && \
    rm -rf src

# Copiar código real
COPY . .

# Compilar aplicación real
RUN cargo build --release

# ============================================
# ETAPA 2: RUNTIME
# ============================================
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    libssl3 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# ⚠️ Cambia el nombre del binario si es distinto
COPY --from=builder /app/target/release/rust-api /app/rust-api

# Usuario no root (buena práctica)
RUN useradd -m -u 1001 appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./rust-api"]
