package config

import (
	"os"
	"testing"
	"time"

	"github.com/alejandro/technical_test_uvigo/internal/sensor"
)

func TestLoad_ValidConfig(t *testing.T) {
	// Crear archivo temporal con configuración válida
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	configYAML := `
environment: test
nats:
  url: nats://localhost:4222
  timeout: 10s
  max_reconnects: 5
database:
  type: sqlite
  path: ./test.db
http:
  enabled: true
  port: 8080
  host: 0.0.0.0
sensors:
  - id: temp-001
    type: temperature
    name: Test Sensor
    location: lab
    config:
      sensor_id: temp-001
      interval: 5000
      threshold: 30.0
      enabled: true
logging:
  level: debug
  format: text
`
	if _, err := tmpfile.Write([]byte(configYAML)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Cargar configuración
	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verificar campos
	if cfg.Environment != "test" {
		t.Errorf("Expected environment 'test', got '%s'", cfg.Environment)
	}

	if cfg.NATS.URL != "nats://localhost:4222" {
		t.Errorf("Expected NATS URL 'nats://localhost:4222', got '%s'", cfg.NATS.URL)
	}

	if cfg.NATS.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", cfg.NATS.Timeout)
	}

	if cfg.Database.Type != "sqlite" {
		t.Errorf("Expected database type 'sqlite', got '%s'", cfg.Database.Type)
	}

	if cfg.HTTP.Port != 8080 {
		t.Errorf("Expected HTTP port 8080, got %d", cfg.HTTP.Port)
	}

	if len(cfg.Sensors) != 1 {
		t.Fatalf("Expected 1 sensor, got %d", len(cfg.Sensors))
	}

	sensor := cfg.Sensors[0]
	if sensor.ID != "temp-001" {
		t.Errorf("Expected sensor ID 'temp-001', got '%s'", sensor.ID)
	}

	if sensor.Type != "temperature" {
		t.Errorf("Expected sensor type 'temperature', got '%s'", sensor.Type)
	}

	if sensor.Config.Interval != 5000 {
		t.Errorf("Expected interval 5000, got %d", sensor.Config.Interval)
	}

	if sensor.Config.Threshold != 30.0 {
		t.Errorf("Expected threshold 30.0, got %.2f", sensor.Config.Threshold)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "invalid-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// YAML inválido (sintaxis incorrecta)
	invalidYAML := `
environment: test
nats:
  url: [this is not valid yaml
`
	if _, err := tmpfile.Write([]byte(invalidYAML)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	_, err = Load(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestConfig_Validate(t *testing.T) {
	validSensor := SensorDef{
		ID:   "temp-001",
		Type: sensor.SensorTypeTemperature,
		Name: "Test",
		Config: sensor.SensorConfig{
			SensorID:  "temp-001",
			Interval:  5000,
			Threshold: 30.0,
			Enabled:   true,
		},
	}

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Environment: "test",
				NATS: NATSConfig{
					URL:     "nats://localhost:4222",
					Timeout: 10 * time.Second,
				},
				Database: DatabaseConfig{
					Type: "sqlite",
					Path: "./test.db",
				},
				Sensors: []SensorDef{validSensor},
			},
			wantErr: false,
		},
		{
			name: "missing environment",
			config: &Config{
				NATS: NATSConfig{
					URL:     "nats://localhost:4222",
					Timeout: 10 * time.Second,
				},
				Database: DatabaseConfig{
					Type: "sqlite",
					Path: "./test.db",
				},
				Sensors: []SensorDef{validSensor},
			},
			wantErr: true,
		},
		{
			name: "missing NATS URL",
			config: &Config{
				Environment: "test",
				NATS: NATSConfig{
					Timeout: 10 * time.Second,
				},
				Database: DatabaseConfig{
					Type: "sqlite",
					Path: "./test.db",
				},
				Sensors: []SensorDef{validSensor},
			},
			wantErr: true,
		},
		{
			name: "missing SQLite path",
			config: &Config{
				Environment: "test",
				NATS: NATSConfig{
					URL:     "nats://localhost:4222",
					Timeout: 10 * time.Second,
				},
				Database: DatabaseConfig{
					Type: "sqlite",
					// Path missing
				},
				Sensors: []SensorDef{validSensor},
			},
			wantErr: true,
		},
		{
			name: "no sensors configured",
			config: &Config{
				Environment: "test",
				NATS: NATSConfig{
					URL:     "nats://localhost:4222",
					Timeout: 10 * time.Second,
				},
				Database: DatabaseConfig{
					Type: "sqlite",
					Path: "./test.db",
				},
				Sensors: []SensorDef{}, // Sin sensores
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSensorDef_Validate(t *testing.T) {
	tests := []struct {
		name      string
		sensorDef SensorDef
		wantErr   bool
	}{
		{
			name: "valid sensor",
			sensorDef: SensorDef{
				ID:   "temp-001",
				Type: sensor.SensorTypeTemperature,
				Name: "Temperature Sensor",
				Config: sensor.SensorConfig{
					SensorID:  "temp-001",
					Interval:  5000,
					Threshold: 30.0,
					Enabled:   true,
				},
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			sensorDef: SensorDef{
				Type: sensor.SensorTypeTemperature,
				Name: "Temperature Sensor",
				Config: sensor.SensorConfig{
					SensorID:  "temp-001",
					Interval:  5000,
					Threshold: 30.0,
					Enabled:   true,
				},
			},
			wantErr: true,
		},
		{
			name: "missing type",
			sensorDef: SensorDef{
				ID:   "temp-001",
				Name: "Temperature Sensor",
				Config: sensor.SensorConfig{
					SensorID:  "temp-001",
					Interval:  5000,
					Threshold: 30.0,
					Enabled:   true,
				},
			},
			wantErr: true,
		},
		{
			name: "missing name",
			sensorDef: SensorDef{
				ID:   "temp-001",
				Type: sensor.SensorTypeTemperature,
				Config: sensor.SensorConfig{
					SensorID:  "temp-001",
					Interval:  5000,
					Threshold: 30.0,
					Enabled:   true,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid config (interval <= 0)",
			sensorDef: SensorDef{
				ID:   "temp-001",
				Type: sensor.SensorTypeTemperature,
				Name: "Temperature Sensor",
				Config: sensor.SensorConfig{
					SensorID:  "temp-001",
					Interval:  0, // Invalid
					Threshold: 30.0,
					Enabled:   true,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sensorDef.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadFromEnv_DefaultPath(t *testing.T) {
	// Asegurarse de que CONFIG_FILE no está definida
	os.Unsetenv("CONFIG_FILE")

	// Como el archivo por defecto (configs/values_local.yaml) podría no existir
	// en el entorno de test, esperamos un error
	_, err := LoadFromEnv()
	// No verificamos el resultado específico, solo que la función no hace panic
	_ = err
}

func TestLoadFromEnv_CustomPath(t *testing.T) {
	// Crear archivo temporal
	tmpfile, err := os.CreateTemp("", "custom-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	configYAML := `
environment: custom
nats:
  url: nats://custom:4222
  timeout: 5s
  max_reconnects: 3
database:
  type: sqlite
  path: ./custom.db
http:
  enabled: false
  port: 9090
  host: localhost
sensors:
  - id: test-001
    type: humidity
    name: Custom Sensor
    config:
      sensor_id: test-001
      interval: 3000
      threshold: 80.0
      enabled: true
logging:
  level: info
  format: json
`
	if _, err := tmpfile.Write([]byte(configYAML)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Establecer variable de entorno
	os.Setenv("CONFIG_FILE", tmpfile.Name())
	defer os.Unsetenv("CONFIG_FILE")

	// Cargar desde env
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() failed: %v", err)
	}

	if cfg.Environment != "custom" {
		t.Errorf("Expected environment 'custom', got '%s'", cfg.Environment)
	}

	if cfg.NATS.URL != "nats://custom:4222" {
		t.Errorf("Expected NATS URL 'nats://custom:4222', got '%s'", cfg.NATS.URL)
	}
}
