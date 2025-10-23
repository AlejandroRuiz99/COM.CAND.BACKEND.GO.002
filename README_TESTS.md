# 🧪 Tests de Integración - Sistema IoT

## 📋 Requisitos Previos

1. El sistema Docker debe estar levantado:
```bash
docker-compose up -d
```

2. Esperar a que el servidor esté completamente iniciado (unos segundos)

## 🚀 Ejecutar Tests

### Comando Simple

```bash
docker-compose --profile test run --rm iot-tests
```

Este comando:
- ✅ Construye la imagen de tests (si no existe)
- ✅ Espera a que el servidor esté healthy
- ✅ Ejecuta todos los tests de integración
- ✅ Elimina el contenedor al finalizar

### Reconstruir y Ejecutar

Si modificaste los tests:

```bash
docker-compose --profile test build iot-tests
docker-compose --profile test run --rm iot-tests
```

## 📊 Tests Incluidos

### TestSystemIntegration

Prueba el flujo completo del sistema:

1. **01_ListInitialSensors** ✅
   - Verifica que se pueden listar sensores
   - Comprueba que existen al menos 4 sensores iniciales

2. **02_RegisterNewSensor**
   - Registra un nuevo sensor con ID único
   - Verifica respuesta del servidor

3. **03_VerifySensorInList**
   - Confirma que el sensor registrado aparece en la lista
   - Valida configuración inicial

4. **04_GetSensorConfig**
   - Obtiene la configuración del sensor
   - Verifica valores configurados

5. **05_UpdateSensorConfig**
   - Actualiza interval y threshold
   - Confirma respuesta exitosa

6. **06_VerifyConfigUpdated**
   - Verifica que los cambios se reflejan
   - Comprueba tanto `config.get` como `sensor.list`

7. **07_QuerySensorReadings**
   - Consulta lecturas del sensor
   - Verifica que hay datos disponibles

8. **08_VerifyAlerts** ✅
   - Se suscribe al subject de alertas
   - Verifica la capacidad de recibir alertas

### TestCLICommands

Prueba los comandos del CLI:
- Skipped cuando se ejecuta en Docker (no tiene binario local)

## 🔍 Ver Logs Detallados

```bash
# Logs del servidor durante los tests
docker-compose logs -f iot-server

# Logs de NATS
docker-compose logs -f nats
```

## 🛠️ Troubleshooting

### Los tests fallan con "nats: no servers available"

Asegúrate de que el sistema está levantado:
```bash
docker-compose ps
docker-compose logs iot-server
```

### Quiero ejecutar un test específico

Modifica `Dockerfile.test` temporalmente:
```dockerfile
CMD ["go", "test", "-v", "./test/integration/...", "-run", "TestSystemIntegration/01_ListInitialSensors", "-timeout", "2m"]
```

### Limpiar estado entre ejecuciones

```bash
docker-compose down -v
docker-compose up -d
# Esperar unos segundos
docker-compose --profile test run --rm iot-tests
```

## ✅ Pruebas Manuales Complementarias

Para verificar manualmente el sistema:

```bash
# 1. Listar sensores
docker-compose run --rm iot-cli sensor list

# 2. Registrar sensor
docker-compose run --rm iot-cli sensor register \
  --id manual-test-001 \
  --type temperature \
  --interval 5000 \
  --threshold 30.0

# 3. Verificar en la lista
docker-compose run --rm iot-cli sensor list

# 4. Actualizar configuración
docker-compose run --rm iot-cli config set manual-test-001 \
  --interval 3000 \
  --threshold 35.0

# 5. Verificar cambios
docker-compose run --rm iot-cli sensor list
docker-compose run --rm iot-cli config get manual-test-001

# 6. Consultar lecturas
docker-compose run --rm iot-cli readings manual-test-001
```

## 📝 Notas

- Los tests usan IDs únicos (timestamp) para evitar conflictos
- Cada ejecución crea un sensor nuevo
- El sistema persiste datos en un volumen Docker
- Para limpiar completamente: `docker-compose down -v`

## 🎯 Casos de Uso Verificados

✅ **Funcionalidades Core:**
- Conexión a NATS
- Listado de sensores
- Registro dinámico de sensores
- Actualización de configuración
- Consulta de lecturas
- Sistema de alertas

✅ **Integración:**
- Docker Compose orquestación
- Volúmenes persistentes
- Health checks
- Variables de entorno
- Networking entre servicios

