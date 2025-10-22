package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alejandro/technical_test_uvigo/internal/config"
	natsclient "github.com/alejandro/technical_test_uvigo/internal/nats"
	"github.com/alejandro/technical_test_uvigo/internal/sensor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var sensorCmd = &cobra.Command{
	Use:   "sensor",
	Short: "Gestionar sensores",
	Long:  `Comandos para listar, registrar y eliminar sensores del sistema`,
}

var registerSensorCmd = &cobra.Command{
	Use:   "register",
	Short: "Registrar un nuevo sensor",
	Long:  `Registra un nuevo sensor en el sistema de forma dinámica`,
	Example: `  iot-cli sensor register --id temp-005 --type temperature --name "Sala 5" --interval 5000 --threshold 30.0
  iot-cli sensor register --id hum-003 --type humidity --interval 3000 --threshold 70`,
	RunE: registerSensor,
}

var listSensorsCmd = &cobra.Command{
	Use:   "list",
	Short: "Listar todos los sensores",
	Long:  `Muestra todos los sensores registrados en el sistema`,
	RunE:  listSensors,
}

// Flags para register
var (
	sensorID   string
	sensorType string
	sensorName string
	interval   int
	threshold  float64
	enabled    bool
)

func init() {
	// Flags para register
	registerSensorCmd.Flags().StringVar(&sensorID, "id", "", "ID único del sensor (requerido)")
	registerSensorCmd.Flags().StringVar(&sensorType, "type", "", "Tipo de sensor: temperature, humidity, pressure (requerido)")
	registerSensorCmd.Flags().StringVar(&sensorName, "name", "", "Nombre descriptivo del sensor")
	registerSensorCmd.Flags().IntVar(&interval, "interval", 5000, "Intervalo de muestreo en milisegundos")
	registerSensorCmd.Flags().Float64Var(&threshold, "threshold", 30.0, "Umbral de alerta")
	registerSensorCmd.Flags().BoolVar(&enabled, "enabled", true, "Habilitar sensor")

	registerSensorCmd.MarkFlagRequired("id")
	registerSensorCmd.MarkFlagRequired("type")

	// Añadir subcomandos
	sensorCmd.AddCommand(registerSensorCmd)
	sensorCmd.AddCommand(listSensorsCmd)
}

func registerSensor(cmd *cobra.Command, args []string) error {
	log.WithFields(logrus.Fields{
		"sensor_id": sensorID,
		"type":      sensorType,
	}).Debug("Registrando nuevo sensor")

	// Validar tipo de sensor
	var st sensor.SensorType
	switch sensorType {
	case "temperature":
		st = sensor.SensorTypeTemperature
	case "humidity":
		st = sensor.SensorTypeHumidity
	case "pressure":
		st = sensor.SensorTypePressure
	default:
		return fmt.Errorf("tipo de sensor inválido: %s (debe ser: temperature, humidity, pressure)", sensorType)
	}

	// Crear definición del sensor
	sensorDef := config.SensorDef{
		ID:   sensorID,
		Type: st,
		Name: sensorName,
		Config: sensor.SensorConfig{
			SensorID:  sensorID,
			Interval:  interval,
			Threshold: threshold,
			Enabled:   enabled,
		},
	}

	// Conectar a NATS
	log.Debugf("Conectando a NATS: %s", natsURL)
	client, err := natsclient.NewClient(natsURL)
	if err != nil {
		log.Errorf("Error conectando a NATS: %v", err)
		return fmt.Errorf("error conectando a NATS: %w", err)
	}
	defer client.Close()
	log.Debug("Conexión a NATS establecida")

	// Serializar sensor
	data, err := json.Marshal(sensorDef)
	if err != nil {
		return fmt.Errorf("error serializando sensor: %w", err)
	}

	// Enviar request a NATS
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	subject := natsclient.RegisterSubject()
	msg, err := client.Request(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("error registrando sensor: %w", err)
	}

	// Parsear respuesta
	var response map[string]interface{}
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		return fmt.Errorf("error parseando respuesta: %w", err)
	}

	// Verificar si hay error
	if errMsg, ok := response["error"].(string); ok {
		return fmt.Errorf("error del servidor: %s", errMsg)
	}

	if outputJSON {
		jsonOutput, _ := json.MarshalIndent(response, "", "  ")
		fmt.Println(string(jsonOutput))
	} else {
		printSuccess(fmt.Sprintf("Sensor '%s' registrado exitosamente", sensorID))
		fmt.Printf("\n📊 Detalles:\n")
		fmt.Printf("  ID:        %s\n", sensorID)
		fmt.Printf("  Tipo:      %s\n", sensorType)
		fmt.Printf("  Nombre:    %s\n", sensorName)
		fmt.Printf("  Interval:  %dms\n", interval)
		fmt.Printf("  Threshold: %.2f\n", threshold)
		fmt.Printf("  Estado:    %s\n", map[bool]string{true: "Habilitado", false: "Deshabilitado"}[enabled])
	}

	return nil
}

func listSensors(cmd *cobra.Command, args []string) error {
	// TODO: Implementar endpoint para listar sensores
	// Por ahora solo mostramos un mensaje
	fmt.Println("⚠️  Lista de sensores no disponible")
	fmt.Println("💡 Este comando estará disponible cuando se implemente el endpoint HTTP API en feat-7")
	fmt.Println()
	fmt.Println("Alternativamente, puedes consultar sensores individuales con:")
	fmt.Println("  iot-cli config get <sensor-id>")
	return nil
}
