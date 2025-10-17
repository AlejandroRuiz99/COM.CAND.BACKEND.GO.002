# Changelog

Todos los cambios notables en este proyecto serán documentados en este archivo.

## [Unreleased]

### En desarrollo
- feat-3: Simuladores de sensores

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


