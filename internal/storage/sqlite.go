package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // Driver SQLite puro Go (sin CGO)

	"github.com/alejandro/technical_test_uvigo/internal/sensor"
)

//go:embed schema.sql
var schemaSQL string

// SQLiteRepository implementa sensor.Repository usando SQLite.
// Esta implementación es específica para SQLite pero respeta el contrato
// definido en sensor.Repository, permitiendo intercambiar con otras bases
// de datos (ej: TimescaleDB) sin modificar código de negocio.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository crea una nueva instancia del repositorio SQLite.
// dbPath puede ser un archivo (ej: "./data/sensors.db") o ":memory:" para testing.
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	// Abrir conexión
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configurar límites de conexión para SQLite (single-writer)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Aplicar schema
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to apply schema: %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

// SaveReading guarda una lectura de sensor en la base de datos.
func (r *SQLiteRepository) SaveReading(ctx context.Context, reading *sensor.SensorReading) error {
	query := `
		INSERT INTO sensor_readings (id, sensor_id, type, value, unit, error, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		reading.ID,
		reading.SensorID,
		reading.Type,
		reading.Value,
		reading.Unit,
		reading.Error, // NULL si no hay error
		reading.Timestamp.UTC(),
	)

	if err != nil {
		return fmt.Errorf("failed to save reading %s: %w", reading.ID, err)
	}

	return nil
}

// GetLatestReadings obtiene las últimas N lecturas de un sensor ordenadas por timestamp descendente.
func (r *SQLiteRepository) GetLatestReadings(ctx context.Context, sensorID string, limit int) ([]*sensor.SensorReading, error) {
	query := `
		SELECT id, sensor_id, type, value, unit, error, timestamp
		FROM sensor_readings
		WHERE sensor_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, sensorID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query readings for sensor %s: %w", sensorID, err)
	}
	defer rows.Close()

	var readings []*sensor.SensorReading
	for rows.Next() {
		var r sensor.SensorReading
		var sType string
		var timestamp string

		err := rows.Scan(
			&r.ID,
			&r.SensorID,
			&sType,
			&r.Value,
			&r.Unit,
			&r.Error,
			&timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reading: %w", err)
		}

		// Convertir string a SensorType
		r.Type = sensor.SensorType(sType)

		// Parsear timestamp (SQLite guarda en formato RFC3339)
		r.Timestamp, err = time.Parse(time.RFC3339Nano, timestamp)
		if err != nil {
			// Intentar sin nanosegundos
			r.Timestamp, err = time.Parse(time.RFC3339, timestamp)
			if err != nil {
				return nil, fmt.Errorf("failed to parse timestamp: %w", err)
			}
		}

		readings = append(readings, &r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating readings: %w", err)
	}

	return readings, nil
}

// GetReadingsByTimeRange obtiene lecturas en un rango temporal específico.
func (r *SQLiteRepository) GetReadingsByTimeRange(ctx context.Context, sensorID string, start, end time.Time) ([]*sensor.SensorReading, error) {
	query := `
		SELECT id, sensor_id, type, value, unit, error, timestamp
		FROM sensor_readings
		WHERE sensor_id = ? AND timestamp >= ? AND timestamp <= ?
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, sensorID, start.UTC(), end.UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to query readings in time range: %w", err)
	}
	defer rows.Close()

	var readings []*sensor.SensorReading
	for rows.Next() {
		var r sensor.SensorReading
		var sType string
		var timestamp string

		err := rows.Scan(
			&r.ID,
			&r.SensorID,
			&sType,
			&r.Value,
			&r.Unit,
			&r.Error,
			&timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reading: %w", err)
		}

		r.Type = sensor.SensorType(sType)

		// Parsear timestamp (SQLite guarda en formato RFC3339)
		r.Timestamp, err = time.Parse(time.RFC3339Nano, timestamp)
		if err != nil {
			// Intentar sin nanosegundos
			r.Timestamp, err = time.Parse(time.RFC3339, timestamp)
			if err != nil {
				return nil, fmt.Errorf("failed to parse timestamp: %w", err)
			}
		}

		readings = append(readings, &r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating readings: %w", err)
	}

	return readings, nil
}

// SaveConfig guarda o actualiza la configuración de un sensor.
// Usa UPSERT (INSERT ... ON CONFLICT) para actualizar si ya existe.
func (r *SQLiteRepository) SaveConfig(ctx context.Context, config *sensor.SensorConfig) error {
	query := `
		INSERT INTO sensor_configs (sensor_id, interval, threshold, enabled, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(sensor_id) DO UPDATE SET
			interval = excluded.interval,
			threshold = excluded.threshold,
			enabled = excluded.enabled,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		config.SensorID,
		config.Interval,
		config.Threshold,
		config.Enabled,
	)

	if err != nil {
		return fmt.Errorf("failed to save config for sensor %s: %w", config.SensorID, err)
	}

	return nil
}

// GetConfig obtiene la configuración de un sensor.
func (r *SQLiteRepository) GetConfig(ctx context.Context, sensorID string) (*sensor.SensorConfig, error) {
	query := `
		SELECT sensor_id, interval, threshold, enabled
		FROM sensor_configs
		WHERE sensor_id = ?
	`

	var config sensor.SensorConfig
	var enabled int // SQLite guarda bool como INTEGER

	err := r.db.QueryRowContext(ctx, query, sensorID).Scan(
		&config.SensorID,
		&config.Interval,
		&config.Threshold,
		&enabled,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("config not found for sensor %s", sensorID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get config for sensor %s: %w", sensorID, err)
	}

	config.Enabled = enabled != 0

	return &config, nil
}

// Close cierra la conexión a la base de datos.
func (r *SQLiteRepository) Close() error {
	if err := r.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}
