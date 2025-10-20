package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alejandro/technical_test_uvigo/internal/repository"
	"github.com/alejandro/technical_test_uvigo/internal/sensor"
	natslib "github.com/nats-io/nats.go"
)

// Handler maneja las peticiones NATS relacionadas con sensores
type Handler struct {
	client *Client
	repo   repository.Repository
}

// NewHandler crea un nuevo handler con cliente NATS y repositorio
func NewHandler(client *Client, repo repository.Repository) *Handler {
	return &Handler{
		client: client,
		repo:   repo,
	}
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

// extractSensorID extrae el ID del sensor del subject NATS
// Ejemplo: "sensor.config.get.temp-001" -> "temp-001"
func extractSensorID(subject string) string {
	parts := strings.Split(subject, ".")
	if len(parts) < 4 {
		return ""
	}
	return parts[len(parts)-1]
}
