# technical_test_uvigo

Prueba técnica IoT - Sistema de gestión de sensores con Go y NATS

## Estructura del Proyecto

```
.
├── cmd/
│   ├── iot-server/      # Servidor IoT (main minimalista)
│   └── iot-cli/         # CLI client (comandos Cobra)
├── internal/
│   ├── app/             # Lógica de inicialización del servidor
│   ├── sensor/          # Lógica de negocio (tipos, validaciones)
│   ├── simulator/       # Simulador con worker pool
│   ├── nats/            # Cliente de mensajería y handlers
│   ├── repository/      # Interface de persistencia
│   ├── storage/         # Implementaciones de Repository
│   ├── config/          # Configuración con Viper
│   └── logger/          # Logger con Logrus
├── configs/             # Archivos de configuración YAML
└── docs/                # Documentación
```

**Por qué esta estructura:**

- `cmd/` contiene ejecutables desacoplados (server: siempre on, CLI: on-demand)
- `internal/app/` encapsula inicialización del servidor (main.go de 28 líneas)
- `internal/` asegura que el código no pueda ser importado desde fuera (regla del compilador Go)
- Separación por responsabilidad: dominio vs infraestructura
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

## CLI Client

El sistema incluye un **CLI client completo** (`iot-cli`) para gestionar sensores remotamente vía NATS.

### Comandos disponibles

**1. Registrar sensor dinámicamente**
```bash
./bin/iot-cli sensor register \
  --id pressure-002 \
  --type pressure \
  --name "Sensor Lab" \
  --interval 3000 \
  --threshold 1013.25
```

**2. Consultar configuración**
```bash
./bin/iot-cli config get temp-001
```

**3. Actualizar configuración**
```bash
./bin/iot-cli config set temp-001 \
  --interval 2000 \
  --threshold 28.5 \
  --enabled=false
```

**4. Obtener lecturas históricas**
```bash
./bin/iot-cli readings temp-001 --limit 10
```

### Flags globales

- `--nats-url string` - URL del servidor NATS (default: `nats://localhost:4222`)
- `--json` - Output en formato JSON (útil para scripts)
- `--debug` - Activar logs verbosos con Logrus

**Ejemplos:**
```bash
# Conectar a servidor remoto
./bin/iot-cli --nats-url nats://prod-server:4222 config get temp-001

# Output JSON para procesamiento
./bin/iot-cli --json readings temp-001 | jq '.[] | .value'

# Modo debug
./bin/iot-cli --debug sensor register --id test-001 --type temperature
```

### Características del CLI

- ✅ **Tablas formateadas** con estadísticas (promedio, máx, mín)
- ✅ **Validación de entrada** antes de enviar a servidor
- ✅ **Logging estructurado** con Logrus (logs a stderr, output a stdout)
- ✅ **Desacoplado del servidor** (se comunica solo via NATS)

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

### 3. Compilar binarios

```bash
# Compilar servidor y CLI
go build -o bin/iot-server.exe ./cmd/iot-server
go build -o bin/iot-cli.exe ./cmd/iot-cli
```

### 4. Ejecutar el servidor IoT

```bash
./bin/iot-server.exe
```

### 5. Usar el CLI (en otra terminal)

```bash
# Ver sensores configurados
./bin/iot-cli config get temp-001

# Registrar nuevo sensor
./bin/iot-cli sensor register --id temp-999 --type temperature

# Ver lecturas
./bin/iot-cli readings temp-001 --limit 5
```

### 6. Verificar funcionamiento

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

### 7. Detener el servidor

Presiona `Ctrl+C` en la terminal del servidor para un shutdown limpio.

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
┌─────────────┐
│  iot-cli    │ ← Cliente remoto (Cobra + NATS)
│  (Cobra)    │
└──────┬──────┘
       │ NATS Request/Reply
       │ (config, readings, register)
       ↓
┌────────────────────────────────────────────┐
│     iot-server (internal/app)              │
│                                            │
│  ┌──────────────────────────────────────┐ │
│  │  NATS Handlers                       │ │
│  │  - sensor.config.get/set.<id>        │ │
│  │  - sensor.readings.query.<id>        │ │
│  │  - sensor.register                   │ │
│  └──────────────────────────────────────┘ │
│           │                                │
│           ↓                                │
│  ┌──────────────────────────────────────┐ │
│  │  Simulator (Worker Pool)             │ │
│  │  ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ │ │
│  │  │ W1 │ │ W2 │ │ W3 │ │ W4 │ │ W5 │ │ │
│  │  └────┘ └────┘ └────┘ └────┘ └────┘ │ │
│  │       │ TaskQueue (100 slots) │      │ │
│  │  temp-001 │ hum-001 │ press-001      │ │
│  └──────────────────────────────────────┘ │
│           │                                │
│           ↓                                │
│  ┌──────────────────────────────────────┐ │
│  │  Actions:                            │ │
│  │  1. SaveReading → Repository         │ │
│  │  2. Publish → NATS                   │ │
│  │  3. Check Alert                      │ │
│  └──────────────────────────────────────┘ │
└────────────────────────────────────────────┘
         │            │
         ↓            ↓
    [SQLite]    [NATS Server]
                     │
                     ↓ Pub/Sub
              [Subscribers] ← feat-7 (Docker)
                [Dashboards]
```

**Características clave:**
- ✅ **CLI y Server desacoplados** (comunicación solo via NATS)
- ✅ **Worker Pool** escalable (5 workers, queue de 100 tareas)
- ✅ **Main.go minimalista** (28 líneas, lógica en `internal/app`)
- ✅ **Registro dinámico** de sensores sin reiniciar servidor

