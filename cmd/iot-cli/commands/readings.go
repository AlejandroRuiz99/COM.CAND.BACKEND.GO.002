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

var readingsCmd = &cobra.Command{
	Use:   "readings [sensor-id]",
	Short: "Consultar lecturas de un sensor",
	Long:  `Obtiene las últimas N lecturas de un sensor específico`,
	Args:  cobra.ExactArgs(1),
	Example: `  iot-cli readings temp-001
  iot-cli readings temp-001 --limit 20
  iot-cli readings temp-001 --json`,
	RunE: getReadings,
}

var limit int

func init() {
	readingsCmd.Flags().IntVarP(&limit, "limit", "l", 10, "Número máximo de lecturas a obtener")
}

func getReadings(cmd *cobra.Command, args []string) error {
	sensorID := args[0]

	// Conectar a NATS
	client, err := natsclient.NewClient(natsURL)
	if err != nil {
		return fmt.Errorf("error conectando a NATS: %w", err)
	}
	defer client.Close()

	// Preparar request
	requestData := map[string]int{"limit": limit}
	data, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("error preparando request: %w", err)
	}

	// Enviar request a NATS
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	subject := natsclient.ReadingsQuerySubject(sensorID)
	msg, err := client.Request(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("error consultando lecturas: %w", err)
	}

	// Parsear respuesta
	var readings []*sensor.SensorReading
	if err := json.Unmarshal(msg.Data, &readings); err != nil {
		// Verificar si es un mensaje de error
		var errResp map[string]string
		if json.Unmarshal(msg.Data, &errResp) == nil {
			if errMsg, ok := errResp["error"]; ok {
				return fmt.Errorf("error del servidor: %s", errMsg)
			}
		}
		return fmt.Errorf("error parseando lecturas: %w", err)
	}

	if len(readings) == 0 {
		fmt.Printf("\n⚠️  No hay lecturas disponibles para el sensor '%s'\n\n", sensorID)
		return nil
	}

	if outputJSON {
		jsonOutput, _ := json.MarshalIndent(readings, "", "  ")
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Printf("\n📈 Últimas %d lecturas del sensor '%s':\n\n", len(readings), sensorID)

		tbl := table.New("ID", "Tipo", "Valor", "Unidad", "Timestamp", "Error")
		for _, reading := range readings {
			timestamp := reading.Timestamp.Format("2006-01-02 15:04:05")
			errorMsg := "-"
			if reading.Error != nil {
				errorMsg = *reading.Error
			}

			tbl.AddRow(
				reading.ID,
				string(reading.Type),
				fmt.Sprintf("%.2f", reading.Value),
				reading.Unit,
				timestamp,
				errorMsg,
			)
		}

		tbl.Print()
		fmt.Println()

		// Estadísticas básicas
		var sum float64
		var max, min float64
		var errorCount int

		for i, r := range readings {
			if r.Error == nil {
				sum += r.Value
				if i == 0 {
					max = r.Value
					min = r.Value
				} else {
					if r.Value > max {
						max = r.Value
					}
					if r.Value < min {
						min = r.Value
					}
				}
			} else {
				errorCount++
			}
		}

		validReadings := len(readings) - errorCount
		if validReadings > 0 {
			avg := sum / float64(validReadings)
			fmt.Printf("📊 Estadísticas:\n")
			fmt.Printf("  Promedio: %.2f\n", avg)
			fmt.Printf("  Máximo:   %.2f\n", max)
			fmt.Printf("  Mínimo:   %.2f\n", min)
			if errorCount > 0 {
				fmt.Printf("  Errores:  %d/%d (%.1f%%)\n", errorCount, len(readings), float64(errorCount)/float64(len(readings))*100)
			}
			fmt.Println()
		}
	}

	return nil
}
