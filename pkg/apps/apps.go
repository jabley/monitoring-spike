package apps

import (
	"log"
	"os"
	"path"
	"runtime"
	"time"

	cgm "github.com/circonus-labs/circonus-gometrics"
)

// GetDefaultConfig returns the value from the environment, or the provided fallback if the environment is empty.
func GetDefaultConfig(name, fallback string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	return fallback
}

// MonitorProcess is used to capture periodic process metrics
func MonitorProcess(name string, quit chan struct{}, metrics *cgm.CirconusMetrics) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			metrics.Gauge("memory_usage", GetMemStats().Alloc)
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

// NewMetrics returns a new CirconusMetrics ready for use
func NewMetrics(serviceName string) *cgm.CirconusMetrics {
	cmc := &cgm.Config{}
	cmc.Debug = false // set to true for debug messages

	// seen slow networking in Docker that means it fails to connect to the HTTPTrap
	cmc.CheckManager.Broker.MaxResponseTime = "5000ms"

	// Circonus API Token key (https://login.circonus.com/user/tokens)
	cmc.CheckManager.API.TokenKey = os.Getenv("CIRCONUS_API_TOKEN")

	// Set an instance ID so that we have consistent checks even when connecting my laptop to different networks
	_, an := path.Split(os.Args[0])
	cmc.CheckManager.Check.InstanceID = "jabley.local:" + an

	if serviceName != "" {
		cmc.CheckManager.Check.InstanceID = cmc.CheckManager.Check.InstanceID + ":" + serviceName
	}

	metrics, err := cgm.NewCirconusMetrics(cmc)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return metrics
}
