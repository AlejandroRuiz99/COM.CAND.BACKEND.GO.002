package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Modo interactivo de la CLI",
	Long: `Inicia un modo interactivo donde puedes ejecutar comandos sin el prefijo 'iot-cli'.

Ejemplos:
  sensor list
  sensor register --type temperature --id temp-999
  config get temp-001
  readings latest temp-001 5
  exit (para salir)`,
	RunE: runInteractive,
}

func init() {
	// El comando se registrará en Execute() de root.go
}

// RegisterInteractiveCommand registra el comando interactive al rootCmd
func RegisterInteractiveCommand(root *cobra.Command) {
	root.AddCommand(interactiveCmd)
}

func runInteractive(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("╔═══════════════════════════════════════════════════════╗")
	fmt.Println("║      IoT CLI - Modo Interactivo                      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════╝")
	fmt.Printf("📡 Conectado a: %s\n", natsURL)
	fmt.Println("\nComandos disponibles:")
	fmt.Println("  sensor list")
	fmt.Println("  sensor register --type <type> --id <id>")
	fmt.Println("  config get <sensor-id>")
	fmt.Println("  config set <sensor-id> --enabled=true --interval=3000")
	fmt.Println("  readings latest <sensor-id> [limit]")
	fmt.Println("  help               - Mostrar ayuda")
	fmt.Println("  exit               - Salir del modo interactivo")
	fmt.Println()

	for {
		fmt.Print("iot> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error leyendo entrada: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)

		// Salir del modo interactivo
		if input == "exit" || input == "quit" || input == "q" {
			fmt.Println("👋 Saliendo del modo interactivo...")
			return nil
		}

		// Comando vacío
		if input == "" {
			continue
		}

		// Mostrar ayuda
		if input == "help" || input == "?" {
			showInteractiveHelp()
			continue
		}

		// Dividir el input en argumentos
		cmdArgs := parseCommand(input)

		// Ejecutar el comando
		executeCommand(cmdArgs)
	}
}

func parseCommand(input string) []string {
	// Parser simple que respeta comillas
	var args []string
	var current strings.Builder
	inQuotes := false

	for _, char := range input {
		switch char {
		case '"':
			inQuotes = !inQuotes
		case ' ':
			if inQuotes {
				current.WriteRune(char)
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

func executeCommand(args []string) {
	if len(args) == 0 {
		return
	}

	// Crear una nueva instancia del root command para cada ejecución
	// Esto evita conflictos de estado entre comandos (más limpio que resetear)
	newRootCmd := createRootCommand()

	// Configurar argumentos
	newRootCmd.SetArgs(args)

	// Ejecutar comando (sin output de errores duplicados)
	newRootCmd.SilenceErrors = true
	newRootCmd.SilenceUsage = true

	if err := newRootCmd.Execute(); err != nil {
		// Imprimir error solo si no es de Cobra
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func showInteractiveHelp() {
	fmt.Println("\n📖 Ayuda - Comandos disponibles:\n")
	fmt.Println("Sensores:")
	fmt.Println("  sensor list                           - Listar todos los sensores")
	fmt.Println("  sensor register --type TYPE --id ID   - Registrar nuevo sensor")
	fmt.Println()
	fmt.Println("Configuración:")
	fmt.Println("  config get SENSOR_ID                  - Obtener config de un sensor")
	fmt.Println("  config set SENSOR_ID [opciones]       - Actualizar config")
	fmt.Println()
	fmt.Println("Lecturas:")
	fmt.Println("  readings latest SENSOR_ID [LIMIT]     - Últimas N lecturas")
	fmt.Println()
	fmt.Println("Otros:")
	fmt.Println("  help, ?                               - Mostrar esta ayuda")
	fmt.Println("  exit, quit, q                         - Salir")
	fmt.Println()
}
