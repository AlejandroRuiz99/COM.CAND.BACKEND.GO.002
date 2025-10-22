package commands

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	natsURL    string
	outputJSON bool
	debug      bool
	log        *logrus.Logger
)

var rootCmd = &cobra.Command{
	Use:   "iot-cli",
	Short: "CLI para gestionar el sistema IoT de sensores",
	Long: `iot-cli es una herramienta de línea de comandos para interactuar 
con el sistema IoT de sensores a través de NATS.

Permite registrar sensores, consultar configuraciones, obtener lecturas
y gestionar el sistema de forma remota.`,
	Version: "0.6.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Inicializar logger
	log = logrus.New()
	log.SetOutput(os.Stderr) // Logs a stderr, output normal a stdout
	log.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    false,
	})
	log.SetLevel(logrus.WarnLevel) // Por defecto solo warnings/errors

	// Flags globales
	rootCmd.PersistentFlags().StringVar(&natsURL, "nats-url", "nats://localhost:4222", "URL del servidor NATS")
	rootCmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "Output en formato JSON")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Activar modo debug (logs verbosos)")

	// PreRun para configurar el logger según flags
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if debug {
			log.SetLevel(logrus.DebugLevel)
			log.Debug("Modo debug activado")
		}
	}

	// Añadir subcomandos
	rootCmd.AddCommand(sensorCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(readingsCmd)
}

// printError imprime un error formateado
func printError(err error) {
	fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
}

// printSuccess imprime un mensaje de éxito
func printSuccess(msg string) {
	fmt.Printf("✅ %s\n", msg)
}
