package dashboard

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/krisfur/go-resource-monitor/metrics"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
)

const (
	sparklinePoints = 30
)

var (
	cpuHistory     []float64
	memHistory     []float64
	diskHistory    []float64
	netSentHistory []float64
	netRecvHistory []float64
	mu             sync.Mutex
)

var gopherFrames = [][]string{
	{
		"[cyan]      ´.-::::::-.´[-]",
		"[cyan]  .:-::::::::::::::-:.[-]",
		"[cyan]  ´_::    ::    ::_´[-]",
		"[cyan]   .:( [white]^   :: ^   [cyan]):.[-]",
		"[cyan]   ´::   ([beige]..[cyan])   ::´[-]",
		"[cyan]   ´:::::::[white]UU[cyan]:::::::´[-]",
		"[cyan]   .::::::::::::::::.[-]",
		"[beige]   O[cyan]::::::::::::::::[beige]O[-]",
		"[cyan]   -::::::::::::::::-[-]",
		"[cyan]   ´::::::::::::::::´[-]",
		"[cyan]    .::::::::::::::.[-]",
		"[beige]      oO[cyan]:::::::[beige]Oo[-]",
	},
	{
		"[cyan]      ´.-::::::-.´[-]",
		"[cyan]  .:-::::::::::::::-:.[-]",
		"[cyan]  ´_::    ::    ::_´[-]",
		"[cyan]   .:(    [white]^::    ^[cyan]):.[-]",
		"[cyan]   ´::    ([beige]..[cyan])  ::´[-]",
		"[cyan]   ´:::::::[white]UU[cyan]:::::::´[-]",
		"[cyan]   .::::::::::::::::.[-]",
		"[beige]   O[cyan]::::::::::::::::[beige]O[-]",
		"[cyan]   -::::::::::::::::-[-]",
		"[cyan]   ´::::::::::::::::´[-]",
		"[cyan]    .::::::::::::::.[-]",
		"[beige]      oO[cyan]:::::::[beige]Oo[-]",
	},
}

func renderBar(label string, value float64, barWidth int) string {
	filled := int((value / 100.0) * float64(barWidth))
	empty := barWidth - filled
	bar := "[" + strings.Repeat("█", filled) + strings.Repeat(" ", empty) + "]"
	paddedLabel := fmt.Sprintf("%-8s", label)
	return fmt.Sprintf("[yellow]%s[-] %s %.1f%%", paddedLabel, bar, value)
}

func renderSparkline(history []float64) string {
	bars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var sparkline strings.Builder
	for _, point := range history {
		index := int((point / 100.0) * float64(len(bars)-1))
		if index >= len(bars) {
			index = len(bars) - 1
		} else if index < 0 {
			index = 0
		}
		sparkline.WriteRune(bars[index])
	}
	return sparkline.String()
}

func addPoint(history *[]float64, value float64) {
	mu.Lock()
	defer mu.Unlock()
	*history = append(*history, value)
	if len(*history) > sparklinePoints {
		*history = (*history)[1:] //this trims the history to not get unbounded slice growth!
	}
}

func normalizeHistory(history []float64) []float64 {
	mu.Lock()
	defer mu.Unlock()
	maxVal := 0.0
	for _, v := range history {
		if v > maxVal {
			maxVal = v
		}
	}
	normalized := make([]float64, len(history))
	for i, v := range history {
		if maxVal > 0 {
			normalized[i] = (v / maxVal) * 100.0
		} else {
			normalized[i] = 0
		}
	}
	return normalized
}

func getLocalIP() string {
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				return ip.String()
			}
		}
	}
	return "N/A"
}

func StartUI(metricsChan <-chan metrics.Metrics, quitChan chan<- struct{}) {
	app := tview.NewApplication()

	// Gopher Art Box
	gopherBox := tview.NewTextView()
	gopherBox.SetDynamicColors(true)
	gopherBox.SetBorder(true)
	gopherBox.SetTitle("Go Gopher")
	go func() {
		frame := 0
		for {
			app.QueueUpdateDraw(func() {
				gopherBox.SetText(strings.Join(gopherFrames[frame], "\n"))
			})
			frame = (frame + 1) % len(gopherFrames)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// System Info Box
	sysInfoBox := tview.NewTextView()
	sysInfoBox.SetDynamicColors(true)
	sysInfoBox.SetBorder(true)
	sysInfoBox.SetTitle("System Info")

	// Metrics Box
	metricsBox := tview.NewTextView()
	metricsBox.SetDynamicColors(true)
	metricsBox.SetBorder(true)
	metricsBox.SetTitle("Metrics")
	metricsBox.SetChangedFunc(func() {
		app.Draw()
	})

	// Footer Box
	footerBox := tview.NewTextView()
	footerBox.SetDynamicColors(true)
	footerBox.SetBorder(false)
	footerBox.SetText("[yellow]Press Q to quit.")

	// Layout
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(gopherBox, 13, 0, false)
	flex.AddItem(sysInfoBox, 7, 0, false)
	flex.AddItem(metricsBox, 0, 1, true)
	flex.AddItem(footerBox, 1, 0, false)

	// Populate System Info once
	go func() {
		hostInfo, _ := host.Info()
		cpuInfo, _ := cpu.Info()
		ip := getLocalIP()

		sysInfoText := fmt.Sprintf(
			"[yellow]OS:[-] %s %s\n[yellow]Host:[-] %s\n[yellow]Kernel:[-] %s\n[yellow]Uptime:[-] %d mins\n[yellow]CPU:[-] %s (%d cores)\n[yellow]Local IP:[-] %s",
			hostInfo.Platform, hostInfo.PlatformVersion,
			hostInfo.Hostname,
			hostInfo.KernelVersion,
			hostInfo.Uptime/60,
			cpuInfo[0].ModelName, len(cpuInfo),
			ip,
		)

		app.QueueUpdateDraw(func() {
			sysInfoBox.SetText(sysInfoText)
		})
	}()

	// Metrics Update Loop
	go func() {
		for metric := range metricsChan {
			addPoint(&cpuHistory, metric.CPUUsage)
			addPoint(&memHistory, metric.MemoryUsage)
			addPoint(&diskHistory, metric.DiskUsage)
			addPoint(&netSentHistory, metric.NetSentMBps)
			addPoint(&netRecvHistory, metric.NetRecvMBps)

			cpuSpark := renderSparkline(normalizeHistory(cpuHistory))
			memSpark := renderSparkline(normalizeHistory(memHistory))
			diskSpark := renderSparkline(normalizeHistory(diskHistory))
			sentSpark := renderSparkline(normalizeHistory(netSentHistory))
			recvSpark := renderSparkline(normalizeHistory(netRecvHistory))

			cpuTempStr := "N/A"
			if metric.CPUTemp > 0 {
				cpuTempStr = fmt.Sprintf("%.0f°C", metric.CPUTemp)
			}

			text := fmt.Sprintf(
				"%s\n[green]%s[-]\n%s\n[green]%s[-]\n%s\n[green]%s[-]\n\n"+
					"[cyan]--------------------------------[-]\n[yellow]Network Stats[-]\n[cyan]--------------------------------[-]\n"+
					"[green]Sent MBps:[-] %.2f MB/s\n[green]%s[-]\n"+
					"[blue]Recv MBps:[-] %.2f MB/s\n[blue]%s[-]\n\n"+
					"[cyan]--------------------------------[-]\n[yellow]System Stats[-]\n[cyan]--------------------------------[-]\n"+
					"[yellow]CPU Temp:[-] %s\n"+
					"[yellow]Battery:[-] %.2f%% (%s)",
				renderBar("CPU", metric.CPUUsage, 20),
				cpuSpark,
				renderBar("Memory", metric.MemoryUsage, 20),
				memSpark,
				renderBar("Disk", metric.DiskUsage, 20),
				diskSpark,
				metric.NetSentMBps,
				sentSpark,
				metric.NetRecvMBps,
				recvSpark,
				cpuTempStr,
				metric.BatteryPercent, metric.BatteryState,
			)

			app.QueueUpdateDraw(func() {
				metricsBox.SetText(text)
			})
		}
	}()

	// Key Handler
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q', 'Q':
			close(quitChan)
			app.Stop()
			return nil
		}
		return event
	})

	if err := app.SetRoot(flex, true).SetFocus(metricsBox).Run(); err != nil {
		panic(err)
	}
}
