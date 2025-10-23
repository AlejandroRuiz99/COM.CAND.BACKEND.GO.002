# Docker - Guía de Uso

## 🚀 Comandos Básicos

### Levantar el Sistema

```bash
# Levantar NATS y el servidor IoT
docker-compose up -d

# Ver los logs
docker-compose logs -f iot-server
```

### Detener el Sistema

```bash
# Detener servicios
docker-compose down

# Detener y eliminar volúmenes (limpieza completa)
docker-compose down -v
```

## 🧪 Ejecutar Tests de Integración

```bash
# Ejecutar tests (el sistema debe estar levantado)
docker-compose --profile test run --rm iot-tests
```

Los tests verifican:
- ✅ Listar sensores iniciales
- ✅ Registrar nuevo sensor
- ✅ Verificar sensor en la lista
- ✅ Obtener configuración
- ✅ Actualizar configuración
- ✅ Verificar cambios reflejados
- ✅ Consultar lecturas
- ✅ Verificar alertas

## 🖥️ Usar el CLI

```bash
# Listar sensores
docker-compose run --rm iot-cli sensor list

# Registrar nuevo sensor
docker-compose run --rm iot-cli sensor register \
  --id my-sensor \
  --type temperature \
  --interval 5000 \
  --threshold 30.0

# Actualizar configuración
docker-compose run --rm iot-cli config set my-sensor \
  --interval 3000 \
  --threshold 35.0

# Obtener configuración
docker-compose run --rm iot-cli config get my-sensor

# Consultar lecturas
docker-compose run --rm iot-cli readings my-sensor

# Modo interactivo
docker-compose run --rm iot-cli interactive
```

## 🔧 Reconstruir Imágenes

```bash
# Reconstruir todas las imágenes
docker-compose build

# Reconstruir sin caché
docker-compose build --no-cache

# Reconstruir solo el servidor
docker-compose build iot-server

# Reconstruir solo los tests
docker-compose --profile test build iot-tests
```

## 📊 Monitoreo

```bash
# Ver estado de los contenedores
docker-compose ps

# Ver logs en tiempo real
docker-compose logs -f

# Ver logs solo del servidor
docker-compose logs -f iot-server

# Ver logs de NATS
docker-compose logs -f nats
```

## 🐛 Troubleshooting

### Limpiar todo y empezar de cero

```bash
docker-compose down -v
docker-compose build --no-cache
docker-compose up -d
```

### Ver base de datos

```bash
# Acceder al contenedor del servidor
docker exec -it iot-server sh

# Dentro del contenedor
ls -la /data/
```

### Verificar conectividad NATS

```bash
# NATS monitor web (abrir en navegador)
http://localhost:8222
```

## 📝 Arquitectura

```
┌─────────────────────────────────────────┐
│           Docker Compose                 │
│                                          │
│  ┌──────────┐  ┌──────────┐            │
│  │   NATS   │◄─┤ IoT      │            │
│  │  :4222   │  │ Server   │            │
│  └──────────┘  └──────────┘            │
│                     │                    │
│                     ▼                    │
│              ┌──────────┐               │
│              │ SQLite   │               │
│              │ Volume   │               │
│              └──────────┘               │
│                                          │
│  Herramientas (profiles):               │
│  ┌──────────┐  ┌──────────┐            │
│  │ IoT CLI  │  │  Tests   │            │
│  └──────────┘  └──────────┘            │
└─────────────────────────────────────────┘
```

## 🎯 Flujo de Trabajo Completo

```bash
# 1. Levantar el sistema
docker-compose up -d

# 2. Esperar a que esté listo (unos segundos)
docker-compose logs -f iot-server

# 3. Ejecutar tests
docker-compose --profile test run --rm iot-tests

# 4. Usar el CLI
docker-compose run --rm iot-cli sensor list

# 5. Detener cuando termines
docker-compose down
```

