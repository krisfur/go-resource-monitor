# 📈 GO Resource Monitor

A lightweight resource monitor dashboard written in Go using TUI.

## Features:
- 🐹 Animated ASCII Gopher mascot
- 📊 Real-time CPU, Memory, Disk usage sparklines
- 🌐 Network throughput stats
- 🔋 Battery status
- 🎨 Clean, colorful UI

## Looks
![screencast](screencast.gif)

## Usage:
1. Clone the repository:
```
git clone https://github.com/yourusername/go-resource-monitor.git
cd go-resource-monitor
```

2. Download dependencies:
```
go mod tidy
```

3. Run the dashboard:
```
go run main.go
```

4. Alternatively, if you prefer to install it globally:
```
go install github.com/krisfur/go-resource-monitor@v0.1.0
go-resource-monitor
```

5. Press Q to quit.

## Dependencies:
- tview (https://github.com/rivo/tview)
- tcell (https://github.com/gdamore/tcell)
- gopsutil (https://github.com/shirou/gopsutil)
- distatus/battery (https://github.com/distatus/battery) for battery info


Dependencies are managed via Go modules.

## Notes:
- CPU temperature may be unavailable depending on your system.
- Root privileges may be required for certain metrics on some systems.

## License:
MIT
