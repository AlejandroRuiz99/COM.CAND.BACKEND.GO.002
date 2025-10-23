# Sistema IoT - Gestión de Sensores con Go y NATS

Prueba técnica: Sistema de gestión de sensores IoT con mensajería NATS, worker pool pattern y persistencia SQLite.

## 🚀 Quick Start

```bash
# 1. Levantar el sistema
docker-compose up -d

# 2. Ejecutar tests de integración
docker-compose --profile test run --rm iot-tests

# 3. Usar el CLI
docker-compose run --rm iot-cli sensor list
```

## 📋 Tabla de Contenidos

- [Características](#características)
- [Arquitectura](#arquitectura)
- [Uso](#uso)
- [Tests](#tests)
- [Estructura del Proyecto](#estructura-del-proyecto)
- [Decisiones Técnicas](#decisiones-técnicas)

## ✨ Características

- ✅ **Worker Pool Pattern** - Procesamiento escalable de sensores (5 workers, queue de 100 tareas)
- ✅ **NATS Messaging** - Comunicación pub/sub y request/reply
- ✅ **CLI Completo** - Gestión remota de sensores con Cobra
- ✅ **Persistencia SQLite** - Repository pattern para fácil migración
- ✅ **Logging Estructurado** - Logrus con niveles y formatos configurables
- ✅ **Hot Configuration** - Actualización de sensores sin reiniciar
- ✅ **Docker Ready** - Docker Compose con health checks
- ✅ **Tests Completos** - Unitarios + Integración end-to-end

## 🏗️ Arquitectura

```
┌─────────────┐
│  iot-cli    │ ← Cliente CLI (Cobra + NATS)
└──────┬──────┘
       │ NATS Request/Reply
       ↓
┌────────────────────────────────────────┐
│     iot-server                         │
│  ┌──────────────────────────────────┐  │
│  │  NATS Handlers                   │  │
│  │  - sensor.config.get/set.<id>    │  │
│  │  - sensor.readings.query.<id>    │  │
│  │  - sensor.register               │  │
│  │  - sensor.list                   │  │
│  └──────────────────────────────────┘  │
│           │                             │
│  ┌──────────────────────────────────┐  │
│  │  Simulator (Worker Pool)         │  │
│  │  ┌────┐ ┌────┐ ┌────┐ ┌────┐    │  │
│  │  │ W1 │ │ W2 │ │ W3 │ │ W4 │    │  │
│  │  └────┘ └────┘ └────┘ └────┘    │  │
│  │       TaskQueue (100 slots)      │  │
│  └──────────────────────────────────┘  │
│           │                             │
│  ┌──────────────────────────────────┐  │
│  │  Actions:                        │  │
│  │  1. SaveReading → Repository     │  │
│  │  2. Publish → NATS               │  │
│  │  3. Check Alert                  │  │
│  └──────────────────────────────────┘  │
└────────────────────────────────────────┘
         │            │
         ↓            ↓
    [SQLite]    [NATS Server]
```

## 🐳 Uso

### Comandos Docker Compose

```bash
# Levantar el sistema (NATS + IoT Server)
docker-compose up -d

# Ver logs
docker-compose logs -f iot-server

# Ver estado
docker-compose ps

# Detener el sistema
docker-compose down

# Detener y eliminar volúmenes
docker-compose down -v
```

### CLI - Gestión de Sensores

**Listar sensores:**
```bash
docker-compose run --rm iot-cli sensor list
```

**Registrar nuevo sensor:**
```bash
docker-compose run --rm iot-cli sensor register \
  --id pressure-002 \
  --type pressure \
  --interval 3000 \
  --threshold 1013.25
```

**Actualizar configuración:**
```bash
docker-compose run --rm iot-cli config set temp-001 \
  --interval 2000 \
  --threshold 28.5
```

**Consultar lecturas:**
```bash
docker-compose run --rm iot-cli readings temp-001
```

**Modo interactivo:**
```bash
docker-compose run --rm iot-cli interactive
```

### CLI Local (Desarrollo)

Si tienes Go instalado y prefieres desarrollo local:

```bash
# 1. Levantar solo NATS
docker run -d --name nats -p 4222:4222 nats:2.10-alpine

# 2. Compilar binarios
go build -o bin/iot-server ./cmd/iot-server
go build -o bin/iot-cli ./cmd/iot-cli

# 3. Ejecutar servidor
./bin/iot-server

# 4. Usar CLI (en otra terminal)
./bin/iot-cli sensor list
./bin/iot-cli config get temp-001
```

## 🧪 Tests

### Tests de Integración

```bash
# Ejecutar tests end-to-end
docker-compose --profile test run --rm iot-tests
```

**Los tests verifican:**
- ✅ Listado de sensores iniciales
- ✅ Registro dinámico de sensores
- ✅ Actualización de configuración
- ✅ Consulta de lecturas
- ✅ Sistema de alertas
- ✅ Persistencia en SQLite
- ✅ Reflejo de cambios en tiempo real

### Tests Unitarios

```bash
# Ejecutar tests localmente
go test ./... -v

# Con cobertura
go test ./... -cover

# Tests específicos
go test ./internal/simulator/... -v
go test ./internal/nats/... -v
```

Ver [README_TESTS.md](README_TESTS.md) para más detalles.

## 📁 Estructura del Proyecto

```
.
├── cmd/
│   ├── iot-server/        # Servidor IoT (main minimalista)
│   └── iot-cli/           # CLI client (Cobra)
├── internal/
│   ├── app/               # Inicialización del servidor
│   ├── sensor/            # Lógica de negocio
│   ├── simulator/         # Worker pool pattern
│   ├── nats/              # Mensajería y handlers
│   ├── repository/        # Interface de persistencia
│   ├── storage/           # Implementación SQLite
│   ├── config/            # Configuración (Viper)
│   └── logger/            # Logging (Logrus)
├── configs/               # YAML de configuración
├── test/integration/      # Tests end-to-end
├── docker-compose.yml     # Orquestación
├── Dockerfile             # Imagen del servidor
├── Dockerfile.cli         # Imagen del CLI
└── Dockerfile.test        # Imagen de tests
```

**¿Por qué esta estructura?**

- `cmd/` → Ejecutables desacoplados (server vs CLI)
- `internal/app/` → Encapsula inicialización (main.go de 28 líneas)
- `internal/` → Código no importable desde fuera (regla del compilador Go)
- Separación por responsabilidad: dominio vs infraestructura
- Facilita testing independiente por paquete
- Escalable sin reestructurar

Basado en el [Standard Go Project Layout](https://github.com/golang-standards/project-layout).

## 🎯 Decisiones Técnicas

### ¿Por qué Worker Pool en vez de 1 goroutine/sensor?

- ✅ Memoria constante independiente del número de sensores
- ✅ Menos context switches del scheduler de Go
- ✅ Patrón usado en sistemas IoT reales (EdgeX Foundry, Mainflux)
- ✅ Escalable a 1000+ sensores sin degradación

### ¿Por qué SQLite en vez de TimescaleDB?

Para datos time-series de sensores IoT, **TimescaleDB** sería ideal (hypertables, agregaciones automáticas, retención). Sin embargo, usamos **SQLite** por pragmatismo:

- Sin dependencias externas (driver puro Go sin CGO)
- Testing rápido con DB en memoria (`:memory:`)
- Suficiente para < 100K lecturas/día

La interface `Repository` permite cambiar a TimescaleDB creando `internal/storage/timescale.go` sin tocar lógica de negocio.

### ¿Por qué NATS?

- Subjects jerárquicos: `sensor.readings.<type>.<id>`
- Cliente con reconnect automático
- Request/Reply para configuración dinámica
- Testing con servidor NATS embebido

### Configuración

El sistema usa **Viper** para cargar configuración desde YAML + variables de entorno.

**Archivo:** `configs/values_local.yaml`

**Variables de entorno** (prefijo `IOT_`):
```bash
export IOT_NATS_URL=nats://production:4222
export IOT_DATABASE_PATH=/data/sensors.db
export IOT_LOG_LEVEL=warn
```

### Logging

**Logrus** con logging estructurado:

```json
{
  "level": "info",
  "msg": "[Simulator] Sensor added",
  "sensor_id": "temp-001",
  "type": "temperature",
  "interval": 5000,
  "time": "2025-10-23T15:04:05Z"
}
```

## 🔍 Monitoreo

**NATS Monitoring:**
- URL: http://localhost:8222
- Proporciona métricas de conexiones, mensajes, subscripciones

**Logs del servidor:**
```bash
docker-compose logs -f iot-server
```

## 📚 Documentación Adicional

- [DOCKER.md](DOCKER.md) - Guía completa de Docker Compose
- [README_TESTS.md](README_TESTS.md) - Guía de testing detallada
- [CHANGELOG.md](CHANGELOG.md) - Historial de cambios por feature

## 🤝 Contribuir

1. Los cambios deben incluir tests
2. Ejecutar `go fmt` antes de commit
3. Ejecutar tests: `docker-compose --profile test run --rm iot-tests`
4. Documentar decisiones de diseño en commits

## 📝 Licencia

Este proyecto es una prueba técnica para demostración de habilidades.
