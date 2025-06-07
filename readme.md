# 📈 GO Resource Monitor

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)    [![Go](https://img.shields.io/badge/Go-1.24.4-blue)](https://go.dev/)

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
1. Install package globally
```
go install github.com/krisfur/go-resource-monitor@v0.1.3
```
2. Run it
```
go-resource-monitor
```
3. Press Q to quit.

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
