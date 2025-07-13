//go:build linux

package metrics

import (
	"strings"

	"github.com/shirou/gopsutil/v3/host"
)

// getCPUTemperature gets CPU temperature on Linux using lm-sensors via gopsutil
func GetCPUTemperature() float64 {
	temps, err := host.SensorsTemperatures()
	if err != nil {
		return 0.0
	}

	for _, t := range temps {
		// Try multiple common CPU temperature sensor patterns
		sensorKey := strings.ToLower(t.SensorKey)
		if strings.Contains(sensorKey, "package id 0") ||
			strings.Contains(sensorKey, "cpu") ||
			strings.Contains(sensorKey, "tctl") || // AMD CPU temperature
			strings.Contains(sensorKey, "core") ||
			strings.Contains(sensorKey, "die") ||
			strings.Contains(sensorKey, "thermal") ||
			strings.Contains(sensorKey, "acpi") || // ACPI thermal zones
			strings.Contains(sensorKey, "temp") { // Generic temperature sensors
			return t.Temperature
		}
	}

	// If no CPU-specific sensor found, try to use the first available temperature sensor
	// (often ACPI thermal zones can be a reasonable fallback)
	if len(temps) > 0 {
		return temps[0].Temperature
	}

	return 0.0
}
