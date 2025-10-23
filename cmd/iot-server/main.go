package main

import (
	"github.com/alejandro/technical_test_uvigo/internal/app"
	"github.com/alejandro/technical_test_uvigo/internal/config"
	"github.com/alejandro/technical_test_uvigo/internal/logger"
)

func main() {
	// 1. Logger básico para arrancar
	logger.Init("info", "text")

	// 2. Cargar configuración (Viper + YAML)
	cfg, err := config.LoadFromEnv()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// 3. Reconfigurar logger con valores del config
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)

	// 4. Crear y ejecutar servidor
	server := app.NewServer(cfg)
	if err := server.Run(); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
