package storage

import (
	"context"
	"testing"
	"time"

	"github.com/alejandro/technical_test_uvigo/internal/sensor"
)

func TestSQLiteRepository_SaveAndGetReading(t *testing.T) {
	// Setup: DB en memoria
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// Crear lectura de prueba
	reading := &sensor.SensorReading{
		ID:        "read-001",
		SensorID:  "temp-001",
		Type:      sensor.SensorTypeTemperature,
		Value:     23.5,
		Unit:      "°C",
		Timestamp: time.Now().UTC(),
	}

	// Test: Guardar lectura
	err = repo.SaveReading(ctx, reading)
	if err != nil {
		t.Fatalf("SaveReading failed: %v", err)
	}

	// Test: Recuperar lectura
	readings, err := repo.GetLatestReadings(ctx, "temp-001", 10)
	if err != nil {
		t.Fatalf("GetLatestReadings failed: %v", err)
	}

	// Verificaciones
	if len(readings) != 1 {
		t.Fatalf("expected 1 reading, got %d", len(readings))
	}

	if readings[0].ID != reading.ID {
		t.Errorf("expected ID %s, got %s", reading.ID, readings[0].ID)
	}

	if readings[0].Value != reading.Value {
		t.Errorf("expected value %f, got %f", reading.Value, readings[0].Value)
	}

	if readings[0].SensorID != reading.SensorID {
		t.Errorf("expected sensor_id %s, got %s", reading.SensorID, readings[0].SensorID)
	}
}

func TestSQLiteRepository_GetLatestReadings_OrderedByTime(t *testing.T) {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// Insertar múltiples lecturas en orden no temporal
	baseTime := time.Now().UTC()
	readings := []*sensor.SensorReading{
		{
			ID:        "read-001",
			SensorID:  "temp-001",
			Type:      sensor.SensorTypeTemperature,
			Value:     20.0,
			Unit:      "°C",
			Timestamp: baseTime.Add(-3 * time.Hour),
		},
		{
			ID:        "read-002",
			SensorID:  "temp-001",
			Type:      sensor.SensorTypeTemperature,
			Value:     25.0,
			Unit:      "°C",
			Timestamp: baseTime.Add(-1 * time.Hour),
		},
		{
			ID:        "read-003",
			SensorID:  "temp-001",
			Type:      sensor.SensorTypeTemperature,
			Value:     22.0,
			Unit:      "°C",
			Timestamp: baseTime.Add(-2 * time.Hour),
		},
	}

	for _, r := range readings {
		if err := repo.SaveReading(ctx, r); err != nil {
			t.Fatalf("SaveReading failed: %v", err)
		}
	}

	// Obtener últimas 2 lecturas
	latest, err := repo.GetLatestReadings(ctx, "temp-001", 2)
	if err != nil {
		t.Fatalf("GetLatestReadings failed: %v", err)
	}

	if len(latest) != 2 {
		t.Fatalf("expected 2 readings, got %d", len(latest))
	}

	// Verificar orden descendente (más reciente primero)
	if latest[0].ID != "read-002" {
		t.Errorf("expected first reading to be read-002, got %s", latest[0].ID)
	}

	if latest[1].ID != "read-003" {
		t.Errorf("expected second reading to be read-003, got %s", latest[1].ID)
	}
}

func TestSQLiteRepository_SaveReadingWithError(t *testing.T) {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	errorMsg := "sensor timeout"
	reading := &sensor.SensorReading{
		ID:        "read-err-001",
		SensorID:  "temp-001",
		Type:      sensor.SensorTypeTemperature,
		Value:     0,
		Unit:      "°C",
		Error:     &errorMsg,
		Timestamp: time.Now().UTC(),
	}

	// Guardar lectura con error
	err = repo.SaveReading(ctx, reading)
	if err != nil {
		t.Fatalf("SaveReading failed: %v", err)
	}

	// Recuperar y verificar
	readings, err := repo.GetLatestReadings(ctx, "temp-001", 1)
	if err != nil {
		t.Fatalf("GetLatestReadings failed: %v", err)
	}

	if len(readings) != 1 {
		t.Fatalf("expected 1 reading, got %d", len(readings))
	}

	if readings[0].Error == nil {
		t.Error("expected error field to be populated")
	} else if *readings[0].Error != errorMsg {
		t.Errorf("expected error '%s', got '%s'", errorMsg, *readings[0].Error)
	}
}

func TestSQLiteRepository_SaveAndGetConfig(t *testing.T) {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// Crear config
	config := &sensor.SensorConfig{
		SensorID:  "temp-001",
		Interval:  1000,
		Threshold: 30.0,
		Enabled:   true,
	}

	// Guardar config
	err = repo.SaveConfig(ctx, config)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Recuperar config
	retrieved, err := repo.GetConfig(ctx, "temp-001")
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	// Verificaciones
	if retrieved.SensorID != config.SensorID {
		t.Errorf("expected sensor_id %s, got %s", config.SensorID, retrieved.SensorID)
	}

	if retrieved.Interval != config.Interval {
		t.Errorf("expected interval %d, got %d", config.Interval, retrieved.Interval)
	}

	if retrieved.Threshold != config.Threshold {
		t.Errorf("expected threshold %f, got %f", config.Threshold, retrieved.Threshold)
	}

	if retrieved.Enabled != config.Enabled {
		t.Errorf("expected enabled %v, got %v", config.Enabled, retrieved.Enabled)
	}
}

func TestSQLiteRepository_UpdateConfig(t *testing.T) {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// Config inicial
	config := &sensor.SensorConfig{
		SensorID:  "temp-001",
		Interval:  1000,
		Threshold: 30.0,
		Enabled:   true,
	}

	err = repo.SaveConfig(ctx, config)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Actualizar config
	config.Interval = 500
	config.Threshold = 25.0

	err = repo.SaveConfig(ctx, config)
	if err != nil {
		t.Fatalf("SaveConfig (update) failed: %v", err)
	}

	// Verificar actualización
	retrieved, err := repo.GetConfig(ctx, "temp-001")
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	if retrieved.Interval != 500 {
		t.Errorf("expected updated interval 500, got %d", retrieved.Interval)
	}

	if retrieved.Threshold != 25.0 {
		t.Errorf("expected updated threshold 25.0, got %f", retrieved.Threshold)
	}
}

func TestSQLiteRepository_GetReadingsByTimeRange(t *testing.T) {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// Crear lecturas en diferentes tiempos
	baseTime := time.Date(2025, 10, 16, 12, 0, 0, 0, time.UTC)
	readings := []*sensor.SensorReading{
		{
			ID:        "read-001",
			SensorID:  "temp-001",
			Type:      sensor.SensorTypeTemperature,
			Value:     20.0,
			Unit:      "°C",
			Timestamp: baseTime.Add(-3 * time.Hour), // 09:00
		},
		{
			ID:        "read-002",
			SensorID:  "temp-001",
			Type:      sensor.SensorTypeTemperature,
			Value:     25.0,
			Unit:      "°C",
			Timestamp: baseTime.Add(-1 * time.Hour), // 11:00
		},
		{
			ID:        "read-003",
			SensorID:  "temp-001",
			Type:      sensor.SensorTypeTemperature,
			Value:     30.0,
			Unit:      "°C",
			Timestamp: baseTime, // 12:00
		},
	}

	for _, r := range readings {
		if err := repo.SaveReading(ctx, r); err != nil {
			t.Fatalf("SaveReading failed: %v", err)
		}
	}

	// Buscar en rango 10:00 - 12:30
	start := baseTime.Add(-2 * time.Hour)
	end := baseTime.Add(30 * time.Minute)

	rangeReadings, err := repo.GetReadingsByTimeRange(ctx, "temp-001", start, end)
	if err != nil {
		t.Fatalf("GetReadingsByTimeRange failed: %v", err)
	}

	// Debe devolver read-002 y read-003 (dentro del rango)
	if len(rangeReadings) != 2 {
		t.Fatalf("expected 2 readings in range, got %d", len(rangeReadings))
	}

	// Verificar que son las correctas
	ids := make(map[string]bool)
	for _, r := range rangeReadings {
		ids[r.ID] = true
	}

	if !ids["read-002"] || !ids["read-003"] {
		t.Error("expected read-002 and read-003 in time range")
	}
}
