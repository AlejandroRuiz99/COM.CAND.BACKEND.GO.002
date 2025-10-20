package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alejandro/technical_test_uvigo/internal/config"
	"github.com/alejandro/technical_test_uvigo/internal/logger"
	natsclient "github.com/alejandro/technical_test_uvigo/internal/nats"
	"github.com/alejandro/technical_test_uvigo/internal/repository"
	"github.com/alejandro/technical_test_uvigo/internal/simulator"
	"github.com/alejandro/technical_test_uvigo/internal/storage"
	"github.com/sirupsen/logrus"
)

func main() {
	// Inicializar logger básico temporal (se reconfigurará después de cargar config)
	logger.Init("info", "text")
	log := logger.GetLogger()

	log.Info("═══════════════════════════════════════════════════════")
	log.Info("   IoT Sensor Server - Starting...")
	log.Info("═══════════════════════════════════════════════════════")

	// ══════════════════════════════════════════════════════════════
	// PASO 1: Cargar configuración
	// ══════════════════════════════════════════════════════════════
	log.Info("[Main] Loading configuration...")
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("[Main] Failed to load configuration: %v", err)
	}

	// Reconfigurar logger con los valores del config
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	log = logger.GetLogger()

	log.WithFields(logrus.Fields{
		"environment": cfg.Environment,
		"nats_url":    cfg.NATS.URL,
		"db_type":     cfg.Database.Type,
		"db_path":     cfg.Database.Path,
		"http_port":   cfg.HTTP.Port,
		"sensors":     len(cfg.Sensors),
	}).Info("[Main] Configuration loaded")

	log.Infof("[Main] NATS URL: %s", cfg.NATS.URL)
	log.Infof("[Main] Database: %s (%s)", cfg.Database.Type, cfg.Database.Path)
	log.Infof("[Main] HTTP API: enabled=%v, port=%d", cfg.HTTP.Enabled, cfg.HTTP.Port)
	log.Infof("[Main] Sensors configured: %d", len(cfg.Sensors))

	// ══════════════════════════════════════════════════════════════
	// PASO 2: Conectar a NATS
	// ══════════════════════════════════════════════════════════════
	log.Info("[Main] Connecting to NATS...")
	natsClient, err := natsclient.NewClient(cfg.NATS.URL)
	if err != nil {
		log.Fatalf("[Main] Failed to connect to NATS: %v", err)
	}
	defer natsClient.Close()
	log.Info("[Main] ✓ Connected to NATS")

	// ══════════════════════════════════════════════════════════════
	// PASO 3: Inicializar base de datos
	// ══════════════════════════════════════════════════════════════
	log.Info("[Main] Initializing database...")
	var repo repository.Repository
	switch cfg.Database.Type {
	case "sqlite":
		repo, err = storage.NewSQLiteRepository(cfg.Database.Path)
		if err != nil {
			log.Fatalf("[Main] Failed to initialize SQLite: %v", err)
		}
		log.Infof("[Main] ✓ SQLite database initialized: %s", cfg.Database.Path)
	case "influxdb":
		// TODO: feat-8 - Implementar InfluxDB repository
		log.Fatal("[Main] InfluxDB support not yet implemented (coming in feat-8)")
	default:
		log.Fatalf("[Main] Unknown database type: %s", cfg.Database.Type)
	}
	defer repo.Close()

	// ══════════════════════════════════════════════════════════════
	// PASO 4: Registrar handlers NATS para configuración remota
	// ══════════════════════════════════════════════════════════════
	log.Info("[Main] Registering NATS handlers...")
	handler := natsclient.NewHandler(natsClient, repo)
	if err := handler.HandleConfigRequests(); err != nil {
		log.Fatalf("[Main] Failed to register NATS handlers: %v", err)
	}
	log.Info("[Main] ✓ NATS handlers registered")
	log.Info("[Main]   - sensor.config.get.*")
	log.Info("[Main]   - sensor.config.set.*")

	// ══════════════════════════════════════════════════════════════
	// PASO 5: Inicializar simulador único y añadir sensores
	// ══════════════════════════════════════════════════════════════
	log.Info("[Main] Initializing simulator...")
	sim := simulator.New(repo, natsClient)

	// Añadir sensores desde la configuración
	log.Info("[Main] Adding sensors to simulator...")
	for _, sensorDef := range cfg.Sensors {
		if err := sim.AddSensor(sensorDef); err != nil {
			log.Fatalf("[Main] Failed to add sensor %s: %v", sensorDef.ID, err)
		}

		status := "ENABLED"
		if !sensorDef.Config.Enabled {
			status = "DISABLED"
		}
		log.WithFields(logrus.Fields{
			"sensor_id": sensorDef.ID,
			"type":      sensorDef.Type,
			"interval":  sensorDef.Config.Interval,
			"threshold": sensorDef.Config.Threshold,
			"status":    status,
		}).Infof("[Main]   - %s (%s): interval=%dms, threshold=%.2f [%s]",
			sensorDef.ID, sensorDef.Type, sensorDef.Config.Interval, sensorDef.Config.Threshold, status)
	}

	// Los sensores están siendo procesados por el worker pool
	log.Infof("[Main] Simulator ready with %d sensors", sim.GetSensorCount())
	log.Info("[Main] ✓ Workers processing sensor readings")

	// ══════════════════════════════════════════════════════════════
	// PASO 6: TODO feat-6 - Iniciar servidor HTTP API (si está habilitado)
	// ══════════════════════════════════════════════════════════════
	if cfg.HTTP.Enabled {
		log.Info("[Main] HTTP API enabled but not yet implemented (coming in feat-6)")
		log.Infof("[Main] Will be available at http://%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	}

	// ══════════════════════════════════════════════════════════════
	// PASO 7: Sistema en marcha - Mostrar banner
	// ══════════════════════════════════════════════════════════════
	log.Info("═══════════════════════════════════════════════════════")
	log.Info("   🚀 IoT Sensor Server is RUNNING")
	log.Info("═══════════════════════════════════════════════════════")
	log.Info("")
	log.Info("📊 System Status:")
	log.Infof("   • NATS:      %s ✓", cfg.NATS.URL)
	log.Infof("   • Database:  %s ✓", cfg.Database.Type)
	log.Infof("   • Sensors:   %d active", sim.GetSensorCount())
	log.Info("")
	log.Info("📡 Publishing to NATS subjects:")
	log.Info("   • sensor.readings.<type>.<id>  (sensor readings)")
	log.Info("   • sensor.alerts.<type>.<id>    (threshold alerts)")
	log.Info("")
	log.Info("🔧 NATS request/reply endpoints:")
	log.Info("   • sensor.config.get.<id>       (get sensor config)")
	log.Info("   • sensor.config.set.<id>       (update sensor config)")
	log.Info("")
	log.Info("Press Ctrl+C to stop...")
	log.Info("═══════════════════════════════════════════════════════")

	// ══════════════════════════════════════════════════════════════
	// PASO 8: Esperar señal de terminación (Graceful Shutdown)
	// ══════════════════════════════════════════════════════════════
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Bloquear hasta recibir señal
	sig := <-sigChan
	log.Info("")
	log.Info("═══════════════════════════════════════════════════════")
	log.Infof("   🛑 Received signal: %s", sig)
	log.Info("   Shutting down gracefully...")
	log.Info("═══════════════════════════════════════════════════════")

	// ══════════════════════════════════════════════════════════════
	// PASO 9: Shutdown ordenado
	// ══════════════════════════════════════════════════════════════

	// 1. Detener simulador
	log.Info("[Shutdown] Stopping simulator...")
	sim.Stop()
	log.Info("[Shutdown] ✓ Simulator stopped")

	// 2. Cerrar conexión NATS (drain + close)
	log.Info("[Shutdown] Closing NATS connection...")
	natsClient.Close()
	log.Info("[Shutdown] ✓ NATS connection closed")

	// 3. Cerrar base de datos
	log.Info("[Shutdown] Closing database...")
	repo.Close()
	log.Info("[Shutdown] ✓ Database closed")

	// Pequeña pausa para asegurar que todos los logs se escriben
	time.Sleep(100 * time.Millisecond)

	log.Info("═══════════════════════════════════════════════════════")
	log.Info("   ✓ IoT Sensor Server stopped successfully")
	log.Info("═══════════════════════════════════════════════════════")
	fmt.Println()
}
