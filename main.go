package main

import (
	"github.com/krisfur/go-resource-monitor/dashboard"
	"github.com/krisfur/go-resource-monitor/metrics"
)

func main() {
	// Debug: Print available temperature sensors
	//metrics.DebugSensors()

	metricsChan := make(chan metrics.Metrics)
	quitChan := make(chan struct{})

	go metrics.CollectMetrics(metricsChan, quitChan)

	dashboard.StartUI(metricsChan, quitChan)
}
