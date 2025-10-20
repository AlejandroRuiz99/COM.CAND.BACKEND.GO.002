package repository

import (
	"context"
	"time"

	"github.com/alejandro/technical_test_uvigo/internal/sensor"
)

// Repository define el contrato de persistencia para sensores.
// Esta interfaz es agnóstica de la implementación (SQLite, PostgreSQL, TimescaleDB, etc.)
// permitiendo cambiar la base de datos sin modificar la lógica de negocio.
type Repository interface {
	// SaveReading persiste una lectura de sensor
	SaveReading(ctx context.Context, reading *sensor.SensorReading) error

	// GetLatestReadings obtiene las últimas N lecturas de un sensor
	GetLatestReadings(ctx context.Context, sensorID string, limit int) ([]*sensor.SensorReading, error)

	// GetReadingsByTimeRange obtiene lecturas en un rango temporal
	GetReadingsByTimeRange(ctx context.Context, sensorID string, start, end time.Time) ([]*sensor.SensorReading, error)

	// SaveConfig guarda o actualiza la configuración de un sensor
	SaveConfig(ctx context.Context, config *sensor.SensorConfig) error

	// GetConfig obtiene la configuración de un sensor
	GetConfig(ctx context.Context, sensorID string) (*sensor.SensorConfig, error)

	// Close cierra la conexión a la base de datos
	Close() error
}
