package sensor

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Simulator genera lecturas de sensores simuladas
type Simulator struct {
	sensorID   string
	sensorType SensorType
	config     *SensorConfig
	mu         sync.RWMutex
	rand       *rand.Rand
}

// NewSimulator crea un nuevo simulador de sensor
func NewSimulator(sensorID string, sensorType SensorType, config *SensorConfig) *Simulator {
	return &Simulator{
		sensorID:   sensorID,
		sensorType: sensorType,
		config:     config,
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateReading genera una lectura simulada del sensor
func (s *Simulator) GenerateReading() *SensorReading {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reading := &SensorReading{
		ID:        fmt.Sprintf("read-%d", time.Now().UnixNano()),
		SensorID:  s.sensorID,
		Type:      s.sensorType,
		Timestamp: time.Now().UTC(),
	}

	// Simular error con 5% de probabilidad
	if s.rand.Float64() < 0.05 {
		errorMsg := s.generateErrorMessage()
		reading.Error = &errorMsg
		reading.Value = 0
		reading.Unit = s.getUnit()
		return reading
	}

	// Generar valor según el tipo de sensor
	reading.Value = s.generateValue()
	reading.Unit = s.getUnit()

	return reading
}

// generateValue genera un valor aleatorio según el tipo de sensor
func (s *Simulator) generateValue() float64 {
	switch s.sensorType {
	case SensorTypeTemperature:
		// Temperatura entre 15°C y 35°C con variación
		base := 25.0
		variation := (s.rand.Float64() - 0.5) * 20.0 // ±10°C
		return base + variation

	case SensorTypeHumidity:
		// Humedad entre 30% y 80%
		base := 55.0
		variation := (s.rand.Float64() - 0.5) * 50.0 // ±25%
		value := base + variation
		// Clamp entre 0 y 100
		if value < 0 {
			return 0
		}
		if value > 100 {
			return 100
		}
		return value

	case SensorTypePressure:
		// Presión entre 980 hPa y 1040 hPa
		base := 1010.0
		variation := (s.rand.Float64() - 0.5) * 60.0 // ±30 hPa
		return base + variation

	default:
		return 0
	}
}

// getUnit retorna la unidad según el tipo de sensor
func (s *Simulator) getUnit() string {
	switch s.sensorType {
	case SensorTypeTemperature:
		return "°C"
	case SensorTypeHumidity:
		return "%"
	case SensorTypePressure:
		return "hPa"
	default:
		return ""
	}
}

// generateErrorMessage genera un mensaje de error aleatorio
func (s *Simulator) generateErrorMessage() string {
	errors := []string{
		"sensor timeout",
		"reading error",
		"connection lost",
		"calibration error",
		"sensor malfunction",
	}
	return errors[s.rand.Intn(len(errors))]
}

// Run ejecuta el simulador periódicamente hasta que el context se cancele
// callback se llama con cada lectura generada
func (s *Simulator) Run(ctx context.Context, callback func(*SensorReading) error) {
	s.mu.RLock()
	interval := time.Duration(s.config.Interval) * time.Millisecond
	s.mu.RUnlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			reading := s.GenerateReading()

			// Llamar al callback (ej: publicar en NATS, guardar en DB)
			if err := callback(reading); err != nil {
				// En producción esto se loggearía
				fmt.Printf("callback error for sensor %s: %v\n", s.sensorID, err)
			}

			// Actualizar ticker si el interval cambió
			s.mu.RLock()
			newInterval := time.Duration(s.config.Interval) * time.Millisecond
			s.mu.RUnlock()

			if newInterval != interval {
				interval = newInterval
				ticker.Reset(interval)
			}
		}
	}
}

// UpdateConfig actualiza la configuración del simulador de forma segura
func (s *Simulator) UpdateConfig(config *SensorConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// GetConfig obtiene una copia de la configuración actual
func (s *Simulator) GetConfig() *SensorConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Retornar copia para evitar modificaciones externas
	return &SensorConfig{
		SensorID:  s.config.SensorID,
		Interval:  s.config.Interval,
		Threshold: s.config.Threshold,
		Enabled:   s.config.Enabled,
	}
}
