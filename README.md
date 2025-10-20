# technical_test_uvigo

Prueba técnica IoT - Sistema de gestión de sensores con Go y NATS

## Estructura del Proyecto

```
.
├── cmd/iot-server/      # Punto de entrada (main.go)
├── internal/
│   ├── sensor/          # Lógica de negocio (tipos, validaciones)
│   ├── nats/            # Cliente de mensajería
│   └── storage/         # Persistencia
├── configs/             # Archivos de configuración
└── docs/                # Documentación
```

**Por qué esta estructura:**

- `cmd/` contiene ejecutables. Permite tener múltiples binarios si se necesitan (server, CLI, etc.)
- `internal/` asegura que el código no pueda ser importado desde fuera (regla del compilador Go)
- Separación por responsabilidad: `sensor/` es dominio, `nats/` y `storage/` son infraestructura
- Facilita testing: cada paquete se testea independientemente
- Escalable: añadir features no requiere reestructurar

Basado en el [Standard Go Project Layout](https://github.com/golang-standards/project-layout).

## Persistencia

Para datos time-series de sensores IoT, lo ideal sería **TimescaleDB** (hypertables, agregaciones automáticas, retención). Sin embargo, usamos **SQLite** para esta prueba técnica por pragmatismo:

- Sin dependencias externas (driver puro Go sin CGO)
- Testing rápido con DB en memoria (`:memory:`)
- Suficiente para < 100K lecturas/día

La interface `Repository` desacopla la persistencia: cambiar de SQLite a TimescaleDB solo requiere crear `internal/storage/timescale.go` sin tocar lógica de negocio.

## Mensajería

**NATS** para comunicación pub/sub y request/reply:

- Subjects jerárquicos: `sensor.readings.<type>.<id>`, `sensor.config.<get|set>.<id>`
- Cliente con reconnect automático y timeouts configurables
- Handlers para configuración dinámica de sensores vía NATS
- Testing con servidor NATS embebido

## Simuladores

**Generación automática** de lecturas de sensores con **Worker Pool Pattern**:

- **Arquitectura escalable**: 5 workers fijos procesan todos los sensores
- **Task Queue**: Buffer de 100 tareas con backpressure automático
- Valores realistas por tipo: temperatura (15-35°C), humedad (30-80%), presión (980-1040 hPa)
- Simulación de errores aleatorios (5% probabilidad)
- Thread-safe: configuración actualizable en caliente

**¿Por qué Worker Pool y no 1 goroutine/sensor?**
- ✅ Memoria constante independiente del número de sensores
- ✅ Menos context switches del scheduler de Go
- ✅ Patrón usado en sistemas IoT reales (EdgeX Foundry, Mainflux)
- ✅ Escalable a 1000+ sensores sin degradación

## Configuración

El sistema usa **Viper** para cargar configuración desde archivos YAML + variables de entorno.

### Archivos de configuración

- `configs/values_local.yaml` - Desarrollo local (incluido en repo)
- Otras configuraciones (INT, QA, PROD) deben estar en un **repositorio externo** de configuraciones

### Variables de entorno

**Cargar archivo específico:**
```bash
export CONFIG_FILE=/path/to/values_qa.yaml
```

**Override de valores individuales** (prefijo `IOT_`):
```bash
export IOT_NATS_URL=nats://production:4222
export IOT_DATABASE_TYPE=influxdb
export IOT_DATABASE_PATH=/data/sensors.db
export IOT_LOGGING_LEVEL=warn
```

Si no se especifica `CONFIG_FILE`, usa `values_local.yaml` por defecto.

## Logging

El sistema usa **Logrus** para logging estructurado:

- **Niveles**: debug, info, warn, error
- **Formatos**: text (desarrollo), json (producción)
- **Configuración** en YAML:
  ```yaml
  logging:
    level: info
    format: json
  ```

**Logs con campos estructurados:**
```json
{
  "level": "info",
  "msg": "[Manager] Started simulator",
  "sensor_id": "temp-001",
  "type": "temperature",
  "time": "2025-10-20T15:04:05Z"
}
```

## Uso

### 1. Instalar dependencias

```bash
go mod download
```

### 2. Levantar NATS Server

**Opción A: Con Docker**
```bash
docker run -d --name nats -p 4222:4222 -p 8222:8222 nats:2.10-alpine -js -m 8222
```

**Opción B: Instalación local**
Descarga desde [nats.io](https://nats.io/download/)

### 3. Ejecutar el servidor IoT

```bash
go run cmd/iot-server/main.go
```

O compilar y ejecutar:
```bash
go build -o bin/iot-server.exe cmd/iot-server/main.go
./bin/iot-server.exe
```

### 4. Verificar funcionamiento

El servidor mostrará logs como:
```
═══════════════════════════════════════════════════════
   🚀 IoT Sensor Server is RUNNING
═══════════════════════════════════════════════════════

📊 System Status:
   • NATS:      nats://localhost:4222 ✓
   • Database:  sqlite ✓
   • Simulators: 4 active

📡 Publishing to NATS subjects:
   • sensor.readings.<type>.<id>  (sensor readings)
   • sensor.alerts.<type>.<id>    (threshold alerts)
```

**Logs de lecturas:**
```
[Manager] ALERT: Sensor temp-001 exceeded threshold: 32.45 °C > 30.00 °C
```

### 5. Detener el servidor

Presiona `Ctrl+C` para un shutdown limpio.

## Testing

```bash
# Ejecutar todos los tests
go test ./...

# Con cobertura
go test ./... -cover

# Tests específicos
go test ./internal/config/... -v
go test ./internal/manager/... -v
```

## Arquitectura

```
Usuario/Front
     │
     ↓
[API REST]  ← feat-6 (próximamente)
     │
     ↓
┌────────────────────────────────────┐
│     IoT Server (main.go)           │
│                                    │
│  ┌──────────────────────────┐     │
│  │  SimulatorManager        │     │
│  │  - temp-001              │     │
│  │  - hum-001               │     │
│  │  - press-001             │     │
│  └──────────────────────────┘     │
│           │                        │
│           ↓                        │
│  ┌──────────────────────────┐     │
│  │  Callbacks:              │     │
│  │  1. SaveReading → DB     │     │
│  │  2. Publish → NATS       │     │
│  │  3. Check Alert          │     │
│  └──────────────────────────┘     │
└────────────────────────────────────┘
         │            │
         ↓            ↓
    [SQLite]    [NATS Server]
                     │
                     ↓
              [Subscribers] ← feat-7 (próximamente)
```

