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
- Testing con servidor NATS embebido.

