package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/alejandro/technical_test_uvigo/internal/config"
	"github.com/alejandro/technical_test_uvigo/internal/repository"
	"github.com/alejandro/technical_test_uvigo/internal/sensor"
	natslib "github.com/nats-io/nats.go"
)

// Handler maneja las peticiones NATS relacionadas con sensores
type Handler struct {
	client    *Client
	repo      repository.Repository
	addSensor func(config.SensorDef) error // Callback para añadir sensores dinámicamente
}

// NewHandler crea un nuevo handler con cliente NATS y repositorio
func NewHandler(client *Client, repo repository.Repository) *Handler {
	return &Handler{
		client: client,
		repo:   repo,
	}
}

// SetAddSensorCallback configura el callback para añadir sensores dinámicamente
func (h *Handler) SetAddSensorCallback(callback func(config.SensorDef) error) {
	h.addSensor = callback
}

// HandleConfigRequests inicia los handlers para peticiones de configuración (GET y SET)
func (h *Handler) HandleConfigRequests() error {
	// Handler para obtener configuración (GET)
	_, err := h.client.Subscribe("sensor.config.get.*", func(msg *natslib.Msg) {
		h.handleConfigGet(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to config.get: %w", err)
	}

	// Handler para actualizar configuración (SET)
	_, err = h.client.Subscribe("sensor.config.set.*", func(msg *natslib.Msg) {
		h.handleConfigSet(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to config.set: %w", err)
	}

	// Handler para consultar últimas lecturas
	_, err = h.client.Subscribe("sensor.readings.query.*", func(msg *natslib.Msg) {
		h.handleReadingsQuery(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to readings.query: %w", err)
	}

	// Handler para registrar nuevos sensores
	_, err = h.client.Subscribe("sensor.register", func(msg *natslib.Msg) {
		h.handleRegister(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to sensor.register: %w", err)
	}

	return nil
}

// handleConfigGet procesa peticiones de obtener configuración
func (h *Handler) handleConfigGet(msg *natslib.Msg) {
	// Extraer sensor ID del subject (sensor.config.get.<id>)
	sensorID := extractSensorID(msg.Subject)
	if sensorID == "" {
		h.replyError(msg, "invalid subject format")
		return
	}

	// Obtener configuración del repositorio
	config, err := h.repo.GetConfig(context.Background(), sensorID)
	if err != nil || config == nil {
		h.replyError(msg, fmt.Sprintf("config not found for sensor %s", sensorID))
		return
	}

	// Responder con la configuración
	data, err := json.Marshal(config)
	if err != nil {
		h.replyError(msg, "failed to marshal config")
		return
	}

	msg.Respond(data)
}

// handleConfigSet procesa peticiones de actualizar configuración
func (h *Handler) handleConfigSet(msg *natslib.Msg) {
	// Extraer sensor ID del subject (sensor.config.set.<id>)
	sensorID := extractSensorID(msg.Subject)
	if sensorID == "" {
		h.replyError(msg, "invalid subject format")
		return
	}

	// Parsear configuración del mensaje
	var config sensor.SensorConfig
	if err := json.Unmarshal(msg.Data, &config); err != nil {
		h.replyError(msg, "invalid config format")
		return
	}

	// Validar configuración
	if err := config.Validate(); err != nil {
		h.replyError(msg, fmt.Sprintf("invalid config: %v", err))
		return
	}

	// Guardar en repositorio
	if err := h.repo.SaveConfig(context.Background(), &config); err != nil {
		h.replyError(msg, fmt.Sprintf("failed to save config: %v", err))
		return
	}

	// Responder con éxito
	response := map[string]string{"status": "ok"}
	data, _ := json.Marshal(response)
	msg.Respond(data)
}

// replyError envía una respuesta de error en formato JSON
func (h *Handler) replyError(msg *natslib.Msg, errorMsg string) {
	response := map[string]string{"error": errorMsg}
	data, _ := json.Marshal(response)
	msg.Respond(data)
}

// handleReadingsQuery procesa peticiones para obtener últimas lecturas de un sensor
func (h *Handler) handleReadingsQuery(msg *natslib.Msg) {
	// Extraer sensor ID del subject (sensor.readings.query.<id>)
	sensorID := extractSensorID(msg.Subject)
	if sensorID == "" {
		h.replyError(msg, "invalid subject format")
		return
	}

	// Parsear límite opcional del body
	limit := 10 // Default
	if len(msg.Data) > 0 {
		var req struct {
			Limit int `json:"limit"`
		}
		if err := json.Unmarshal(msg.Data, &req); err == nil && req.Limit > 0 {
			limit = req.Limit
		}
	}

	// Obtener lecturas del repositorio
	readings, err := h.repo.GetLatestReadings(context.Background(), sensorID, limit)
	if err != nil {
		h.replyError(msg, fmt.Sprintf("failed to get readings: %v", err))
		return
	}

	// Responder con las lecturas
	data, err := json.Marshal(readings)
	if err != nil {
		h.replyError(msg, "failed to marshal readings")
		return
	}

	msg.Respond(data)
}

// handleRegister procesa peticiones para registrar nuevos sensores dinámicamente
func (h *Handler) handleRegister(msg *natslib.Msg) {
	// Verificar que el callback esté configurado
	if h.addSensor == nil {
		h.replyError(msg, "sensor registration not configured")
		return
	}

	// Parsear definición del sensor
	var sensorDef config.SensorDef
	if err := json.Unmarshal(msg.Data, &sensorDef); err != nil {
		h.replyError(msg, fmt.Sprintf("invalid sensor definition: %v", err))
		return
	}

	// Validar definición
	if sensorDef.ID == "" {
		h.replyError(msg, "sensor ID is required")
		return
	}
	if sensorDef.Type == "" {
		h.replyError(msg, "sensor type is required")
	}

	// Validar configuración
	if err := sensorDef.Config.Validate(); err != nil {
		h.replyError(msg, fmt.Sprintf("invalid config: %v", err))
		return
	}

	// Asegurar que el sensor_id en config coincide
	sensorDef.Config.SensorID = sensorDef.ID

	// Añadir sensor
	if err := h.addSensor(sensorDef); err != nil {
		h.replyError(msg, fmt.Sprintf("failed to add sensor: %v", err))
		return
	}

	// Responder con éxito
	response := map[string]interface{}{
		"status":    "ok",
		"sensor_id": sensorDef.ID,
		"message":   fmt.Sprintf("sensor %s registered successfully", sensorDef.ID),
	}
	data, _ := json.Marshal(response)
	msg.Respond(data)
}

// extractSensorID extrae el ID del sensor del subject NATS
// Ejemplo: "sensor.config.get.temp-001" -> "temp-001"
func extractSensorID(subject string) string {
	parts := strings.Split(subject, ".")
	if len(parts) < 4 {
		return ""
	}
	return parts[len(parts)-1]
}

// parseQueryLimit extrae el límite de una query string
func parseQueryLimit(query string, defaultLimit int) int {
	if query == "" {
		return defaultLimit
	}
	limit, err := strconv.Atoi(query)
	if err != nil || limit <= 0 {
		return defaultLimit
	}
	return limit
}
