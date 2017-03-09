package apps

import (
	"fmt"
	"os"
	"runtime"
	"time"

	libhoney "github.com/honeycombio/libhoney-go"
)

// GetDefaultConfig returns the value from the environment, or the provided fallback if the environment is empty.
func GetDefaultConfig(name, fallback string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	return fallback
}

// MonitorProcess is used to capture periodic process metrics
func MonitorProcess(name string, quit chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			// TOOD(jabley): send process memory and other gauges to metrics collection service
			m := GetMemStats()
			ev := libhoney.NewEvent()
			ev.Dataset = name + "-memory-stats"
			ev.AddField("memory_size", m.Alloc)
			if err := ev.Send(); err != nil {
				fmt.Printf("Problem sending to honeycomb: %#v\n", err)
			}
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

// GetMemStats returns a non-nill *runtime.MemStats. This is a mildly expensive call, so don't hammer it.
func GetMemStats() *runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return &m
}
