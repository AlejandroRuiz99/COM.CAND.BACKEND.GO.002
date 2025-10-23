package sensor

import (
	"testing"
	"time"
)

func TestSensorReading_Validate(t *testing.T) {
	tests := []struct {
		name    string
		reading SensorReading
		wantErr bool
	}{
		{
			name: "valid temperature reading",
			reading: SensorReading{
				ID:        "read-001",
				SensorID:  "temp-001",
				Type:      SensorTypeTemperature,
				Value:     23.5,
				Unit:      "°C",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid humidity reading",
			reading: SensorReading{
				ID:        "read-002",
				SensorID:  "hum-001",
				Type:      SensorTypeHumidity,
				Value:     65.0,
				Unit:      "%",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid pressure reading",
			reading: SensorReading{
				ID:        "read-003",
				SensorID:  "pres-001",
				Type:      SensorTypePressure,
				Value:     1013.25,
				Unit:      "hPa",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing sensor id",
			reading: SensorReading{
				ID:        "read-004",
				Type:      SensorTypeTemperature,
				Value:     23.5,
				Unit:      "°C",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing reading id",
			reading: SensorReading{
				SensorID:  "temp-001",
				Type:      SensorTypeTemperature,
				Value:     23.5,
				Unit:      "°C",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing type",
			reading: SensorReading{
				ID:        "read-005",
				SensorID:  "temp-001",
				Value:     23.5,
				Unit:      "°C",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing unit",
			reading: SensorReading{
				ID:        "read-006",
				SensorID:  "temp-001",
				Type:      SensorTypeTemperature,
				Value:     23.5,
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "zero timestamp",
			reading: SensorReading{
				ID:       "read-007",
				SensorID: "temp-001",
				Type:     SensorTypeTemperature,
				Value:    23.5,
				Unit:     "°C",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.reading.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSensorConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  SensorConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: SensorConfig{
				SensorID:  "temp-001",
				Interval:  1000,
				Threshold: 30.0,
				Enabled:   true,
			},
			wantErr: false,
		},
		{
			name: "missing sensor id",
			config: SensorConfig{
				Interval:  1000,
				Threshold: 30.0,
				Enabled:   true,
			},
			wantErr: true,
		},
		{
			name: "invalid interval (zero)",
			config: SensorConfig{
				SensorID:  "temp-001",
				Interval:  0,
				Threshold: 30.0,
				Enabled:   true,
			},
			wantErr: true,
		},
		{
			name: "invalid interval (negative)",
			config: SensorConfig{
				SensorID:  "temp-001",
				Interval:  -100,
				Threshold: 30.0,
				Enabled:   true,
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

