package nats

import "testing"

func TestReadingSubject(t *testing.T) {
	tests := []struct {
		name       string
		sensorType string
		sensorID   string
		want       string
	}{
		{
			name:       "temperature sensor",
			sensorType: "temperature",
			sensorID:   "temp-001",
			want:       "sensor.readings.temperature.temp-001",
		},
		{
			name:       "humidity sensor",
			sensorType: "humidity",
			sensorID:   "hum-001",
			want:       "sensor.readings.humidity.hum-001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReadingSubject(tt.sensorType, tt.sensorID)
			if got != tt.want {
				t.Errorf("ReadingSubject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigGetSubject(t *testing.T) {
	got := ConfigGetSubject("temp-001")
	want := "sensor.config.get.temp-001"
	if got != want {
		t.Errorf("ConfigGetSubject() = %v, want %v", got, want)
	}
}

func TestConfigSetSubject(t *testing.T) {
	got := ConfigSetSubject("temp-001")
	want := "sensor.config.set.temp-001"
	if got != want {
		t.Errorf("ConfigSetSubject() = %v, want %v", got, want)
	}
}

func TestAlertSubject(t *testing.T) {
	got := AlertSubject("temperature", "temp-001")
	want := "sensor.alerts.temperature.temp-001"
	if got != want {
		t.Errorf("AlertSubject() = %v, want %v", got, want)
	}
}
