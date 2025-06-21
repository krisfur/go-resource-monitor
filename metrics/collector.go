package metrics

import (
	"fmt"
	"strings"
	"time"

	"github.com/distatus/battery"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type Metrics struct {
	CPUUsage       float64
	MemoryUsage    float64
	DiskUsage      float64
	NetSentMBps    float64
	NetRecvMBps    float64
	CPUTemp        float64
	BatteryPercent float64
	BatteryState   string
}

// DebugSensors prints all available temperature sensors (call this once at startup)
func DebugSensors() {
	hostInfo, _ := host.Info()
	fmt.Printf("Platform: %s %s\n", hostInfo.Platform, hostInfo.PlatformVersion)

	temps, err := host.SensorsTemperatures()
	if err != nil {
		fmt.Printf("Error getting sensors: %v\n", err)
		return
	}

	fmt.Printf("Available temperature sensors (%d found):\n", len(temps))
	for i, t := range temps {
		fmt.Printf("  %d: Key='%s', Temp=%.1f°C\n", i, t.SensorKey, t.Temperature)
	}

	// Platform-specific guidance
	if hostInfo.Platform == "darwin" {
		fmt.Printf("\n[macOS] Note: CPU temperature may require additional tools like:\n")
		fmt.Printf("  - osx-cpu-temp: brew install osx-cpu-temp\n")
		fmt.Printf("  - iStats: gem install iStats\n")
		fmt.Printf("  - Or use the built-in thermal zones shown above\n")
	} else if hostInfo.Platform == "linux" {
		fmt.Printf("\n[Linux] Temperature sensors should work with lm-sensors.\n")
		fmt.Printf("  If no sensors found, try: sudo sensors-detect\n")
	}
}

func CollectMetrics(metricsChan chan<- Metrics, quitChan <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var prevNetSent, prevNetRecv uint64

	for {
		select {
		case <-ticker.C:
			cpuPercent, _ := cpu.Percent(0, false)

			memStats, _ := mem.VirtualMemory()
			diskStats, _ := disk.Usage("/")

			netIO, _ := net.IOCounters(false)
			var sentMBps, recvMBps float64
			if len(netIO) > 0 {
				sent := netIO[0].BytesSent
				recv := netIO[0].BytesRecv

				sentMBps = float64(sent-prevNetSent) / 1024 / 1024
				recvMBps = float64(recv-prevNetRecv) / 1024 / 1024

				prevNetSent = sent
				prevNetRecv = recv
			}

			// CPU Temperature
			cpuTemp := 0.0
			temps, _ := host.SensorsTemperatures()
			for _, t := range temps {
				// Try multiple common CPU temperature sensor patterns
				sensorKey := strings.ToLower(t.SensorKey)
				if strings.Contains(sensorKey, "package id 0") ||
					strings.Contains(sensorKey, "cpu") ||
					strings.Contains(sensorKey, "tctl") || // AMD CPU temperature
					strings.Contains(sensorKey, "core") ||
					strings.Contains(sensorKey, "die") ||
					strings.Contains(sensorKey, "thermal") || // macOS thermal zones
					strings.Contains(sensorKey, "acpi") || // ACPI thermal zones
					strings.Contains(sensorKey, "temp") { // Generic temperature sensors
					cpuTemp = t.Temperature
					break
				}
			}

			// If no CPU-specific sensor found, try to use the first available temperature sensor
			// (often ACPI thermal zones can be a reasonable fallback)
			if cpuTemp == 0.0 && len(temps) > 0 {
				cpuTemp = temps[0].Temperature
			}

			// Battery
			batteryPercent := 0.0
			batteryState := "N/A"
			batStats, _ := battery.GetAll()
			if len(batStats) > 0 {
				bat := batStats[0]
				batteryPercent = (bat.Current / bat.Full) * 100
				batteryState = bat.State.String()
			}

			metricsChan <- Metrics{
				CPUUsage:       cpuPercent[0],
				MemoryUsage:    memStats.UsedPercent,
				DiskUsage:      diskStats.UsedPercent,
				NetSentMBps:    sentMBps,
				NetRecvMBps:    recvMBps,
				CPUTemp:        cpuTemp,
				BatteryPercent: batteryPercent,
				BatteryState:   batteryState,
			}
		case <-quitChan:
			return
		}
	}
}
