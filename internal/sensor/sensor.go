package sensor

import (
	"errors"
	"time"
)

// SensorType representa el tipo de sensor
type SensorType string

const (
	SensorTypeTemperature SensorType = "temperature"
	SensorTypeHumidity    SensorType = "humidity"
	SensorTypePressure    SensorType = "pressure"
)

// Sensor representa un sensor físico del dispositivo IoT
type Sensor struct {
	ID       string     `json:"id"`
	Type     SensorType `json:"type"`
	Name     string     `json:"name"`
	Location string     `json:"location,omitempty"`
}

// SensorConfig contiene la configuración de un sensor
type SensorConfig struct {
	SensorID  string  `json:"sensor_id" yaml:"sensor_id" mapstructure:"sensor_id"`
	Interval  int     `json:"interval" yaml:"interval" mapstructure:"interval"`    // Intervalo de muestreo en ms
	Threshold float64 `json:"threshold" yaml:"threshold" mapstructure:"threshold"` // Umbral de alerta
	Enabled   bool    `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
}

// Validate valida la configuración del sensor
func (c *SensorConfig) Validate() error {
	if c.SensorID == "" {
		return errors.New("sensor_id is required")
	}
	if c.Interval <= 0 {
		return errors.New("interval must be greater than 0")
	}
	return nil
}

// SensorReading representa una lectura de un sensor
type SensorReading struct {
	ID        string     `json:"id"`
	SensorID  string     `json:"sensor_id"`
	Type      SensorType `json:"type"`
	Value     float64    `json:"value"`
	Unit      string     `json:"unit"`
	Error     *string    `json:"error,omitempty"` // Error de lectura si existe
	Timestamp time.Time  `json:"timestamp"`
}

// Validate valida los campos obligatorios de una lectura
func (r *SensorReading) Validate() error {
	if r.ID == "" {
		return errors.New("reading id is required")
	}
	if r.SensorID == "" {
		return errors.New("sensor_id is required")
	}
	if r.Type == "" {
		return errors.New("sensor type is required")
	}
	if r.Unit == "" {
		return errors.New("unit is required")
	}
	if r.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}
	return nil
}

// IsError indica si la lectura contiene un error
func (r *SensorReading) IsError() bool {
	return r.Error != nil && *r.Error != ""
}
