# Decisiones Técnicas - Sistema IoT

Este documento explica las 10 decisiones más importantes de diseño y arquitectura del proyecto.

## 1. Worker Pool Pattern en vez de 1 Goroutine por Sensor

**Decisión:** Usar 5 workers fijos con una task queue en lugar de crear una goroutine por sensor.

**Por qué:**
Cuando empecé a diseñar el simulador, mi primer instinto fue "una goroutine por sensor, total son baratas". Pero luego pensé: ¿qué pasa si hay 1000 sensores? ¿10000? El scheduler de Go se empezaría a quejar. 

En sistemas IoT reales como EdgeX Foundry usan worker pools precisamente por esto. La ventaja es clara:
- Memoria constante: 5 workers siempre, no importa cuántos sensores
- Control total sobre la concurrencia
- Más fácil de monitorear y debuggear
- Si un sensor tarda mucho, no bloquea a otros (queue con timeout)

**Trade-off:** Un poco más de complejidad en el código, pero vale la pena.

## 2. SQLite en vez de TimescaleDB/InfluxDB

**Decisión:** Usar SQLite para persistencia.

**Por qué:**
Esto fue 100% pragmatismo. Para una prueba técnica, no tiene sentido montar TimescaleDB (que sería lo ideal para time-series). SQLite me da:
- Cero dependencias externas - solo un archivo
- Tests súper rápidos con `:memory:`
- Driver puro Go sin CGO (compila en cualquier lado)
- Suficiente para < 100K lecturas/día

Además, usé el Repository pattern precisamente para poder cambiar a TimescaleDB después sin tocar nada de la lógica de negocio. Solo crearía `internal/storage/timescale.go` y listo.

**En producción:** Definitivamente TimescaleDB o InfluxDB.

## 3. NATS para Mensajería

**Decisión:** NATS como broker de mensajes.

**Por qué:**
Comparé NATS vs RabbitMQ vs Kafka. Para IoT, NATS es la opción obvia:
- Súper ligero (~20MB de RAM vs 500MB+ de Kafka)
- Latencia bajísima (microsegundos)
- Subjects jerárquicos perfectos para sensores: `sensor.readings.temperature.temp-001`
- Request/Reply nativo (no necesito REST para config)
- Testing fácil con nats-server embebido

Kafka sería overkill para esto. RabbitMQ es bueno pero más pesado y complejo.

## 4. Repository Pattern para Desacoplar Persistencia

**Decisión:** Crear interface `Repository` en vez de usar SQLite directamente.

**Por qué:**
Esta es arquitectura 101 pero mucha gente se la salta. La ventaja es brutal:
- Tests ultra rápidos con mock repository (sin DB)
- Puedo cambiar de SQLite a Postgres sin tocar simulador/handlers
- Facilita testing: `repository_test.go` prueba solo persistencia
- En producción podría tener múltiples implementaciones: SQL + cache

El código no es más complicado, solo más flexible.

## 5. CLI y Server Completamente Desacoplados

**Decisión:** CLI y Server se comunican SOLO vía NATS, cero dependencias compartidas.

**Por qué:**
Al principio pensé en hacer el CLI como librería que importara código del server. Error. 

Separándolos completamente:
- CLI puede estar en otro lenguaje (Python, Node) - solo necesita NATS
- Puedo actualizar server sin recompilar CLI
- Testing independiente
- Deploy independiente (CLI en laptop, server en cloud)
- Escalabilidad: múltiples CLIs contra un cluster de servers

**Bonus:** El CLI tiene Cobra con modo interactivo que está bastante pulido.

## 6. Docker Multi-stage Builds

**Decisión:** Usar multi-stage builds (builder + runtime).

**Por qué:**
Nadie quiere una imagen de 1GB con compiladores y toolchains. Multi-stage me da:
- Imagen del server: 15MB (vs 800MB+ con builder incluido)
- Imagen del CLI: 12MB
- Alpine Linux en runtime = superficie de ataque mínima

En CI/CD esto ahorra un montón de tiempo de push/pull.

## 7. Configuración con Viper (YAML + ENV)

**Decisión:** Viper para config en vez de JSON puro o flags.

**Por qué:**
Necesitaba flexibilidad para diferentes entornos:
- Dev: `configs/values_local.yaml`
- Prod: variables de entorno `IOT_*`
- Override selectivo: puedo cargar YAML y sobreescribir solo NATS_URL

Viper me da todo esto gratis. Además, el hot-reload de config es trivial si lo necesito después.

## 8. Logrus para Logging Estructurado

**Decisión:** Logrus en vez de log estándar de Go.

**Por qué:**
En producción necesitas logs en JSON para parsear con ELK/Grafana. El log estándar de Go es texto plano.

Logrus me da:
- Niveles (debug, info, warn, error)
- Campos estructurados: `log.WithFields(logrus.Fields{"sensor_id": id})`
- Output configurable (JSON en prod, text en dev)
- Hooks para enviar a sistemas externos

**Trade-off:** Sí, sé que hay librerías más modernas (zerolog, zap) pero Logrus es estable y suficiente.

## 9. Table-Driven Tests

**Decisión:** Todos los tests con table-driven pattern.

**Por qué:**
Esto es Go idiomático. En vez de:
```go
func TestValidate_MissingID(t *testing.T) { ... }
func TestValidate_MissingType(t *testing.T) { ... }
func TestValidate_MissingName(t *testing.T) { ... }
```

Hago:
```go
tests := []struct {
    name string
    input Config
    wantErr bool
}{
    {"missing ID", Config{...}, true},
    {"missing type", Config{...}, true},
    ...
}
```

Ventajas:
- Añadir casos de test es trivial (una línea)
- Output claro con subtests
- Menos código duplicado
- Más fácil de mantener

## 10. Main.go Minimalista (28 líneas)

**Decisión:** Mover toda la lógica a `internal/app/server.go`.

**Por qué:**
Esto fue un refactor consciente. Al principio main.go tenía 208 líneas (inicialización, config, graceful shutdown, etc.). Era un desastre.

Ahora main.go hace solo:
```go
func main() {
    cfg := config.LoadFromEnv()
    srv := app.NewServer(cfg)
    srv.Run()
}
```

**Ventajas:**
- Testing: puedo testear `app.Server` sin main
- Reutilización: puedo usar Server en otros contextos (tests E2E, benchmarks)
- Claridad: main es el entry point, App es la aplicación
- Deploy: puedo tener múltiples mains (iot-server, iot-worker, etc.) usando la misma App

---

## Resumen

Todas estas decisiones tienen un hilo común: **simplicidad, testabilidad y escalabilidad**. 

Elegí herramientas maduras y patrones probados en lugar de experimentar. En una prueba técnica, quiero demostrar que sé construir sistemas mantenibles, no que conozco la última librería de moda.

¿Hay cosas que haría diferente en producción? Claro (ver `MEJORAS_PRODUCTIVAS.md`). Pero para el scope de esta prueba, estas decisiones son las correctas.

