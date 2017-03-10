package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jabley/monitoring-spike/pkg/apps"
	"github.com/jabley/monitoring-spike/pkg/servers"
)

func main() {
	name := apps.GetDefaultConfig("NAME", "")
	port := apps.GetDefaultConfig("PORT", "")

	metrics := apps.NewMetrics(name)

	server := servers.NewBackendServer(name, metrics)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	quit := make(chan struct{})
	go apps.MonitorProcess(name, quit, metrics)

	errorChan := make(chan error, 1)

	go servers.ListenAndServe(port, server, errorChan)

	for {
		select {
		case err := <-errorChan:
			if err != nil {
				log.Fatal(err)
			}
		case s := <-signalChan:
			log.Println(fmt.Sprintf("Captured %v. Exiting ...", s))
			d := time.Now().Add(1 * time.Second)
			ctx, cancel := context.WithDeadline(context.Background(), d)
			defer cancel()
			server.Shutdown(ctx)
			quit <- struct{}{}
			metrics.Flush()
			os.Exit(0)
		}
	}
}
