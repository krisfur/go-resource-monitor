//go:build !darwin && !linux

package metrics

import (
	"strings"

	"github.com/shirou/gopsutil/v3/host"
)

// getCPUTemperature gets CPU temperature on other platforms using generic gopsutil approach
func GetCPUTemperature() float64 {
	temps, err := host.SensorsTemperatures()
	if err != nil {
		return 0.0
	}

	for _, t := range temps {
		// Try multiple common CPU temperature sensor patterns
		sensorKey := strings.ToLower(t.SensorKey)
		if strings.Contains(sensorKey, "cpu") ||
			strings.Contains(sensorKey, "core") ||
			strings.Contains(sensorKey, "die") ||
			strings.Contains(sensorKey, "thermal") ||
			strings.Contains(sensorKey, "temp") {
			return t.Temperature
		}
	}

	// If no CPU-specific sensor found, try to use the first available temperature sensor
	if len(temps) > 0 {
		return temps[0].Temperature
	}

	return 0.0
}
