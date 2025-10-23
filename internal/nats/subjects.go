package nats

import "fmt"

// Subjects NATS organizados jerárquicamente
const (
	SubjectReadings      = "sensor.readings"       // sensor.readings.<type>.<id>
	SubjectReadingsQuery = "sensor.readings.query" // sensor.readings.query.<id>
	SubjectConfig        = "sensor.config"         // sensor.config.<get|set>.<id>
	SubjectAlerts        = "sensor.alerts"         // sensor.alerts.<type>.<id>
	SubjectRegister      = "sensor.register"       // sensor.register
	SubjectList          = "sensor.list"           // sensor.list
)

// ReadingSubject construye el subject para publicar una lectura
// Ejemplo: "sensor.readings.temperature.temp-001"
func ReadingSubject(sensorType, sensorID string) string {
	return fmt.Sprintf("%s.%s.%s", SubjectReadings, sensorType, sensorID)
}

// ConfigGetSubject construye el subject para obtener configuración
// Ejemplo: "sensor.config.get.temp-001"
func ConfigGetSubject(sensorID string) string {
	return fmt.Sprintf("%s.get.%s", SubjectConfig, sensorID)
}

// ConfigSetSubject construye el subject para actualizar configuración
// Ejemplo: "sensor.config.set.temp-001"
func ConfigSetSubject(sensorID string) string {
	return fmt.Sprintf("%s.set.%s", SubjectConfig, sensorID)
}

// AlertSubject construye el subject para publicar alertas
// Ejemplo: "sensor.alerts.temperature.temp-001"
func AlertSubject(sensorType, sensorID string) string {
	return fmt.Sprintf("%s.%s.%s", SubjectAlerts, sensorType, sensorID)
}

// ReadingsQuerySubject construye el subject para consultar lecturas
// Ejemplo: "sensor.readings.query.temp-001"
func ReadingsQuerySubject(sensorID string) string {
	return fmt.Sprintf("%s.%s", SubjectReadingsQuery, sensorID)
}

// RegisterSubject retorna el subject para registrar nuevos sensores
func RegisterSubject() string {
	return SubjectRegister
}

// ListSubject retorna el subject para listar todos los sensores
func ListSubject() string {
	return SubjectList
}
