# Changelog

Todos los cambios notables en este proyecto serán documentados en este archivo.

## [Unreleased]

### Changed

- Tests consolidados con table-driven pattern en `config_test.go`
- README.md simplificado enfocado en Docker Compose
- CHANGELOG.md reorganizado con features correctas

## [0.8.0] - 2025-10-23

### Added (feat-8: Tests de Integración End-to-End)

- Tests de integración end-to-end con 8 subtests (registro, configuración, lecturas, alertas)
- Dockerfile.test e integración con docker-compose usando profile `test`
- IDs únicos por ejecución usando timestamps para evitar conflictos
- Debug logging y validación exhaustiva de persistencia SQLite

### Fixed

- Handler `sensor.register` guarda configuración en repositorio antes de añadir al simulador
- Tests usan estructura JSON correcta con campo `Config` anidado
- Variable de entorno `NATS_URL` configurada correctamente para tests en Docker

## [0.7.0] - 2025-10-22

### Added (feat-7: Docker Compose Setup)

- Docker multi-stage builds para servidor (~15MB), CLI (~12MB) y tests
- docker-compose.yml con NATS, iot-server, profiles para CLI y tests
- Networking automático (`iot-network`), healthchecks y persistencia en volumen `iot-data`
- Documentación completa (README.md, DOCKER.md, README_TESTS.md)

### Changed

- Configuración SQLite usa path absoluto `/data/sensors.db` para Docker
- README reorganizado con Docker Compose como opción principal

## [0.6.0] - 2025-10-22

### Added (feat-6: CLI Client)

- CLI completo con Cobra: `sensor register`, `config get/set`, `readings`
- Flags globales: `--nats-url`, `--json`, `--debug` con logging Logrus
- Tablas formateadas con estadísticas (promedio, máx, mín)
- Modo interactivo para uso fluido sin repetir `iot-cli`

### Changed

- Main.go del servidor reducido a 28 líneas (-86% código)
- Lógica de inicialización movida a `internal/app/server.go`
- CLI y Server completamente desacoplados (comunicación solo vía NATS)

## [0.5.0] - 2025-10-21

### Added (feat-5: Handlers NATS adicionales)

- **Nuevos handlers NATS** (completa 100% requisitos):
  - `sensor.readings.query.<id>` - Consultar últimas N lecturas de un sensor
  - `sensor.register` - Registrar sensores dinámicamente sin reiniciar servidor
- Callback en handlers para integración con simulador
- Tests de integración para nuevos handlers (coverage nats: 65.9%)

## [0.4.0] - 2025-10-20

### Added (feat-4: Servicio Orquestador + Configuración)

- Simulador con Worker Pool Pattern (5 workers + task queue de 100 slots)
- Configuración multi-entorno con Viper (YAML + variables `IOT_*`)
- Handlers NATS básicos: `sensor.config.get/set.<id>`
- Sistema de alertas con thresholds configurables y logging Logrus

### Fixed

- Race condition en `Stop()` del simulador al cerrar workers

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


