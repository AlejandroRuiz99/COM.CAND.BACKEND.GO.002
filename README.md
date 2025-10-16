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

