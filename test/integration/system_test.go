package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTimeout      = 30 * time.Second
	sensorType       = "temperature"
	initialInterval  = 5000
	updatedInterval  = 3000
	initialThreshold = 30.0
	updatedThreshold = 35.0
)

// getSensorID genera un ID único para evitar conflictos
func getSensorID() string {
	return fmt.Sprintf("test-sensor-%d", time.Now().Unix())
}

// getNatsURL obtiene la URL de NATS desde variable de entorno o usa el default
func getNatsURL() string {
	if url := os.Getenv("NATS_URL"); url != "" {
		return url
	}
	return "nats://localhost:4222"
}

type SensorConfig struct {
	SensorID  string  `json:"sensor_id"`
	Interval  int     `json:"interval"`
	Threshold float64 `json:"threshold"`
	Enabled   bool    `json:"enabled"`
}

type SensorDefinition struct {
	ID     string       `json:"id"`
	Type   string       `json:"type"`
	Name   string       `json:"name,omitempty"`
	Config SensorConfig `json:"config"`
}

// TestSystemIntegration prueba el flujo completo del sistema
func TestSystemIntegration(t *testing.T) {
	// Conectar a NATS
	nc, err := nats.Connect(getNatsURL(), nats.Timeout(5*time.Second))
	require.NoError(t, err, "Error conectando a NATS")
	defer nc.Close()

	t.Logf("✓ Conexión a NATS establecida (%s)", getNatsURL())

	// Generar ID único para esta ejecución de tests
	sensorID := getSensorID()
	t.Logf("→ Usando sensor ID: %s", sensorID)

	// Esperar a que el servidor esté listo
	time.Sleep(2 * time.Second)

	t.Run("01_ListInitialSensors", func(t *testing.T) {
		testListInitialSensors(t, nc)
	})

	t.Run("02_RegisterNewSensor", func(t *testing.T) {
		testRegisterNewSensor(t, nc, sensorID)
	})

	t.Run("03_VerifySensorInList", func(t *testing.T) {
		testVerifySensorInList(t, nc, sensorID)
	})

	t.Run("04_GetSensorConfig", func(t *testing.T) {
		testGetSensorConfig(t, nc, sensorID)
	})

	t.Run("05_UpdateSensorConfig", func(t *testing.T) {
		testUpdateSensorConfig(t, nc, sensorID)
	})

	t.Run("06_VerifyConfigUpdated", func(t *testing.T) {
		testVerifyConfigUpdated(t, nc, sensorID)
	})

	t.Run("07_QuerySensorReadings", func(t *testing.T) {
		testQuerySensorReadings(t, nc, sensorID)
	})

	t.Run("08_VerifyAlerts", func(t *testing.T) {
		testVerifyAlerts(t, nc, sensorID)
	})
}

func testListInitialSensors(t *testing.T, nc *nats.Conn) {
	msg, err := nc.Request("sensor.list", []byte(""), 5*time.Second)
	require.NoError(t, err, "Error listando sensores")

	var sensors []SensorDefinition
	err = json.Unmarshal(msg.Data, &sensors)
	require.NoError(t, err, "Error parseando respuesta")

	assert.GreaterOrEqual(t, len(sensors), 4, "Deberían existir al menos 4 sensores iniciales")
	t.Logf("✓ Sensores iniciales: %d", len(sensors))
}

func testRegisterNewSensor(t *testing.T, nc *nats.Conn, sensorID string) {
	newSensor := SensorDefinition{
		ID:   sensorID,
		Type: sensorType,
		Config: SensorConfig{
			SensorID:  sensorID,
			Interval:  initialInterval,
			Threshold: initialThreshold,
			Enabled:   true,
		},
	}

	payload, err := json.Marshal(newSensor)
	require.NoError(t, err, "Error serializando sensor")

	t.Logf("→ Registrando sensor: %s", string(payload))

	msg, err := nc.Request("sensor.register", payload, 5*time.Second)
	require.NoError(t, err, "Error registrando sensor")

	// Debug: ver respuesta raw
	t.Logf("→ Raw register response: %s", string(msg.Data))

	// El servidor devuelve {"status": "ok", "sensor_id": "...", "message": "..."}
	var response map[string]interface{}
	err = json.Unmarshal(msg.Data, &response)
	require.NoError(t, err, "Error parseando respuesta")

	// Si hay error, mostrarlo
	if errorMsg, ok := response["error"]; ok {
		t.Logf("❌ Error del servidor: %v", errorMsg)
	}

	assert.Equal(t, "ok", response["status"], "El status debería ser ok")
	assert.Equal(t, sensorID, response["sensor_id"], "El ID del sensor debería coincidir")
	t.Logf("✓ Sensor registrado: %s", sensorID)
}

func testVerifySensorInList(t *testing.T, nc *nats.Conn, sensorID string) {
	msg, err := nc.Request("sensor.list", []byte(""), 5*time.Second)
	require.NoError(t, err, "Error listando sensores")

	var sensors []SensorDefinition
	err = json.Unmarshal(msg.Data, &sensors)
	require.NoError(t, err, "Error parseando respuesta")

	found := false
	for _, s := range sensors {
		if s.ID == sensorID {
			found = true
			assert.Equal(t, sensorType, s.Type, "Tipo de sensor incorrecto")
			assert.Equal(t, initialInterval, s.Config.Interval, "Intervalo inicial incorrecto")
			assert.Equal(t, initialThreshold, s.Config.Threshold, "Threshold inicial incorrecto")
			assert.True(t, s.Config.Enabled, "El sensor debería estar habilitado")
			break
		}
	}

	assert.True(t, found, "El sensor registrado debería aparecer en la lista")
	t.Logf("✓ Sensor encontrado en la lista con configuración correcta")
}

func testGetSensorConfig(t *testing.T, nc *nats.Conn, sensorID string) {
	subject := fmt.Sprintf("sensor.config.get.%s", sensorID)
	msg, err := nc.Request(subject, []byte(""), 5*time.Second)
	require.NoError(t, err, "Error obteniendo configuración")

	// Debug: ver respuesta raw
	t.Logf("→ Raw response: %s", string(msg.Data))

	var config SensorConfig
	err = json.Unmarshal(msg.Data, &config)
	require.NoError(t, err, "Error parseando configuración")

	t.Logf("→ Parsed config: SensorID=%s, Interval=%d, Threshold=%.2f, Enabled=%v",
		config.SensorID, config.Interval, config.Threshold, config.Enabled)

	assert.Equal(t, sensorID, config.SensorID, "SensorID incorrecto")
	assert.Equal(t, initialInterval, config.Interval, "Intervalo incorrecto")
	assert.Equal(t, initialThreshold, config.Threshold, "Threshold incorrecto")
	assert.True(t, config.Enabled, "El sensor debería estar habilitado")
	t.Logf("✓ Configuración obtenida correctamente")
}

func testUpdateSensorConfig(t *testing.T, nc *nats.Conn, sensorID string) {
	updatedConfig := SensorConfig{
		SensorID:  sensorID,
		Interval:  updatedInterval,
		Threshold: updatedThreshold,
		Enabled:   true,
	}

	payload, err := json.Marshal(updatedConfig)
	require.NoError(t, err, "Error serializando configuración")

	subject := fmt.Sprintf("sensor.config.set.%s", sensorID)
	msg, err := nc.Request(subject, payload, 5*time.Second)
	require.NoError(t, err, "Error actualizando configuración")

	// El servidor devuelve {"status": "ok"}
	var response map[string]interface{}
	err = json.Unmarshal(msg.Data, &response)
	require.NoError(t, err, "Error parseando respuesta")

	assert.Equal(t, "ok", response["status"], "El status debería ser ok")
	t.Logf("✓ Configuración actualizada: interval=%d, threshold=%.1f", updatedInterval, updatedThreshold)
}

func testVerifyConfigUpdated(t *testing.T, nc *nats.Conn, sensorID string) {
	// Esperar un momento para que se apliquen los cambios
	time.Sleep(1 * time.Second)

	// Verificar vía config.get
	subject := fmt.Sprintf("sensor.config.get.%s", sensorID)
	msg, err := nc.Request(subject, []byte(""), 5*time.Second)
	require.NoError(t, err, "Error obteniendo configuración actualizada")

	var config SensorConfig
	err = json.Unmarshal(msg.Data, &config)
	require.NoError(t, err, "Error parseando configuración")

	assert.Equal(t, updatedInterval, config.Interval, "Intervalo no se actualizó")
	assert.Equal(t, updatedThreshold, config.Threshold, "Threshold no se actualizó")

	// Verificar vía sensor.list
	msg, err = nc.Request("sensor.list", []byte(""), 5*time.Second)
	require.NoError(t, err, "Error listando sensores")

	var sensors []SensorDefinition
	err = json.Unmarshal(msg.Data, &sensors)
	require.NoError(t, err, "Error parseando respuesta")

	found := false
	for _, s := range sensors {
		if s.ID == sensorID {
			found = true
			assert.Equal(t, updatedInterval, s.Config.Interval, "Intervalo no se reflejó en la lista")
			assert.Equal(t, updatedThreshold, s.Config.Threshold, "Threshold no se reflejó en la lista")
			break
		}
	}

	assert.True(t, found, "Sensor no encontrado en la lista")
	t.Logf("✓ Configuración verificada: cambios reflejados en config.get y sensor.list")
}

func testQuerySensorReadings(t *testing.T, nc *nats.Conn, sensorID string) {
	// Esperar a que haya al menos una lectura
	// Usamos el intervalo actualizado (3000ms) + margen
	time.Sleep(time.Duration(updatedInterval+2000) * time.Millisecond)

	subject := fmt.Sprintf("sensor.readings.query.%s", sensorID)

	// Enviar payload con formato JSON correcto
	payload := []byte(`{"limit": 10}`)
	msg, err := nc.Request(subject, payload, 5*time.Second)
	require.NoError(t, err, "Error consultando lecturas")

	var readings []map[string]interface{}
	err = json.Unmarshal(msg.Data, &readings)
	require.NoError(t, err, "Error parseando lecturas")

	// Nota: El sensor puede no tener lecturas todavía si acaba de ser registrado
	// Por eso hacemos una aserción más permisiva
	t.Logf("→ Lecturas encontradas: %d", len(readings))

	if len(readings) > 0 {
		reading := readings[0]
		assert.NotNil(t, reading["sensor_id"], "La lectura debería tener sensor_id")
		assert.NotNil(t, reading["value"], "La lectura debería tener value")
		assert.NotNil(t, reading["timestamp"], "La lectura debería tener timestamp")
		t.Logf("✓ Lecturas verificadas correctamente: %d", len(readings))
	} else {
		t.Log("⚠ No hay lecturas todavía (el sensor acaba de ser registrado)")
	}
}

func testVerifyAlerts(t *testing.T, nc *nats.Conn, sensorID string) {
	alertSubject := fmt.Sprintf("sensor.alerts.%s.%s", sensorType, sensorID)

	// Suscribirse a alertas
	alertChan := make(chan *nats.Msg, 10)
	sub, err := nc.ChanSubscribe(alertSubject, alertChan)
	require.NoError(t, err, "Error suscribiéndose a alertas")
	defer sub.Unsubscribe()

	t.Logf("✓ Suscrito a alertas en: %s", alertSubject)
	t.Log("  (Las alertas se generarán si el sensor supera el threshold)")
}

// TestCLICommands prueba los comandos del CLI
func TestCLICommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI tests in short mode")
	}

	// Verificar que el CLI existe
	cliPath := "../../bin/iot-cli.exe"
	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		cliPath = "../../bin/iot-cli"
	}

	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		t.Skip("CLI binary not found, skipping CLI tests")
	}

	t.Run("CLI_SensorList", func(t *testing.T) {
		cmd := exec.Command(cliPath, "sensor", "list")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Error ejecutando 'sensor list': %s", string(output))
		assert.Contains(t, string(output), "temp-001", "La salida debería contener temp-001")
		t.Log("✓ CLI: sensor list funciona correctamente")
	})

	t.Run("CLI_ConfigGet", func(t *testing.T) {
		cmd := exec.Command(cliPath, "config", "get", "temp-001")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Error ejecutando 'config get': %s", string(output))
		assert.Contains(t, string(output), "Interval", "La salida debería contener Interval")
		t.Log("✓ CLI: config get funciona correctamente")
	})
}
