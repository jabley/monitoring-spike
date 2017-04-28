package main

import (
	"context"
	"flag"
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
	flag.Parse()

	port := apps.GetDefaultConfig("PORT", "8080")

	errorChan := make(chan error, 1)

	backends := servers.NewBackends(errorChan)

	srv := servers.NewFrontendServer(backends)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	quit := make(chan struct{})
	go apps.MonitorProcess("frontend", quit)

	go servers.ListenAndServe(port, srv, errorChan)

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
			srv.Shutdown(ctx)
			quit <- struct{}{}
			os.Exit(0)
		}
	}
}
