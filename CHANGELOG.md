# Changelog

Todos los cambios notables en este proyecto serán documentados en este archivo.

## [Unreleased]

### En desarrollo
- feat-0: Project setup & foundation

## [0.1.0] - 2025-10-16

### Added (feat-0: Project Setup & Foundation)

- Estructura de carpetas siguiendo Standard Go Layout
- Inicialización de módulo Go 1.25.3 con dependencias (NATS, SQLite)
- Tipos de dominio básicos:
  - `Sensor`: Representación de sensor físico
  - `SensorType`: Enum para tipos (temperature, humidity, pressure)
  - `SensorConfig`: Configuración de sensores
  - `SensorReading`: Lectura de sensor con timestamp y manejo de errores
- Validación de tipos con método `Validate()` para SensorReading y SensorConfig
- Tests unitarios con table-driven tests (TDD)
- Esqueleto de aplicación principal con graceful shutdown
- Documentación inicial:
  - README
  - CHANGELOG para seguimiento de cambios

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


