# Changelog

Todos los cambios notables en este proyecto serán documentados en este archivo.

## [Unreleased]

## [0.6.0] - 2025-10-22

### Added (feat-6: CLI Client)

- **CLI client completo** (`iot-cli`) con Cobra framework
  - `sensor register` - Registrar sensores dinámicamente con validación
  - `config get/set` - Consultar/modificar configuración de sensores
  - `readings` - Obtener lecturas históricas con estadísticas (promedio, máx, mín)
- **Flags globales**:
  - `--nats-url` - Especificar servidor NATS (default: localhost:4222)
  - `--json` - Output en formato JSON para integración con scripts
  - `--debug` - Activar logs verbosos con Logrus
- **Tablas formateadas** con `rodaine/table` para mejor UX
- **Arquitectura simplificada**:
  - Main.go reducido a 28 líneas (antes: 208 líneas, -86% código)
  - Paquete `internal/app` encapsula toda la lógica de inicialización del servidor
  - Separación completa: CLI y Server desacoplados
- **Logger mejorado**:
  - CLI con Logrus (logs a stderr, output a stdout)
  - Inicialización en dos fases: básica → configurada por Viper

### Changed

- Refactorizado `cmd/iot-server/main.go` para usar `internal/app/server.go`
- CLI ahora usa Logrus en lugar de fmt para logs de debug/error

## [0.5.0] - 2025-10-21

### Added (feat-5: Handlers NATS adicionales)

- **Nuevos handlers NATS** (completa 100% requisitos):
  - `sensor.readings.query.<id>` - Consultar últimas N lecturas de un sensor
  - `sensor.register` - Registrar sensores dinámicamente sin reiniciar servidor
- Callback en handlers para integración con simulador
- Tests de integración para nuevos handlers (coverage nats: 65.9%)

## [0.4.0] - 2025-10-20

### Added (feat-4: Servicio Orquestador + Configuración)

- Sistema de configuración multi-entorno con Viper (YAML + variables de entorno `IOT_*`)
- Logging estructurado con Logrus (niveles configurables, formatos JSON/text)
- **Simulador con Worker Pool Pattern** (5 workers + task queue de 100 slots)
- Gestión dinámica de sensores (`AddSensor`, `RemoveSensor`, `UpdateSensorConfig`)
- Handlers NATS básicos:
  - `sensor.config.get.<id>` - Obtener configuración
  - `sensor.config.set.<id>` - Actualizar configuración
- Main.go completo con inicialización de componentes y graceful shutdown
- Sistema de alertas cuando valores exceden thresholds configurables
- Reorganización arquitectónica: `internal/simulator/`, `internal/repository/`, `internal/logger/`
- Tests completos con cobertura del 80.7% (config 87.5%, simulator 81.7%, sensor 94.1%)

### Fixed

- Race condition en `Stop()` del simulador (panic al cerrar workers)
- Workers ahora manejan correctamente el cierre del task queue

## [0.3.0] - 2025-10-17

### Added (feat-3: Sensor Simulators)

- Simulador con generación automática de lecturas periódicas
- Valores realistas por tipo de sensor (temperatura, humedad, presión)
- Simulación de errores de lectura (5% probabilidad)
- Ejecución con goroutines y cancelación via context
- Configuración dinámica thread-safe con sync.RWMutex

## [0.3.0] - 2025-10-16

### Added (feat-2: NATS Client & Messaging)

- Cliente NATS con wrapper y opciones optimizadas (reconnect, timeouts)
- Subjects helpers organizados jerárquicamente (readings, config, alerts)
- Handlers request/reply para configuración de sensores (GET/SET)
- Tests de integración con servidor NATS embebido
- Mock repository para testing de handlers

## [0.2.0] - 2025-10-16

### Added (feat-1: Repository Pattern & SQLite Persistence)

- Interface `Repository` para desacoplar persistencia de lógica de negocio
- Implementación `SQLiteRepository` con driver puro Go (sin CGO)
- Schema SQL embedido con índices optimizados para time-series
- Tests de integración con DB en memoria (cobertura >90%)
- CRUD completo: SaveReading, GetLatestReadings, GetByTimeRange, SaveConfig

---

## [0.1.0] - 2025-10-16

### Added (feat-0: Project Setup & Foundation)

- Estructura del proyecto siguiendo Standard Go Layout (cmd/, internal/)
- Tipos de dominio: Sensor, SensorType, SensorConfig, SensorReading con validación
- Tests unitarios con table-driven tests (cobertura 100%)
- Módulo Go 1.25.3 con dependencias NATS y SQLite
- Main básico y documentación inicial (README, CHANGELOG)

---

## Formato de Commits

Este proyecto usa commits convencionales:

- `feat:` Nueva funcionalidad
- `fix:` Corrección de bugs
- `test:` Añadir o modificar tests
- `docs:` Cambios en documentación
- `refactor:` Refactorización de código

## Branches

- `master`: Rama principal (producción)
- `development`: Integración continua
- `feat-N`: Features en desarrollo


