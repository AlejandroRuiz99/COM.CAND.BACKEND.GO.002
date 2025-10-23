-- Schema para almacenamiento de datos de sensores IoT
-- Diseñado para ser compatible con SQLite y fácilmente migrable a TimescaleDB/PostgreSQL

-- Tabla de configuraciones de sensores
CREATE TABLE IF NOT EXISTS sensor_configs (
    sensor_id TEXT PRIMARY KEY,
    interval INTEGER NOT NULL CHECK(interval > 0),
    threshold REAL NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de lecturas de sensores (time-series data)
CREATE TABLE IF NOT EXISTS sensor_readings (
    id TEXT PRIMARY KEY,
    sensor_id TEXT NOT NULL,
    type TEXT NOT NULL,
    value REAL NOT NULL,
    unit TEXT NOT NULL,
    error TEXT,
    timestamp TIMESTAMP NOT NULL
);

-- Índice compuesto para queries por sensor + timestamp (optimización principal)
-- Este índice acelera: WHERE sensor_id = ? ORDER BY timestamp DESC
CREATE INDEX IF NOT EXISTS idx_readings_sensor_time 
    ON sensor_readings(sensor_id, timestamp DESC);

-- Índice para queries temporales globales
CREATE INDEX IF NOT EXISTS idx_readings_timestamp 
    ON sensor_readings(timestamp DESC);

-- Nota para migración a TimescaleDB:
-- 1. Cambiar tipos TIMESTAMP a TIMESTAMPTZ
-- 2. Añadir: SELECT create_hypertable('sensor_readings', 'timestamp');
-- 3. Añadir retention policy: SELECT add_retention_policy('sensor_readings', INTERVAL '30 days');
-- 4. Los índices se gestionan automáticamente con hypertables

