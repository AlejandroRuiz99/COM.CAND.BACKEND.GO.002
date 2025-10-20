package config

import (
	"fmt"

	"os"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/alejandro/technical_test_uvigo/internal/sensor"
)

// Config representa la configuración completa del sistema IoT
type Config struct {
	Environment string         `mapstructure:"environment"`
	NATS        NATSConfig     `mapstructure:"nats"`
	Database    DatabaseConfig `mapstructure:"database"`
	HTTP        HTTPConfig     `mapstructure:"http"`
	Sensors     []SensorDef    `mapstructure:"sensors"`
	Logging     LoggingConfig  `mapstructure:"logging"`
}

// NATSConfig contiene la configuración del servidor NATS
type NATSConfig struct {
	URL           string        `mapstructure:"url"`
	Timeout       time.Duration `mapstructure:"timeout"`
	MaxReconnects int           `mapstructure:"max_reconnects"`
}

// DatabaseConfig contiene la configuración de la base de datos
type DatabaseConfig struct {
	Type string `mapstructure:"type"` // "sqlite", "influxdb"
	Path string `mapstructure:"path"` // Para SQLite

	// Para InfluxDB (feat-8)
	URL    string `mapstructure:"url"`
	Token  string `mapstructure:"token"`
	Org    string `mapstructure:"org"`
	Bucket string `mapstructure:"bucket"`
}

// HTTPConfig contiene la configuración del servidor HTTP (feat-6)
type HTTPConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Host    string `mapstructure:"host"`
}

// SensorDef define un sensor a inicializar al arranque
type SensorDef struct {
	ID       string              `mapstructure:"id"`
	Type     sensor.SensorType   `mapstructure:"type"`
	Name     string              `mapstructure:"name"`
	Location string              `mapstructure:"location"`
	Config   sensor.SensorConfig `mapstructure:"config"`
}

// LoggingConfig contiene la configuración de logging
type LoggingConfig struct {
	Level  string `mapstructure:"level"`  // "debug", "info", "warn", "error"
	Format string `mapstructure:"format"` // "json", "text"
}

// Validate valida la configuración completa
func (c *Config) Validate() error {
	if c.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	// Validar NATS
	if c.NATS.URL == "" {
		return fmt.Errorf("nats.url is required")
	}
	if c.NATS.Timeout <= 0 {
		return fmt.Errorf("nats.timeout must be greater than 0")
	}

	// Validar Database
	if c.Database.Type == "" {
		return fmt.Errorf("database.type is required")
	}
	if c.Database.Type == "sqlite" && c.Database.Path == "" {
		return fmt.Errorf("database.path is required for sqlite")
	}

	// Validar Sensors
	if len(c.Sensors) == 0 {
		return fmt.Errorf("at least one sensor must be configured")
	}
	for i, s := range c.Sensors {
		if err := s.Validate(); err != nil {
			return fmt.Errorf("sensor[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate valida la definición de un sensor
func (s *SensorDef) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("sensor id is required")
	}
	if s.Type == "" {
		return fmt.Errorf("sensor type is required")
	}
	if s.Name == "" {
		return fmt.Errorf("sensor name is required")
	}
	return s.Config.Validate()
}

// Load carga la configuración desde un archivo usando Viper
func Load(filepath string) (*Config, error) {
	v := viper.New()

	// Configurar Viper
	v.SetConfigFile(filepath)
	v.SetConfigType("yaml")

	// Permitir override con variables de entorno
	// Ejemplo: IOT_NATS_URL overrides nats.url
	v.SetEnvPrefix("IOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Leer archivo
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filepath, err)
	}

	// Unmarshal a struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validar
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// LoadFromEnv carga la configuración usando la variable de entorno CONFIG_FILE
// Si no está definida, busca valores_local.yaml por defecto
// Soporta override de cualquier valor con variables de entorno IOT_*
func LoadFromEnv() (*Config, error) {
	v := viper.New()

	// Configurar path del archivo desde variable de entorno
	configPath := os.Getenv("CONFIG_FILE")
	if configPath == "" {
		// Default
		v.SetConfigName("values_local")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	} else {
		v.SetConfigFile(configPath)
	}

	// Permitir override con variables de entorno
	v.SetEnvPrefix("IOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Leer archivo
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Unmarshal a struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validar
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}
