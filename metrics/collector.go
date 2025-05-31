package metrics

import (
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
				if t.SensorKey == "Package id 0" || strings.Contains(strings.ToLower(t.SensorKey), "cpu") {
					cpuTemp = t.Temperature
					break
				}
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
