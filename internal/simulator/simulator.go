package simulator

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/alejandro/technical_test_uvigo/internal/config"
	"github.com/alejandro/technical_test_uvigo/internal/logger"
	natsclient "github.com/alejandro/technical_test_uvigo/internal/nats"
	"github.com/alejandro/technical_test_uvigo/internal/repository"
	"github.com/alejandro/technical_test_uvigo/internal/sensor"
	"github.com/sirupsen/logrus"
)

const (
	defaultWorkers = 5   // Worker pool size
	taskQueueSize  = 100 // Task queue buffer
)

// sensorState mantiene el estado de un sensor individual
type sensorState struct {
	def      config.SensorDef
	ticker   *time.Ticker
	lastRead time.Time
	rand     *rand.Rand
}

// readingTask representa una tarea de lectura de sensor
type readingTask struct {
	sensorID string
	state    *sensorState
}

// Simulator gestiona múltiples sensores con worker pool
type Simulator struct {
	sensors    map[string]*sensorState // key: sensor ID
	repo       repository.Repository
	natsClient natsclient.Publisher
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	taskQueue  chan readingTask
	workers    int
}

// New crea una nueva instancia del simulador con worker pool
func New(repo repository.Repository, natsClient natsclient.Publisher) *Simulator {
	return NewWithWorkers(repo, natsClient, defaultWorkers)
}

// NewWithWorkers crea un simulador con número específico de workers
func NewWithWorkers(repo repository.Repository, natsClient natsclient.Publisher, workers int) *Simulator {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Simulator{
		sensors:    make(map[string]*sensorState),
		repo:       repo,
		natsClient: natsClient,
		ctx:        ctx,
		cancel:     cancel,
		taskQueue:  make(chan readingTask, taskQueueSize),
		workers:    workers,
	}

	// Iniciar worker pool
	s.startWorkerPool()

	return s
}

// startWorkerPool inicia los workers que procesarán las lecturas
func (s *Simulator) startWorkerPool() {
	logger.Infof("[Simulator] Starting worker pool with %d workers", s.workers)

	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
}

// AddSensor añade un sensor al simulador
func (s *Simulator) AddSensor(sensorDef config.SensorDef) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sensors[sensorDef.ID]; exists {
		return fmt.Errorf("sensor %s already exists", sensorDef.ID)
	}

	// Guardar configuración en BD
	if err := s.repo.SaveConfig(s.ctx, &sensorDef.Config); err != nil {
		return fmt.Errorf("failed to save config for sensor %s: %w", sensorDef.ID, err)
	}

	// Crear estado del sensor
	state := &sensorState{
		def:      sensorDef,
		ticker:   time.NewTicker(time.Duration(sensorDef.Config.Interval) * time.Millisecond),
		lastRead: time.Now(),
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	s.sensors[sensorDef.ID] = state

	// Si está habilitado, iniciar su ticker goroutine
	if sensorDef.Config.Enabled {
		s.wg.Add(1)
		go s.sensorTicker(sensorDef.ID, state)
	}

	logger.WithFields(logrus.Fields{
		"sensor_id": sensorDef.ID,
		"type":      sensorDef.Type,
		"interval":  sensorDef.Config.Interval,
	}).Info("[Simulator] Sensor added")

	return nil
}

// RemoveSensor elimina un sensor del simulador
func (s *Simulator) RemoveSensor(sensorID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, exists := s.sensors[sensorID]
	if !exists {
		return fmt.Errorf("sensor %s not found", sensorID)
	}

	// Detener ticker
	state.ticker.Stop()

	// Eliminar del mapa
	delete(s.sensors, sensorID)

	logger.WithField("sensor_id", sensorID).Info("[Simulator] Sensor removed")

	return nil
}

// UpdateSensorConfig actualiza la configuración de un sensor
func (s *Simulator) UpdateSensorConfig(sensorID string, newConfig *sensor.SensorConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, exists := s.sensors[sensorID]
	if !exists {
		return fmt.Errorf("sensor %s not found", sensorID)
	}

	// Actualizar config en el estado
	state.def.Config = *newConfig

	// Actualizar ticker si cambió el intervalo
	if state.ticker != nil {
		state.ticker.Reset(time.Duration(newConfig.Interval) * time.Millisecond)
	}

	// Persistir en BD
	if err := s.repo.SaveConfig(s.ctx, newConfig); err != nil {
		return fmt.Errorf("failed to save updated config: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"sensor_id": sensorID,
		"interval":  newConfig.Interval,
		"threshold": newConfig.Threshold,
	}).Info("[Simulator] Sensor config updated")

	return nil
}

// Run no hace nada - los workers ya están corriendo
func (s *Simulator) Run() {
	logger.Infof("[Simulator] Ready with %d workers processing %d sensors", s.workers, len(s.sensors))
}

// worker procesa tareas de lectura del queue (worker pool pattern)
func (s *Simulator) worker(id int) {
	defer s.wg.Done()

	logger.WithField("worker_id", id).Debug("[Simulator] Worker started")

	for {
		select {
		case <-s.ctx.Done():
			logger.WithField("worker_id", id).Debug("[Simulator] Worker stopped")
			return

		case task := <-s.taskQueue:
			// Procesar la lectura del sensor
			s.processReading(task.sensorID, task.state)
		}
	}
}

// sensorTicker envía tareas al worker pool según el intervalo del sensor
func (s *Simulator) sensorTicker(sensorID string, state *sensorState) {
	defer s.wg.Done()

	logger.WithField("sensor_id", sensorID).Debug("[Simulator] Sensor ticker started")

	for {
		select {
		case <-s.ctx.Done():
			logger.WithField("sensor_id", sensorID).Debug("[Simulator] Sensor ticker stopped")
			return

		case <-state.ticker.C:
			// Verificar si el sensor sigue habilitado
			s.mu.RLock()
			enabled := state.def.Config.Enabled
			s.mu.RUnlock()

			if !enabled {
				continue
			}

			// Enviar tarea al worker pool (non-blocking)
			select {
			case s.taskQueue <- readingTask{sensorID: sensorID, state: state}:
				// Tarea enviada
			default:
				// Queue lleno, skip esta lectura
				logger.WithField("sensor_id", sensorID).Warn("[Simulator] Task queue full, skipping reading")
			}
		}
	}
}

// processReading genera y procesa una lectura de un sensor
func (s *Simulator) processReading(sensorID string, state *sensorState) {
	// Generar lectura
	reading := s.generateReading(sensorID, state)

	// 1. Guardar en BD
	if err := s.repo.SaveReading(s.ctx, reading); err != nil {
		logger.WithFields(logrus.Fields{
			"sensor_id": sensorID,
			"error":     err,
		}).Error("[Simulator] Error saving reading")
	}

	// 2. Publicar en NATS
	subject := natsclient.ReadingSubject(string(state.def.Type), sensorID)
	data, err := json.Marshal(reading)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"sensor_id": sensorID,
			"error":     err,
		}).Error("[Simulator] Error marshaling reading")
	} else {
		if err := s.natsClient.Publish(subject, data); err != nil {
			logger.WithFields(logrus.Fields{
				"sensor_id": sensorID,
				"subject":   subject,
				"error":     err,
			}).Error("[Simulator] Error publishing reading")
		}
	}

	// 3. Verificar alertas
	s.checkAndPublishAlert(reading, state)
}

// generateReading genera una lectura simulada
func (s *Simulator) generateReading(sensorID string, state *sensorState) *sensor.SensorReading {
	reading := &sensor.SensorReading{
		ID:        fmt.Sprintf("read-%d", time.Now().UnixNano()),
		SensorID:  sensorID,
		Type:      state.def.Type,
		Timestamp: time.Now().UTC(),
	}

	// Simular error con 5% de probabilidad
	if state.rand.Float64() < 0.05 {
		errorMsg := s.generateErrorMessage(state)
		reading.Error = &errorMsg
		reading.Value = 0
		reading.Unit = s.getUnit(state.def.Type)
		return reading
	}

	// Generar valor según el tipo de sensor
	reading.Value = s.generateValue(state)
	reading.Unit = s.getUnit(state.def.Type)

	return reading
}

// generateValue genera un valor aleatorio según el tipo de sensor
func (s *Simulator) generateValue(state *sensorState) float64 {
	switch state.def.Type {
	case sensor.SensorTypeTemperature:
		// Temperatura entre 15°C y 35°C con variación
		base := 25.0
		variation := (state.rand.Float64() - 0.5) * 20.0 // ±10°C
		return base + variation

	case sensor.SensorTypeHumidity:
		// Humedad entre 30% y 80%
		base := 55.0
		variation := (state.rand.Float64() - 0.5) * 50.0 // ±25%
		value := base + variation
		// Clamp entre 0 y 100
		if value < 0 {
			return 0
		}
		if value > 100 {
			return 100
		}
		return value

	case sensor.SensorTypePressure:
		// Presión entre 980 hPa y 1040 hPa
		base := 1010.0
		variation := (state.rand.Float64() - 0.5) * 60.0 // ±30 hPa
		return base + variation

	default:
		return 0
	}
}

// getUnit retorna la unidad según el tipo de sensor
func (s *Simulator) getUnit(sensorType sensor.SensorType) string {
	switch sensorType {
	case sensor.SensorTypeTemperature:
		return "°C"
	case sensor.SensorTypeHumidity:
		return "%"
	case sensor.SensorTypePressure:
		return "hPa"
	default:
		return ""
	}
}

// generateErrorMessage genera un mensaje de error aleatorio
func (s *Simulator) generateErrorMessage(state *sensorState) string {
	errors := []string{
		"sensor timeout",
		"reading error",
		"connection lost",
		"calibration error",
		"sensor malfunction",
	}
	return errors[state.rand.Intn(len(errors))]
}

// checkAndPublishAlert verifica si el valor excede el umbral y publica alerta
func (s *Simulator) checkAndPublishAlert(reading *sensor.SensorReading, state *sensorState) {
	// Si la lectura tiene error, no verificamos threshold
	if reading.IsError() {
		return
	}

	// Verificar si se excede el umbral
	if reading.Value > state.def.Config.Threshold {
		alert := map[string]interface{}{
			"sensor_id": reading.SensorID,
			"type":      state.def.Type,
			"value":     reading.Value,
			"threshold": state.def.Config.Threshold,
			"unit":      reading.Unit,
			"timestamp": reading.Timestamp,
			"message":   fmt.Sprintf("Sensor %s exceeded threshold: %.2f %s > %.2f %s", reading.SensorID, reading.Value, reading.Unit, state.def.Config.Threshold, reading.Unit),
		}

		// Publicar alerta en NATS
		subject := natsclient.AlertSubject(string(state.def.Type), reading.SensorID)
		data, err := json.Marshal(alert)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"sensor_id": reading.SensorID,
				"error":     err,
			}).Error("[Simulator] Error marshaling alert")
			return
		}

		if err := s.natsClient.Publish(subject, data); err != nil {
			logger.WithFields(logrus.Fields{
				"sensor_id": reading.SensorID,
				"subject":   subject,
				"error":     err,
			}).Error("[Simulator] Error publishing alert")
		} else {
			logger.WithFields(logrus.Fields{
				"sensor_id": reading.SensorID,
				"type":      state.def.Type,
				"value":     reading.Value,
				"threshold": state.def.Config.Threshold,
				"unit":      reading.Unit,
			}).Warn("[Simulator] ALERT: Sensor exceeded threshold")
		}
	}
}

// GetSensorCount retorna el número de sensores activos
func (s *Simulator) GetSensorCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sensors)
}

// ListSensors retorna la lista de IDs de sensores
func (s *Simulator) ListSensors() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sensors := make([]string, 0, len(s.sensors))
	for id := range s.sensors {
		sensors = append(sensors, id)
	}
	return sensors
}

// Stop detiene el simulador, workers y espera a que todas las goroutines terminen
func (s *Simulator) Stop() {
	logger.Info("[Simulator] Stopping...")

	// Detener todos los tickers primero
	s.mu.Lock()
	for _, state := range s.sensors {
		if state.ticker != nil {
			state.ticker.Stop()
		}
	}
	s.mu.Unlock()

	// Cancelar context (detiene workers y tickers)
	s.cancel()

	// Cerrar el task queue (ya no se aceptan más tareas)
	close(s.taskQueue)

	// Esperar a que terminen todas las goroutines (workers + tickers)
	s.wg.Wait()

	logger.Info("[Simulator] Stopped successfully")
}
