package metrics

import (
	"os/exec"
	"strconv"
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

	// New metrics
	DiskReadMBps       float64
	DiskWriteMBps      float64
	DiskReadOps        uint64
	DiskWriteOps       uint64
	MemoryTotal        uint64
	MemoryAvailable    uint64
	MemoryCached       uint64
	SwapUsage          float64
	NetworkPacketsSent uint64
	NetworkPacketsRecv uint64
	UptimeDays         int
	UptimeHours        int
	UptimeMinutes      int

	// GPU metrics
	GPUs []GPUInfo
}

type GPUInfo struct {
	Name        string
	MemoryUsed  uint64
	MemoryTotal uint64
	Temperature float64
	Utilization float64
}

func CollectMetrics(metricsChan chan<- Metrics, quitChan <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var prevNetSent, prevNetRecv uint64
	var prevDiskRead, prevDiskWrite uint64

	for {
		select {
		case <-ticker.C:
			cpuPercent, _ := cpu.Percent(0, false)

			memStats, _ := mem.VirtualMemory()
			diskStats, _ := disk.Usage("/")

			netIO, _ := net.IOCounters(false)
			var sentMBps, recvMBps float64
			var packetsSent, packetsRecv uint64
			if len(netIO) > 0 {
				sent := netIO[0].BytesSent
				recv := netIO[0].BytesRecv
				packetsSent = netIO[0].PacketsSent
				packetsRecv = netIO[0].PacketsRecv

				sentMBps = float64(sent-prevNetSent) / 1024 / 1024
				recvMBps = float64(recv-prevNetRecv) / 1024 / 1024

				prevNetSent = sent
				prevNetRecv = recv
			}

			// Disk I/O
			diskIO, _ := disk.IOCounters()
			var diskReadMBps, diskWriteMBps float64
			var diskReadOps, diskWriteOps uint64
			if len(diskIO) > 0 {
				// Sum up all disk I/O
				var totalRead, totalWrite uint64
				for _, io := range diskIO {
					totalRead += io.ReadBytes
					totalWrite += io.WriteBytes
					diskReadOps += io.ReadCount
					diskWriteOps += io.WriteCount
				}

				diskReadMBps = float64(totalRead-prevDiskRead) / 1024 / 1024
				diskWriteMBps = float64(totalWrite-prevDiskWrite) / 1024 / 1024

				prevDiskRead = totalRead
				prevDiskWrite = totalWrite
			}

			// CPU Temperature - now using platform-specific implementation
			cpuTemp := GetCPUTemperature()

			// Battery
			batteryPercent := 0.0
			batteryState := "N/A"
			batStats, _ := battery.GetAll()
			if len(batStats) > 0 {
				bat := batStats[0]
				batteryPercent = (bat.Current / bat.Full) * 100
				batteryState = bat.State.String()
			}

			// Uptime
			hostInfo, _ := host.Info()
			uptime := hostInfo.Uptime
			uptimeDays := int(uptime / 86400)
			uptimeHours := int((uptime % 86400) / 3600)
			uptimeMinutes := int((uptime % 3600) / 60)

			// GPU Information - Detect GPUs from temperature sensors and system info
			var gpus []GPUInfo

			// Get temperature sensors for GPU detection
			temps, _ := host.SensorsTemperatures()

			// First, detect GPUs from temperature sensors
			for _, t := range temps {
				sensorKey := strings.ToLower(t.SensorKey)
				if strings.Contains(sensorKey, "gpu") ||
					strings.Contains(sensorKey, "amdgpu") ||
					strings.Contains(sensorKey, "nvidia") ||
					strings.Contains(sensorKey, "edge") { // AMD GPU edge temperature

					// Extract GPU name from sensor key
					gpuName := "GPU"
					if strings.Contains(sensorKey, "amdgpu") {
						gpuName = "AMD GPU"
					} else if strings.Contains(sensorKey, "nvidia") {
						gpuName = "NVIDIA GPU"
					} else if strings.Contains(sensorKey, "edge") {
						gpuName = "AMD GPU"
					}

					gpus = append(gpus, GPUInfo{
						Name:        gpuName,
						MemoryUsed:  0, // Not available without GPU-specific libraries
						MemoryTotal: 0, // Not available without GPU-specific libraries
						Temperature: t.Temperature,
						Utilization: 0, // Not available without GPU-specific libraries
					})
				}
			}

			// Check for NVIDIA GPU via nvidia-smi if not already detected
			nvidiaDetected := false
			for _, gpu := range gpus {
				if strings.Contains(strings.ToLower(gpu.Name), "nvidia") {
					nvidiaDetected = true
					break
				}
			}

			if !nvidiaDetected {
				// Try to get NVIDIA GPU temperature and utilization via nvidia-smi
				tempCmd := exec.Command("nvidia-smi", "--query-gpu=temperature.gpu", "--format=csv,noheader,nounits")
				utilCmd := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits")

				var temp, util float64
				if output, err := tempCmd.Output(); err == nil {
					if tempStr := strings.TrimSpace(string(output)); tempStr != "" {
						temp, _ = strconv.ParseFloat(tempStr, 64)
					}
				}
				if output, err := utilCmd.Output(); err == nil {
					if utilStr := strings.TrimSpace(string(output)); utilStr != "" {
						util, _ = strconv.ParseFloat(strings.TrimRight(utilStr, " %"), 64)
					}
				}

				if temp > 0 || util > 0 {
					gpus = append(gpus, GPUInfo{
						Name:        "NVIDIA GPU",
						Temperature: temp,
						Utilization: util,
					})
				}
			}

			// AMD GPU Utilization
			for i, gpu := range gpus {
				if strings.Contains(strings.ToLower(gpu.Name), "amd") {
					// Try to read utilization from sysfs
					if utilData, err := exec.Command("cat", "/sys/class/drm/card0/device/gpu_busy_percent").Output(); err == nil {
						if utilStr := strings.TrimSpace(string(utilData)); utilStr != "" {
							gpus[i].Utilization, _ = strconv.ParseFloat(utilStr, 64)
						}
					}
				}
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

				// New metrics
				DiskReadMBps:       diskReadMBps,
				DiskWriteMBps:      diskWriteMBps,
				DiskReadOps:        diskReadOps,
				DiskWriteOps:       diskWriteOps,
				MemoryTotal:        memStats.Total,
				MemoryAvailable:    memStats.Available,
				MemoryCached:       memStats.Cached,
				SwapUsage:          0.0, // TODO: Fix swap usage calculation
				NetworkPacketsSent: packetsSent,
				NetworkPacketsRecv: packetsRecv,
				UptimeDays:         uptimeDays,
				UptimeHours:        uptimeHours,
				UptimeMinutes:      uptimeMinutes,

				// GPU metrics
				GPUs: gpus,
			}
		case <-quitChan:
			return
		}
	}
}
