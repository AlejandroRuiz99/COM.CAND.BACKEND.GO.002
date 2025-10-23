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

var rootCmd *cobra.Command

func Execute() error {
	if rootCmd == nil {
		rootCmd = createRootCommand()
		// Registrar comando interactivo después de crear el root
		RegisterInteractiveCommand(rootCmd)
	}
	return rootCmd.Execute()
}

func createRootCommand() *cobra.Command {
	// Leer NATS_URL de la variable de entorno si existe
	defaultNatsURL := "nats://localhost:4222"
	if envURL := os.Getenv("NATS_URL"); envURL != "" {
		defaultNatsURL = envURL
	}

	cmd := &cobra.Command{
		Use:   "iot-cli",
		Short: "CLI para gestionar el sistema IoT de sensores",
		Long: `iot-cli es una herramienta de línea de comandos para interactuar 
con el sistema IoT de sensores a través de NATS.

Permite registrar sensores, consultar configuraciones, obtener lecturas
y gestionar el sistema de forma remota.`,
		Version: "0.6.0",
		PersistentPreRun: func(c *cobra.Command, args []string) {
			if debug {
				log.SetLevel(logrus.DebugLevel)
				log.Debug("Modo debug activado")
			}
		},
	}

	// Flags globales
	cmd.PersistentFlags().StringVar(&natsURL, "nats-url", defaultNatsURL, "URL del servidor NATS")
	cmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "Output en formato JSON")
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "Activar modo debug (logs verbosos)")

	// Añadir subcomandos
	cmd.AddCommand(sensorCmd)
	cmd.AddCommand(configCmd)
	cmd.AddCommand(readingsCmd)

	return cmd
}

func init() {
	// Inicializar logger global
	log = logrus.New()
	log.SetOutput(os.Stderr) // Logs a stderr, output normal a stdout
	log.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    false,
	})
	log.SetLevel(logrus.WarnLevel) // Por defecto solo warnings/errors
}

// printError imprime un error formateado
func printError(err error) {
	fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
}

// printSuccess imprime un mensaje de éxito
func printSuccess(msg string) {
	fmt.Printf("✅ %s\n", msg)
}
