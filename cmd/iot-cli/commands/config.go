package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	natsclient "github.com/alejandro/technical_test_uvigo/internal/nats"
	"github.com/alejandro/technical_test_uvigo/internal/sensor"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Gestionar configuración de sensores",
	Long:  `Comandos para obtener y actualizar la configuración de sensores`,
}

var getConfigCmd = &cobra.Command{
	Use:   "get [sensor-id]",
	Short: "Obtener configuración de un sensor",
	Long:  `Obtiene la configuración actual de un sensor específico`,
	Args:  cobra.ExactArgs(1),
	Example: `  iot-cli config get temp-001
  iot-cli config get temp-001 --json`,
	RunE: getConfig,
}

var setConfigCmd = &cobra.Command{
	Use:   "set [sensor-id]",
	Short: "Actualizar configuración de un sensor",
	Long:  `Actualiza la configuración de un sensor específico`,
	Args:  cobra.ExactArgs(1),
	Example: `  iot-cli config set temp-001 --interval 3000 --threshold 28.5
  iot-cli config set temp-001 --interval 2000 --threshold 32.0 --enabled=false`,
	RunE: setConfig,
}

// Flags para set
var (
	setInterval  int
	setThreshold float64
	setEnabled   bool
)

func init() {
	// Flags para set
	setConfigCmd.Flags().IntVar(&setInterval, "interval", 0, "Intervalo de muestreo en milisegundos")
	setConfigCmd.Flags().Float64Var(&setThreshold, "threshold", 0, "Umbral de alerta")
	setConfigCmd.Flags().BoolVar(&setEnabled, "enabled", true, "Habilitar/deshabilitar sensor")

	// Añadir subcomandos
	configCmd.AddCommand(getConfigCmd)
	configCmd.AddCommand(setConfigCmd)
}

func getConfig(cmd *cobra.Command, args []string) error {
	sensorID := args[0]

	// Conectar a NATS
	client, err := natsclient.NewClient(natsURL)
	if err != nil {
		return fmt.Errorf("error conectando a NATS: %w", err)
	}
	defer client.Close()

	// Enviar request a NATS
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	subject := natsclient.ConfigGetSubject(sensorID)
	msg, err := client.Request(ctx, subject, nil)
	if err != nil {
		return fmt.Errorf("error obteniendo configuración: %w", err)
	}

	// Parsear respuesta
	var config sensor.SensorConfig
	if err := json.Unmarshal(msg.Data, &config); err != nil {
		// Verificar si es un mensaje de error
		var errResp map[string]string
		if json.Unmarshal(msg.Data, &errResp) == nil {
			if errMsg, ok := errResp["error"]; ok {
				return fmt.Errorf("error del servidor: %s", errMsg)
			}
		}
		return fmt.Errorf("error parseando configuración: %w", err)
	}

	if outputJSON {
		jsonOutput, _ := json.MarshalIndent(config, "", "  ")
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Printf("\n⚙️  Configuración del sensor '%s':\n\n", sensorID)

		tbl := table.New("Parámetro", "Valor")
		tbl.AddRow("Sensor ID", config.SensorID)
		tbl.AddRow("Intervalo", fmt.Sprintf("%d ms", config.Interval))
		tbl.AddRow("Threshold", fmt.Sprintf("%.2f", config.Threshold))
		tbl.AddRow("Estado", map[bool]string{true: "✅ Habilitado", false: "❌ Deshabilitado"}[config.Enabled])
		tbl.Print()
		fmt.Println()
	}

	return nil
}

func setConfig(cmd *cobra.Command, args []string) error {
	sensorID := args[0]

	// Primero obtener configuración actual
	client, err := natsclient.NewClient(natsURL)
	if err != nil {
		return fmt.Errorf("error conectando a NATS: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Obtener config actual
	subject := natsclient.ConfigGetSubject(sensorID)
	msg, err := client.Request(ctx, subject, nil)
	if err != nil {
		return fmt.Errorf("error obteniendo configuración actual: %w", err)
	}

	var currentConfig sensor.SensorConfig
	if err := json.Unmarshal(msg.Data, &currentConfig); err != nil {
		return fmt.Errorf("error parseando configuración actual: %w", err)
	}

	// Actualizar solo los valores que se especificaron
	if cmd.Flags().Changed("interval") {
		currentConfig.Interval = setInterval
	}
	if cmd.Flags().Changed("threshold") {
		currentConfig.Threshold = setThreshold
	}
	if cmd.Flags().Changed("enabled") {
		currentConfig.Enabled = setEnabled
	}

	// Validar
	if err := currentConfig.Validate(); err != nil {
		return fmt.Errorf("configuración inválida: %w", err)
	}

	// Enviar nueva configuración
	data, err := json.Marshal(currentConfig)
	if err != nil {
		return fmt.Errorf("error serializando configuración: %w", err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	setSubject := natsclient.ConfigSetSubject(sensorID)
	respMsg, err := client.Request(ctx2, setSubject, data)
	if err != nil {
		return fmt.Errorf("error actualizando configuración: %w", err)
	}

	// Verificar respuesta
	var response map[string]string
	if err := json.Unmarshal(respMsg.Data, &response); err != nil {
		return fmt.Errorf("error parseando respuesta: %w", err)
	}

	if errMsg, ok := response["error"]; ok {
		return fmt.Errorf("error del servidor: %s", errMsg)
	}

	if outputJSON {
		jsonOutput, _ := json.MarshalIndent(response, "", "  ")
		fmt.Println(string(jsonOutput))
	} else {
		printSuccess(fmt.Sprintf("Configuración del sensor '%s' actualizada", sensorID))
		fmt.Printf("\n⚙️  Nueva configuración:\n")
		fmt.Printf("  Interval:  %dms\n", currentConfig.Interval)
		fmt.Printf("  Threshold: %.2f\n", currentConfig.Threshold)
		fmt.Printf("  Estado:    %s\n", map[bool]string{true: "Habilitado", false: "Deshabilitado"}[currentConfig.Enabled])
	}

	return nil
}
