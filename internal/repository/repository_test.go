package repository

import (
	"context"
	"testing"
	"time"

	"github.com/alejandro/technical_test_uvigo/internal/sensor"
	"github.com/alejandro/technical_test_uvigo/internal/storage"
)

// TestRepositoryInterface verifica que todas las implementaciones cumplen la interfaz
func TestRepositoryInterface(t *testing.T) {
	var _ Repository = (*storage.SQLiteRepository)(nil)
}

// RepositoryContractTests son tests de contrato que cualquier implementación debe pasar
func RepositoryContractTests(t *testing.T, repo Repository) {
	ctx := context.Background()

	t.Run("SaveAndGetConfig", func(t *testing.T) {
		config := &sensor.SensorConfig{
			SensorID:  "test-001",
			Interval:  5000,
			Threshold: 30.0,
			Enabled:   true,
		}

		// Guardar config
		err := repo.SaveConfig(ctx, config)
		if err != nil {
			t.Fatalf("SaveConfig() failed: %v", err)
		}

		// Obtener config
		retrieved, err := repo.GetConfig(ctx, "test-001")
		if err != nil {
			t.Fatalf("GetConfig() failed: %v", err)
		}

		if retrieved == nil {
			t.Fatal("GetConfig() returned nil")
		}

		// Verificar valores
		if retrieved.SensorID != config.SensorID {
			t.Errorf("Expected SensorID %s, got %s", config.SensorID, retrieved.SensorID)
		}
		if retrieved.Interval != config.Interval {
			t.Errorf("Expected Interval %d, got %d", config.Interval, retrieved.Interval)
		}
		if retrieved.Threshold != config.Threshold {
			t.Errorf("Expected Threshold %.2f, got %.2f", config.Threshold, retrieved.Threshold)
		}
		if retrieved.Enabled != config.Enabled {
			t.Errorf("Expected Enabled %v, got %v", config.Enabled, retrieved.Enabled)
		}
	})

	t.Run("UpdateConfig", func(t *testing.T) {
		// Guardar config inicial
		config := &sensor.SensorConfig{
			SensorID:  "test-002",
			Interval:  5000,
			Threshold: 30.0,
			Enabled:   true,
		}
		repo.SaveConfig(ctx, config)

		// Actualizar
		updatedConfig := &sensor.SensorConfig{
			SensorID:  "test-002",
			Interval:  10000,
			Threshold: 35.0,
			Enabled:   false,
		}
		err := repo.SaveConfig(ctx, updatedConfig)
		if err != nil {
			t.Fatalf("SaveConfig() update failed: %v", err)
		}

		// Verificar actualización
		retrieved, err := repo.GetConfig(ctx, "test-002")
		if err != nil {
			t.Fatalf("GetConfig() failed: %v", err)
		}

		if retrieved.Interval != 10000 {
			t.Errorf("Expected updated Interval 10000, got %d", retrieved.Interval)
		}
		if retrieved.Threshold != 35.0 {
			t.Errorf("Expected updated Threshold 35.0, got %.2f", retrieved.Threshold)
		}
		if retrieved.Enabled != false {
			t.Errorf("Expected updated Enabled false, got %v", retrieved.Enabled)
		}
	})

	t.Run("SaveAndGetReading", func(t *testing.T) {
		reading := &sensor.SensorReading{
			ID:        "reading-001",
			SensorID:  "test-003",
			Type:      sensor.SensorTypeTemperature,
			Value:     25.5,
			Unit:      "°C",
			Timestamp: time.Now().UTC(),
		}

		// Guardar lectura
		err := repo.SaveReading(ctx, reading)
		if err != nil {
			t.Fatalf("SaveReading() failed: %v", err)
		}

		// Obtener últimas lecturas
		readings, err := repo.GetLatestReadings(ctx, "test-003", 10)
		if err != nil {
			t.Fatalf("GetLatestReadings() failed: %v", err)
		}

		if len(readings) == 0 {
			t.Fatal("No readings returned")
		}

		// Verificar la lectura guardada
		found := false
		for _, r := range readings {
			if r.ID == "reading-001" {
				found = true
				if r.Value != 25.5 {
					t.Errorf("Expected Value 25.5, got %.2f", r.Value)
				}
				if r.Unit != "°C" {
					t.Errorf("Expected Unit '°C', got '%s'", r.Unit)
				}
				break
			}
		}

		if !found {
			t.Error("Saved reading not found in GetLatestReadings")
		}
	})

	t.Run("GetLatestReadings_Limit", func(t *testing.T) {
		sensorID := "test-004"

		// Guardar múltiples lecturas
		for i := 0; i < 20; i++ {
			reading := &sensor.SensorReading{
				ID:        time.Now().Format("20060102150405.000000000"),
				SensorID:  sensorID,
				Type:      sensor.SensorTypeTemperature,
				Value:     float64(20 + i),
				Unit:      "°C",
				Timestamp: time.Now().UTC(),
			}
			repo.SaveReading(ctx, reading)
			time.Sleep(1 * time.Millisecond) // Asegurar timestamps únicos
		}

		// Obtener últimas 5 lecturas
		readings, err := repo.GetLatestReadings(ctx, sensorID, 5)
		if err != nil {
			t.Fatalf("GetLatestReadings() failed: %v", err)
		}

		if len(readings) > 5 {
			t.Errorf("Expected at most 5 readings, got %d", len(readings))
		}
	})

	t.Run("GetReadingsByTimeRange", func(t *testing.T) {
		sensorID := "test-005"
		now := time.Now().UTC()

		// Guardar lecturas en diferentes momentos
		timestamps := []time.Time{
			now.Add(-60 * time.Minute),
			now.Add(-30 * time.Minute),
			now.Add(-15 * time.Minute),
			now,
		}

		for i, ts := range timestamps {
			reading := &sensor.SensorReading{
				ID:        ts.Format("20060102150405.000000000"),
				SensorID:  sensorID,
				Type:      sensor.SensorTypeTemperature,
				Value:     float64(20 + i),
				Unit:      "°C",
				Timestamp: ts,
			}
			repo.SaveReading(ctx, reading)
		}

		// Obtener lecturas de los últimos 45 minutos
		start := now.Add(-45 * time.Minute)
		end := now.Add(1 * time.Minute)

		readings, err := repo.GetReadingsByTimeRange(ctx, sensorID, start, end)
		if err != nil {
			t.Fatalf("GetReadingsByTimeRange() failed: %v", err)
		}

		// Deberían ser 3 lecturas (las de -30, -15 y 0 minutos)
		if len(readings) < 2 {
			t.Errorf("Expected at least 2 readings in time range, got %d", len(readings))
		}

		// Verificar que todas las lecturas están en el rango
		for _, r := range readings {
			if r.Timestamp.Before(start) || r.Timestamp.After(end) {
				t.Errorf("Reading timestamp %v is outside range [%v, %v]", r.Timestamp, start, end)
			}
		}
	})

	t.Run("SaveReadingWithError", func(t *testing.T) {
		errorMsg := "sensor timeout"
		reading := &sensor.SensorReading{
			ID:        "reading-error-001",
			SensorID:  "test-006",
			Type:      sensor.SensorTypeTemperature,
			Value:     0,
			Unit:      "°C",
			Error:     &errorMsg,
			Timestamp: time.Now().UTC(),
		}

		err := repo.SaveReading(ctx, reading)
		if err != nil {
			t.Fatalf("SaveReading() with error failed: %v", err)
		}

		// Verificar que se puede recuperar
		readings, err := repo.GetLatestReadings(ctx, "test-006", 10)
		if err != nil {
			t.Fatalf("GetLatestReadings() failed: %v", err)
		}

		found := false
		for _, r := range readings {
			if r.ID == "reading-error-001" {
				found = true
				if r.Error == nil {
					t.Error("Expected error field to be set")
				} else if *r.Error != "sensor timeout" {
					t.Errorf("Expected error 'sensor timeout', got '%s'", *r.Error)
				}
				break
			}
		}

		if !found {
			t.Error("Reading with error not found")
		}
	})

	t.Run("GetConfig_NotFound", func(t *testing.T) {
		_, err := repo.GetConfig(ctx, "nonexistent-sensor")
		if err == nil {
			t.Error("Expected error for nonexistent sensor config, got nil")
		}
	})
}

// TestSQLiteRepository ejecuta los tests de contrato con SQLite
func TestSQLiteRepository(t *testing.T) {
	// Crear repositorio SQLite en memoria para testing
	repo, err := storage.NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("Failed to create SQLite repository: %v", err)
	}
	defer repo.Close()

	// Ejecutar tests de contrato
	RepositoryContractTests(t, repo)
}

// TestRepositoryClose verifica que Close() funciona correctamente
func TestRepositoryClose(t *testing.T) {
	repo, err := storage.NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("Failed to create SQLite repository: %v", err)
	}

	err = repo.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Intentar usar después de cerrar debería fallar
	ctx := context.Background()
	reading := &sensor.SensorReading{
		ID:        "test",
		SensorID:  "test",
		Type:      sensor.SensorTypeTemperature,
		Value:     25.0,
		Unit:      "°C",
		Timestamp: time.Now(),
	}

	err = repo.SaveReading(ctx, reading)
	if err == nil {
		t.Error("Expected error when using repository after Close(), got nil")
	}
}
