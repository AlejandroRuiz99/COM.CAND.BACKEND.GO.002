package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// Init inicializa el logger global con la configuración especificada
func Init(level, format string) {
	Log = logrus.New()
	Log.SetOutput(os.Stdout)

	// Configurar nivel de log
	switch level {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "warn":
		Log.SetLevel(logrus.WarnLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}

	// Configurar formato
	if format == "json" {
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
}

// GetLogger retorna el logger configurado, o uno por defecto si no se ha inicializado
func GetLogger() *logrus.Logger {
	if Log == nil {
		Init("info", "text")
	}
	return Log
}

// WithField crea una entrada de log con un campo adicional
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

// WithFields crea una entrada de log con múltiples campos
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

// Info log de nivel info
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof log de nivel info con formato
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Debug log de nivel debug
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf log de nivel debug con formato
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Warn log de nivel warn
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf log de nivel warn con formato
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error log de nivel error
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf log de nivel error con formato
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal log de nivel fatal (termina el programa)
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf log de nivel fatal con formato (termina el programa)
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}
