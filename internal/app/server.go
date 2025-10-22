package app

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

// Server representa el servidor IoT completo
type Server struct {
	config     *config.Config
	natsClient *natsclient.Client
	repo       repository.Repository
	simulator  *simulator.Simulator
	log        *logrus.Logger
}

// NewServer crea una nueva instancia del servidor
func NewServer(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
		log:    logger.GetLogger(),
	}
}

// Run inicializa y ejecuta el servidor
func (s *Server) Run() error {
	s.log.Info("═══════════════════════════════════════════════════════")
	s.log.Info("   IoT Sensor Server - Initializing...")
	s.log.Info("═══════════════════════════════════════════════════════")

	// 1. Conectar a NATS
	if err := s.initNATS(); err != nil {
		return fmt.Errorf("failed to initialize NATS: %w", err)
	}
	defer s.natsClient.Close()

	// 2. Inicializar base de datos
	if err := s.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer s.repo.Close()

	// 3. Inicializar simulador
	s.simulator = simulator.New(s.repo, s.natsClient)

	// 4. Registrar handlers NATS
	if err := s.registerNATSHandlers(); err != nil {
		return fmt.Errorf("failed to register NATS handlers: %w", err)
	}

	// 5. Cargar sensores desde configuración
	if err := s.loadSensors(); err != nil {
		return fmt.Errorf("failed to load sensors: %w", err)
	}

	// 6. Mostrar banner
	s.printBanner()

	// 7. Esperar señal de terminación
	return s.waitForShutdown()
}

// initNATS inicializa la conexión NATS
func (s *Server) initNATS() error {
	s.log.Infof("Connecting to NATS: %s", s.config.NATS.URL)
	client, err := natsclient.NewClient(s.config.NATS.URL)
	if err != nil {
		return err
	}
	s.natsClient = client
	s.log.Info("✓ NATS connection established")
	return nil
}

// initDatabase inicializa el repositorio de persistencia
func (s *Server) initDatabase() error {
	s.log.Infof("Initializing database: %s", s.config.Database.Type)

	var repo repository.Repository
	var err error

	switch s.config.Database.Type {
	case "sqlite":
		repo, err = storage.NewSQLiteRepository(s.config.Database.Path)
	default:
		return fmt.Errorf("unsupported database type: %s", s.config.Database.Type)
	}

	if err != nil {
		return err
	}

	s.repo = repo
	s.log.Info("✓ Database initialized")
	return nil
}

// registerNATSHandlers registra los handlers NATS
func (s *Server) registerNATSHandlers() error {
	s.log.Info("Registering NATS handlers...")

	handler := natsclient.NewHandler(s.natsClient, s.repo)
	handler.SetAddSensorCallback(s.simulator.AddSensor)

	if err := handler.HandleConfigRequests(); err != nil {
		return err
	}

	s.log.Info("✓ NATS handlers registered:")
	s.log.Info("  - sensor.config.get.*")
	s.log.Info("  - sensor.config.set.*")
	s.log.Info("  - sensor.readings.query.*")
	s.log.Info("  - sensor.register")

	return nil
}

// loadSensors carga los sensores desde la configuración
func (s *Server) loadSensors() error {
	s.log.Infof("Loading %d sensors from configuration...", len(s.config.Sensors))

	for _, sensorDef := range s.config.Sensors {
		if err := s.simulator.AddSensor(sensorDef); err != nil {
			return fmt.Errorf("failed to add sensor %s: %w", sensorDef.ID, err)
		}

		status := map[bool]string{true: "ENABLED", false: "DISABLED"}[sensorDef.Config.Enabled]
		s.log.WithFields(logrus.Fields{
			"sensor_id": sensorDef.ID,
			"type":      sensorDef.Type,
			"interval":  sensorDef.Config.Interval,
			"status":    status,
		}).Infof("  - %s (%s): interval=%dms, threshold=%.2f [%s]",
			sensorDef.ID, sensorDef.Type, sensorDef.Config.Interval, sensorDef.Config.Threshold, status)
	}

	s.log.Infof("✓ %d sensors ready", s.simulator.GetSensorCount())
	return nil
}

// printBanner muestra el banner del sistema
func (s *Server) printBanner() {
	s.log.Info("═══════════════════════════════════════════════════════")
	s.log.Info("   🚀 IoT Sensor Server is RUNNING")
	s.log.Info("═══════════════════════════════════════════════════════")
	s.log.Info("")
	s.log.Info("📊 System Status:")
	s.log.Infof("   • NATS:      %s ✓", s.config.NATS.URL)
	s.log.Infof("   • Database:  %s ✓", s.config.Database.Type)
	s.log.Infof("   • Sensors:   %d active", s.simulator.GetSensorCount())
	s.log.Info("")
	s.log.Info("📡 Publishing to NATS subjects:")
	s.log.Info("   • sensor.readings.<type>.<id>   (sensor readings)")
	s.log.Info("   • sensor.alerts.<type>.<id>     (threshold alerts)")
	s.log.Info("")
	s.log.Info("🔧 NATS request/reply endpoints:")
	s.log.Info("   • sensor.config.get.<id>        (get sensor config)")
	s.log.Info("   • sensor.config.set.<id>        (update sensor config)")
	s.log.Info("   • sensor.readings.query.<id>    (query latest readings)")
	s.log.Info("   • sensor.register               (register new sensors)")
	s.log.Info("")
	s.log.Info("Press Ctrl+C to stop...")
	s.log.Info("═══════════════════════════════════════════════════════")
}

// waitForShutdown espera una señal de terminación y ejecuta shutdown gracefully
func (s *Server) waitForShutdown() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Bloquear hasta recibir señal
	sig := <-sigChan
	s.log.Info("")
	s.log.Info("═══════════════════════════════════════════════════════")
	s.log.Infof("   🛑 Received signal: %s", sig)
	s.log.Info("   Shutting down gracefully...")
	s.log.Info("═══════════════════════════════════════════════════════")

	return s.shutdown()
}

// shutdown ejecuta el cierre ordenado del sistema
func (s *Server) shutdown() error {
	// 1. Detener simulador
	s.log.Info("[Shutdown] Stopping simulator...")
	s.simulator.Stop()
	s.log.Info("[Shutdown] ✓ Simulator stopped")

	// 2. Cerrar conexión NATS
	s.log.Info("[Shutdown] Closing NATS connection...")
	s.natsClient.Close()
	s.log.Info("[Shutdown] ✓ NATS connection closed")

	// 3. Cerrar base de datos
	s.log.Info("[Shutdown] Closing database...")
	s.repo.Close()
	s.log.Info("[Shutdown] ✓ Database closed")

	// Pequeña pausa para asegurar que todos los logs se escriben
	time.Sleep(100 * time.Millisecond)

	s.log.Info("═══════════════════════════════════════════════════════")
	s.log.Info("   ✓ IoT Sensor Server stopped successfully")
	s.log.Info("═══════════════════════════════════════════════════════")
	fmt.Println()

	return nil
}
