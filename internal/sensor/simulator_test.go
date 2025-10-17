package sensor

import (
	"context"
	"testing"
	"time"
)

func TestSimulator_GenerateReading(t *testing.T) {
	config := &SensorConfig{
		SensorID:  "temp-001",
		Interval:  1000,
		Threshold: 30.0,
		Enabled:   true,
	}

	sim := NewSimulator("temp-001", SensorTypeTemperature, config)

	// Generar lectura
	reading := sim.GenerateReading()

	// Verificaciones básicas
	if reading.SensorID != "temp-001" {
		t.Errorf("expected sensor_id temp-001, got %s", reading.SensorID)
	}

	if reading.Type != SensorTypeTemperature {
		t.Errorf("expected type temperature, got %s", reading.Type)
	}

	if reading.Unit == "" {
		t.Error("unit should not be empty")
	}

	if reading.Timestamp.IsZero() {
		t.Error("timestamp should be set")
	}

	// Verificar que el valor está en rango razonable para temperatura
	if reading.Error == nil && (reading.Value < -50 || reading.Value > 100) {
		t.Errorf("temperature value %f out of reasonable range", reading.Value)
	}
}

func TestSimulator_GenerateReading_DifferentTypes(t *testing.T) {
	tests := []struct {
		name         string
		sensorType   SensorType
		expectedUnit string
		minValue     float64
		maxValue     float64
	}{
		{
			name:         "temperature",
			sensorType:   SensorTypeTemperature,
			expectedUnit: "°C",
			minValue:     -50,
			maxValue:     100,
		},
		{
			name:         "humidity",
			sensorType:   SensorTypeHumidity,
			expectedUnit: "%",
			minValue:     0,
			maxValue:     100,
		},
		{
			name:         "pressure",
			sensorType:   SensorTypePressure,
			expectedUnit: "hPa",
			minValue:     800,
			maxValue:     1200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &SensorConfig{
				SensorID:  "test-001",
				Interval:  1000,
				Threshold: 50.0,
				Enabled:   true,
			}

			sim := NewSimulator("test-001", tt.sensorType, config)
			reading := sim.GenerateReading()

			if reading.Unit != tt.expectedUnit {
				t.Errorf("expected unit %s, got %s", tt.expectedUnit, reading.Unit)
			}

			// Verificar que el valor está en rango si no hay error
			if reading.Error == nil {
				if reading.Value < tt.minValue || reading.Value > tt.maxValue {
					t.Errorf("value %f out of range [%f, %f]", reading.Value, tt.minValue, tt.maxValue)
				}
			}
		})
	}
}

func TestSimulator_GenerateReading_WithErrors(t *testing.T) {
	config := &SensorConfig{
		SensorID:  "temp-001",
		Interval:  1000,
		Threshold: 30.0,
		Enabled:   true,
	}

	sim := NewSimulator("temp-001", SensorTypeTemperature, config)

	// Generar muchas lecturas para verificar que a veces hay errores
	errorCount := 0
	successCount := 0
	iterations := 1000

	for i := 0; i < iterations; i++ {
		reading := sim.GenerateReading()
		if reading.Error != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// Verificar que hay errores (probabilidad ~5%, con 1000 iteraciones deberíamos ver algunos)
	if errorCount == 0 {
		t.Error("expected some error readings, got none")
	}

	// Verificar que la mayoría son exitosas
	if successCount < 900 {
		t.Errorf("expected at least 900 successful readings, got %d", successCount)
	}

	t.Logf("Error rate: %.2f%% (%d/%d)", float64(errorCount)/float64(iterations)*100, errorCount, iterations)
}

func TestSimulator_Run(t *testing.T) {
	config := &SensorConfig{
		SensorID:  "temp-001",
		Interval:  100, // 100ms para test rápido
		Threshold: 30.0,
		Enabled:   true,
	}

	// Canal para recibir lecturas
	readingsCh := make(chan *SensorReading, 10)

	// Callback que simula publicación
	callback := func(reading *SensorReading) error {
		readingsCh <- reading
		return nil
	}

	sim := NewSimulator("temp-001", SensorTypeTemperature, config)

	// Ejecutar simulador con context con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Ejecutar en goroutine
	go sim.Run(ctx, callback)

	// Debería generar al menos 3 lecturas en 500ms (interval 100ms)
	readingsReceived := 0
	timeout := time.After(600 * time.Millisecond)

	for {
		select {
		case reading := <-readingsCh:
			readingsReceived++
			if reading.SensorID != "temp-001" {
				t.Errorf("unexpected sensor_id: %s", reading.SensorID)
			}
		case <-timeout:
			if readingsReceived < 3 {
				t.Errorf("expected at least 3 readings, got %d", readingsReceived)
			}
			return
		}
	}
}

func TestSimulator_Run_CancellationStops(t *testing.T) {
	config := &SensorConfig{
		SensorID:  "temp-001",
		Interval:  100,
		Threshold: 30.0,
		Enabled:   true,
	}

	readingsCh := make(chan *SensorReading, 10)
	callback := func(reading *SensorReading) error {
		readingsCh <- reading
		return nil
	}

	sim := NewSimulator("temp-001", SensorTypeTemperature, config)

	ctx, cancel := context.WithCancel(context.Background())

	go sim.Run(ctx, callback)

	// Esperar algunas lecturas
	time.Sleep(250 * time.Millisecond)

	// Vaciar el canal antes de cancelar
	for len(readingsCh) > 0 {
		<-readingsCh
	}

	// Cancelar context
	cancel()

	// Esperar para asegurar que el simulador se detuvo
	time.Sleep(300 * time.Millisecond)

	// Verificar que NO llegaron más lecturas después de cancelar
	if len(readingsCh) > 0 {
		t.Errorf("simulator should have stopped, but received %d readings after cancellation", len(readingsCh))
	}
}

func TestSimulator_UpdateConfig(t *testing.T) {
	initialConfig := &SensorConfig{
		SensorID:  "temp-001",
		Interval:  1000,
		Threshold: 30.0,
		Enabled:   true,
	}

	sim := NewSimulator("temp-001", SensorTypeTemperature, initialConfig)

	// Verificar config inicial
	if sim.GetConfig().Interval != 1000 {
		t.Errorf("expected initial interval 1000, got %d", sim.GetConfig().Interval)
	}

	// Actualizar config
	newConfig := &SensorConfig{
		SensorID:  "temp-001",
		Interval:  500,
		Threshold: 25.0,
		Enabled:   true,
	}

	sim.UpdateConfig(newConfig)

	// Verificar que se actualizó
	if sim.GetConfig().Interval != 500 {
		t.Errorf("expected updated interval 500, got %d", sim.GetConfig().Interval)
	}

	if sim.GetConfig().Threshold != 25.0 {
		t.Errorf("expected updated threshold 25.0, got %f", sim.GetConfig().Threshold)
	}
}
