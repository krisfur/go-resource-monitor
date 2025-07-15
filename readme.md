# 📈 Go Resource Monitor

A real-time system resource monitor with a terminal-based dashboard, written in Go.

## Features

- **CPU Usage**: Real-time CPU utilization percentage
- **CPU Temperature**: Cross-platform CPU temperature monitoring
  - **macOS**: Uses IORegistry for reliable temperature reading on Apple Silicon and Intel Macs
  - **Linux**: Uses lm-sensors via gopsutil for temperature monitoring
  - **Other platforms**: Generic sensor support via gopsutil
- **Memory Usage**: RAM utilization and detailed memory statistics
- **Disk Usage**: Storage utilization and I/O metrics
- **Network Activity**: Real-time network traffic monitoring
- **Battery Status**: Battery percentage and charging state (laptops)
- **GPU Information**: GPU utilization and temperature (when available)
- **System Uptime**: Days, hours, and minutes since boot

## Installation

```bash
go install github.com/krisfur/go-resource-monitor@latest 
```

or from source:

```bash
# Clone the repository
git clone https://github.com/krisfur/go-resource-monitor.git
cd go-resource-monitor

# Install dependencies
go mod tidy

# Build the application
go build -o go-resource-monitor

# Run the monitor
./go-resource-monitor
```

## Usage

```bash
./go-resource-monitor
```

## Platform-Specific Notes

### macOS
- Temperature monitoring uses multiple fallback methods for maximum compatibility
- Works on both Intel and Apple Silicon Macs
- No external dependencies required for basic temperature reading

### Linux
- Temperature monitoring requires lm-sensors
- Install sensors: `sudo apt-get install lm-sensors` (Ubuntu/Debian)
- Configure sensors: `sudo sensors-detect`

## Dependencies

- [gopsutil](https://github.com/shirou/gopsutil) - System and process utilities
- [battery](https://github.com/distatus/battery) - Battery information
- [go-m1cpu](https://github.com/shoenig/go-m1cpu) - Apple Silicon detection
- [tview](https://github.com/rivo/tview) - Terminal UI framework

## License

MIT License
