package simulator

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/alejandro/technical_test_uvigo/internal/config"
	natsclient "github.com/alejandro/technical_test_uvigo/internal/nats"
	"github.com/alejandro/technical_test_uvigo/internal/repository"
	"github.com/alejandro/technical_test_uvigo/internal/sensor"
)

// mockRepository para testing (thread-safe)
type mockRepository struct {
	configs  map[string]*sensor.SensorConfig
	readings []*sensor.SensorReading
	mu       sync.Mutex
}

var _ repository.Repository = (*mockRepository)(nil)

func newMockRepository() *mockRepository {
	return &mockRepository{
		configs:  make(map[string]*sensor.SensorConfig),
		readings: make([]*sensor.SensorReading, 0),
	}
}

func (m *mockRepository) SaveReading(ctx context.Context, reading *sensor.SensorReading) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readings = append(m.readings, reading)
	return nil
}

func (m *mockRepository) GetLatestReadings(ctx context.Context, sensorID string, limit int) ([]*sensor.SensorReading, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.readings, nil
}

func (m *mockRepository) GetReadingsByTimeRange(ctx context.Context, sensorID string, start, end time.Time) ([]*sensor.SensorReading, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.readings, nil
}

func (m *mockRepository) SaveConfig(ctx context.Context, config *sensor.SensorConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configs[config.SensorID] = config
	return nil
}

func (m *mockRepository) GetConfig(ctx context.Context, sensorID string) (*sensor.SensorConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cfg, ok := m.configs[sensorID]; ok {
		return cfg, nil
	}
	return nil, nil
}

func (m *mockRepository) Close() error {
	return nil
}

// mockNATSClient para testing (thread-safe)
type mockNATSClient struct {
	published []string
	mu        sync.Mutex
}

// Asegurar que mockNATSClient implementa natsclient.Publisher
var _ natsclient.Publisher = (*mockNATSClient)(nil)

func (m *mockNATSClient) Publish(subject string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = append(m.published, subject)
	return nil
}

func (m *mockNATSClient) IsConnected() bool {
	return true
}

func (m *mockNATSClient) Close() error {
	return nil
}

func TestNew(t *testing.T) {
	repo := newMockRepository()
	natsClient := &mockNATSClient{}

	sim := New(repo, natsClient)

	if sim == nil {
		t.Fatal("New() returned nil")
	}

	if sim.sensors == nil {
		t.Error("sensors map not initialized")
	}

	if sim.repo == nil {
		t.Error("repository not set")
	}

	if sim.natsClient == nil {
		t.Error("natsClient not set")
	}
}

func TestAddSensor(t *testing.T) {
	repo := newMockRepository()
	natsClient := &mockNATSClient{}
	sim := New(repo, natsClient)

	sensorDef := config.SensorDef{
		ID:       "test-001",
		Type:     sensor.SensorTypeTemperature,
		Name:     "Test Sensor",
		Location: "test-lab",
		Config: sensor.SensorConfig{
			SensorID:  "test-001",
			Interval:  1000,
			Threshold: 30.0,
			Enabled:   true,
		},
	}

	err := sim.AddSensor(sensorDef)
	if err != nil {
		t.Fatalf("AddSensor() failed: %v", err)
	}

	if sim.GetSensorCount() != 1 {
		t.Errorf("Expected sensor count 1, got %d", sim.GetSensorCount())
	}

	// Verificar que la config se guardó en el repo
	if _, ok := repo.configs["test-001"]; !ok {
		t.Error("Config not saved in repository")
	}
}

func TestAddSensor_Duplicate(t *testing.T) {
	repo := newMockRepository()
	natsClient := &mockNATSClient{}
	sim := New(repo, natsClient)

	sensorDef := config.SensorDef{
		ID:   "test-001",
		Type: sensor.SensorTypeTemperature,
		Name: "Test Sensor",
		Config: sensor.SensorConfig{
			SensorID:  "test-001",
			Interval:  1000,
			Threshold: 30.0,
			Enabled:   true,
		},
	}

	// Añadir primera vez
	err := sim.AddSensor(sensorDef)
	if err != nil {
		t.Fatalf("First AddSensor() failed: %v", err)
	}

	// Intentar añadir duplicado
	err = sim.AddSensor(sensorDef)
	if err == nil {
		t.Error("Expected error when adding duplicate sensor, got nil")
	}
}

func TestRemoveSensor(t *testing.T) {
	repo := newMockRepository()
	natsClient := &mockNATSClient{}
	sim := New(repo, natsClient)

	sensorDef := config.SensorDef{
		ID:   "test-001",
		Type: sensor.SensorTypeTemperature,
		Name: "Test Sensor",
		Config: sensor.SensorConfig{
			SensorID:  "test-001",
			Interval:  1000,
			Threshold: 30.0,
			Enabled:   true,
		},
	}

	// Añadir sensor
	sim.AddSensor(sensorDef)

	// Remover sensor
	err := sim.RemoveSensor("test-001")
	if err != nil {
		t.Fatalf("RemoveSensor() failed: %v", err)
	}

	if sim.GetSensorCount() != 0 {
		t.Errorf("Expected sensor count 0, got %d", sim.GetSensorCount())
	}
}

func TestRemoveSensor_NotFound(t *testing.T) {
	repo := newMockRepository()
	natsClient := &mockNATSClient{}
	sim := New(repo, natsClient)

	err := sim.RemoveSensor("nonexistent")
	if err == nil {
		t.Error("Expected error when removing nonexistent sensor, got nil")
	}
}

func TestUpdateSensorConfig(t *testing.T) {
	repo := newMockRepository()
	natsClient := &mockNATSClient{}
	sim := New(repo, natsClient)

	sensorDef := config.SensorDef{
		ID:   "test-001",
		Type: sensor.SensorTypeTemperature,
		Name: "Test Sensor",
		Config: sensor.SensorConfig{
			SensorID:  "test-001",
			Interval:  1000,
			Threshold: 30.0,
			Enabled:   true,
		},
	}

	sim.AddSensor(sensorDef)

	// Actualizar config
	newConfig := &sensor.SensorConfig{
		SensorID:  "test-001",
		Interval:  2000,
		Threshold: 35.0,
		Enabled:   true,
	}

	err := sim.UpdateSensorConfig("test-001", newConfig)
	if err != nil {
		t.Fatalf("UpdateSensorConfig() failed: %v", err)
	}

	// Verificar que se actualizó en el repo
	savedConfig := repo.configs["test-001"]
	if savedConfig.Interval != 2000 {
		t.Errorf("Expected interval 2000, got %d", savedConfig.Interval)
	}
	if savedConfig.Threshold != 35.0 {
		t.Errorf("Expected threshold 35.0, got %.2f", savedConfig.Threshold)
	}
}

func TestGenerateValue(t *testing.T) {
	repo := newMockRepository()
	natsClient := &mockNATSClient{}
	sim := New(repo, natsClient)

	tests := []struct {
		name       string
		sensorType sensor.SensorType
		minValue   float64
		maxValue   float64
	}{
		{
			name:       "temperature",
			sensorType: sensor.SensorTypeTemperature,
			minValue:   10.0,
			maxValue:   40.0,
		},
		{
			name:       "humidity",
			sensorType: sensor.SensorTypeHumidity,
			minValue:   0.0,
			maxValue:   100.0,
		},
		{
			name:       "pressure",
			sensorType: sensor.SensorTypePressure,
			minValue:   950.0,
			maxValue:   1050.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &sensorState{
				def: config.SensorDef{
					Type: tt.sensorType,
				},
				rand: rand.New(rand.NewSource(time.Now().UnixNano())),
			}

			// Generar 10 valores y verificar que están en rango
			for i := 0; i < 10; i++ {
				value := sim.generateValue(state)
				if value < tt.minValue || value > tt.maxValue {
					t.Errorf("Value %.2f out of range [%.2f, %.2f]", value, tt.minValue, tt.maxValue)
				}
			}
		})
	}
}

func TestGetUnit(t *testing.T) {
	repo := newMockRepository()
	natsClient := &mockNATSClient{}
	sim := New(repo, natsClient)

	tests := []struct {
		sensorType   sensor.SensorType
		expectedUnit string
	}{
		{sensor.SensorTypeTemperature, "°C"},
		{sensor.SensorTypeHumidity, "%"},
		{sensor.SensorTypePressure, "hPa"},
	}

	for _, tt := range tests {
		t.Run(string(tt.sensorType), func(t *testing.T) {
			unit := sim.getUnit(tt.sensorType)
			if unit != tt.expectedUnit {
				t.Errorf("Expected unit '%s', got '%s'", tt.expectedUnit, unit)
			}
		})
	}
}

func TestListSensors(t *testing.T) {
	repo := newMockRepository()
	natsClient := &mockNATSClient{}
	sim := New(repo, natsClient)

	// Añadir varios sensores
	sensors := []string{"temp-001", "hum-001", "press-001"}
	for _, id := range sensors {
		sensorDef := config.SensorDef{
			ID:   id,
			Type: sensor.SensorTypeTemperature,
			Name: "Test Sensor",
			Config: sensor.SensorConfig{
				SensorID:  id,
				Interval:  1000,
				Threshold: 30.0,
				Enabled:   true,
			},
		}
		sim.AddSensor(sensorDef)
	}

	list := sim.ListSensors()
	if len(list) != 3 {
		t.Errorf("Expected 3 sensors, got %d", len(list))
	}

	// Verificar que todos los sensores están en la lista
	sensorMap := make(map[string]bool)
	for _, id := range list {
		sensorMap[id] = true
	}

	for _, id := range sensors {
		if !sensorMap[id] {
			t.Errorf("Sensor %s not found in list", id)
		}
	}
}

func TestGetSensorCount(t *testing.T) {
	repo := newMockRepository()
	natsClient := &mockNATSClient{}
	sim := New(repo, natsClient)

	if sim.GetSensorCount() != 0 {
		t.Errorf("Expected initial count 0, got %d", sim.GetSensorCount())
	}

	// Añadir sensores
	for i := 1; i <= 5; i++ {
		sensorDef := config.SensorDef{
			ID:   fmt.Sprintf("sensor-%d", i),
			Type: sensor.SensorTypeTemperature,
			Name: "Test Sensor",
			Config: sensor.SensorConfig{
				SensorID:  fmt.Sprintf("sensor-%d", i),
				Interval:  1000,
				Threshold: 30.0,
				Enabled:   true,
			},
		}
		sim.AddSensor(sensorDef)

		if sim.GetSensorCount() != i {
			t.Errorf("Expected count %d, got %d", i, sim.GetSensorCount())
		}
	}
}

func TestStop(t *testing.T) {
	repo := newMockRepository()
	natsClient := &mockNATSClient{}
	sim := New(repo, natsClient)

	sensorDef := config.SensorDef{
		ID:   "test-001",
		Type: sensor.SensorTypeTemperature,
		Name: "Test Sensor",
		Config: sensor.SensorConfig{
			SensorID:  "test-001",
			Interval:  100, // Intervalo corto para test
			Threshold: 30.0,
			Enabled:   true,
		},
	}

	// AddSensor ya arranca la goroutine automáticamente
	sim.AddSensor(sensorDef)

	// Esperar un poco para que genere algunas lecturas
	time.Sleep(250 * time.Millisecond)

	// Detener
	sim.Stop()

	// Verificar que se generaron lecturas
	if len(repo.readings) == 0 {
		t.Error("No readings were generated")
	}

	t.Logf("Generated %d readings before stop", len(repo.readings))
}

func TestConcurrentSensors(t *testing.T) {
	repo := newMockRepository()
	natsClient := &mockNATSClient{}
	sim := New(repo, natsClient)

	// Añadir 3 sensores con diferentes intervalos
	sensors := []config.SensorDef{
		{
			ID:   "temp-001",
			Type: sensor.SensorTypeTemperature,
			Name: "Temperature Sensor",
			Config: sensor.SensorConfig{
				SensorID:  "temp-001",
				Interval:  50, // Rápido
				Threshold: 30.0,
				Enabled:   true,
			},
		},
		{
			ID:   "hum-001",
			Type: sensor.SensorTypeHumidity,
			Name: "Humidity Sensor",
			Config: sensor.SensorConfig{
				SensorID:  "hum-001",
				Interval:  100, // Medio
				Threshold: 70.0,
				Enabled:   true,
			},
		},
		{
			ID:   "press-001",
			Type: sensor.SensorTypePressure,
			Name: "Pressure Sensor",
			Config: sensor.SensorConfig{
				SensorID:  "press-001",
				Interval:  200, // Lento
				Threshold: 1020.0,
				Enabled:   true,
			},
		},
	}

	// Añadir todos los sensores (cada uno arranca su goroutine)
	for _, s := range sensors {
		if err := sim.AddSensor(s); err != nil {
			t.Fatalf("Failed to add sensor %s: %v", s.ID, err)
		}
	}

	// Esperar para que todos generen lecturas
	time.Sleep(300 * time.Millisecond)

	// Detener todas las goroutines
	sim.Stop()

	// Verificar que todos los sensores generaron lecturas
	tempReadings := 0
	humReadings := 0
	pressReadings := 0

	for _, r := range repo.readings {
		switch r.SensorID {
		case "temp-001":
			tempReadings++
		case "hum-001":
			humReadings++
		case "press-001":
			pressReadings++
		}
	}

	t.Logf("Readings: temp=%d, hum=%d, press=%d", tempReadings, humReadings, pressReadings)

	// temp-001 debería tener más lecturas (intervalo más corto)
	if tempReadings == 0 {
		t.Error("Temperature sensor generated no readings")
	}
	if humReadings == 0 {
		t.Error("Humidity sensor generated no readings")
	}
	if pressReadings == 0 {
		t.Error("Pressure sensor generated no readings")
	}

	// Verificar orden esperado (más rápido = más lecturas)
	if tempReadings <= pressReadings {
		t.Error("Expected temp sensor to generate more readings than pressure sensor")
	}
}
