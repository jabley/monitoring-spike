package apps

import (
	"os"
	"time"
)

// GetDefaultConfig returns the value from the environment, or the provided fallback if the environment is empty.
func GetDefaultConfig(name, fallback string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	return fallback
}

// MonitorProcess is used to capture periodic process metrics
func MonitorProcess(quit chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			// TOOD(jabley): send process memory and other gauges to metrics collection service
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
