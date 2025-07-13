//go:build darwin

package metrics

import (
	"os/exec"
	"strconv"
	"strings"
)

// GetCPUTemperature gets CPU temperature on macOS using IORegistry
func GetCPUTemperature() float64 {
	// Use IORegistry to read temperature - the only reliable method on macOS
	return tryIORegTemperature()
}

// tryIORegTemperature gets temperature from ioreg (IO Registry)
func tryIORegTemperature() float64 {
	cmd := exec.Command("ioreg", "-r", "-k", "Temperature")
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Temperature") && strings.Contains(line, "=") {
			if strings.Contains(line, "\"Temperature\"") {
				parts := strings.Split(line, "=")
				if len(parts) >= 2 {
					valueStr := strings.TrimSpace(parts[1])
					valueStr = strings.TrimSuffix(valueStr, ";")
					valueStr = strings.TrimSpace(valueStr)

					if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
						// Convert from centigrade to degrees (divide by 100)
						if value > 1000 && value < 20000 {
							temp := value / 100.0
							if temp > 0 && temp < 200 {
								return temp
							}
						}
					}
				}
			}
		}

		// Also check for VirtualTemperature as a fallback
		if strings.Contains(line, "VirtualTemperature") && strings.Contains(line, "=") {
			if strings.Contains(line, "\"VirtualTemperature\"") {
				parts := strings.Split(line, "=")
				if len(parts) >= 2 {
					valueStr := strings.TrimSpace(parts[1])
					valueStr = strings.TrimSuffix(valueStr, ";")
					valueStr = strings.TrimSpace(valueStr)

					if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
						if value > 1000 && value < 20000 {
							temp := value / 100.0
							if temp > 0 && temp < 200 {
								return temp
							}
						}
					}
				}
			}
		}
	}

	return 0.0
}
