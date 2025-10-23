# Builder stage
FROM golang:1.23-alpine AS builder

# Instalar dependencias de compilación
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copiar código fuente
COPY . .

# Descargar dependencias
RUN go mod download

# Compilar el binario
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o iot-server \
    cmd/iot-server/main.go

# Runtime stage
FROM alpine:3.19

# Instalar certificados CA y zona horaria
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copiar binario compilado
COPY --from=builder /app/iot-server /usr/local/bin/iot-server

# Copiar archivos de configuración
COPY --from=builder /app/configs /app/configs

# Crear directorio para datos
RUN mkdir -p /data

# Exponer puertos (si fuera necesario en el futuro)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep iot-server || exit 1

# Comando por defecto
ENTRYPOINT ["iot-server"]

