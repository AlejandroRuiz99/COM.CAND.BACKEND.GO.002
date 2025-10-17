package nats

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alejandro/technical_test_uvigo/internal/sensor"
)

// MockRepository para testing de handlers
type MockRepository struct {
	configs  map[string]*sensor.SensorConfig
	readings map[string][]*sensor.SensorReading
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		configs:  make(map[string]*sensor.SensorConfig),
		readings: make(map[string][]*sensor.SensorReading),
	}
}

func (m *MockRepository) SaveReading(ctx context.Context, reading *sensor.SensorReading) error {
	m.readings[reading.SensorID] = append(m.readings[reading.SensorID], reading)
	return nil
}

func (m *MockRepository) GetLatestReadings(ctx context.Context, sensorID string, limit int) ([]*sensor.SensorReading, error) {
	readings := m.readings[sensorID]
	if len(readings) > limit {
		return readings[:limit], nil
	}
	return readings, nil
}

func (m *MockRepository) GetReadingsByTimeRange(ctx context.Context, sensorID string, start, end time.Time) ([]*sensor.SensorReading, error) {
	return m.readings[sensorID], nil
}

func (m *MockRepository) SaveConfig(ctx context.Context, config *sensor.SensorConfig) error {
	m.configs[config.SensorID] = config
	return nil
}

func (m *MockRepository) GetConfig(ctx context.Context, sensorID string) (*sensor.SensorConfig, error) {
	config, exists := m.configs[sensorID]
	if !exists {
		return nil, nil
	}
	return config, nil
}

func (m *MockRepository) Close() error {
	return nil
}

func TestHandler_ConfigGet(t *testing.T) {
	_, url := setupTestNATS(t)

	client, err := NewClient(url)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	defer client.Close()

	// Setup mock repository con config de prueba
	repo := NewMockRepository()
	testConfig := &sensor.SensorConfig{
		SensorID:  "temp-001",
		Interval:  1000,
		Threshold: 30.0,
		Enabled:   true,
	}
	repo.SaveConfig(context.Background(), testConfig)

	// Crear handler
	handler := NewHandler(client, repo)

	// Iniciar handler de config GET
	err = handler.HandleConfigRequests()
	if err != nil {
		t.Fatalf("HandleConfigRequests() failed: %v", err)
	}

	// Dar tiempo a que se active la suscripción
	time.Sleep(100 * time.Millisecond)

	// Request de configuración
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	subject := ConfigGetSubject("temp-001")
	response, err := client.Request(ctx, subject, nil)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	// Parsear respuesta
	var gotConfig sensor.SensorConfig
	if err := json.Unmarshal(response.Data, &gotConfig); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verificar
	if gotConfig.SensorID != testConfig.SensorID {
		t.Errorf("got sensor_id %s, want %s", gotConfig.SensorID, testConfig.SensorID)
	}
	if gotConfig.Interval != testConfig.Interval {
		t.Errorf("got interval %d, want %d", gotConfig.Interval, testConfig.Interval)
	}
	if gotConfig.Threshold != testConfig.Threshold {
		t.Errorf("got threshold %f, want %f", gotConfig.Threshold, testConfig.Threshold)
	}
}

func TestHandler_ConfigSet(t *testing.T) {
	_, url := setupTestNATS(t)

	client, err := NewClient(url)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	defer client.Close()

	repo := NewMockRepository()
	handler := NewHandler(client, repo)

	// Iniciar handler de config SET
	err = handler.HandleConfigRequests()
	if err != nil {
		t.Fatalf("HandleConfigRequests() failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Nueva configuración
	newConfig := &sensor.SensorConfig{
		SensorID:  "temp-001",
		Interval:  500,
		Threshold: 25.0,
		Enabled:   true,
	}

	configData, err := json.Marshal(newConfig)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	// Request para actualizar config
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	subject := ConfigSetSubject("temp-001")
	response, err := client.Request(ctx, subject, configData)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	// Verificar respuesta de éxito
	var result map[string]string
	if err := json.Unmarshal(response.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %s", result["status"])
	}

	// Verificar que se guardó en el repo
	savedConfig, err := repo.GetConfig(context.Background(), "temp-001")
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	if savedConfig == nil {
		t.Fatal("config was not saved")
	}

	if savedConfig.Interval != 500 {
		t.Errorf("got interval %d, want 500", savedConfig.Interval)
	}
}

func TestHandler_ConfigGetNotFound(t *testing.T) {
	_, url := setupTestNATS(t)

	client, err := NewClient(url)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	defer client.Close()

	repo := NewMockRepository()
	handler := NewHandler(client, repo)

	err = handler.HandleConfigRequests()
	if err != nil {
		t.Fatalf("HandleConfigRequests() failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Request de sensor que no existe
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	subject := ConfigGetSubject("nonexistent")
	response, err := client.Request(ctx, subject, nil)
	if err != nil {
		t.Fatalf("Request() failed: %v", err)
	}

	// Debería retornar error
	var result map[string]string
	if err := json.Unmarshal(response.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["error"] == "" {
		t.Error("expected error in response for nonexistent sensor")
	}
}
