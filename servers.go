package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"
)

func newMainServer(backends []backend) *http.Server {
	serveMux := http.NewServeMux()

	serveMux.Handle("/", requestID(instrument(serviceTime(mainHandler(backends)))))
	serveMux.Handle("/_status", statusHandler())

	return newServer(serveMux)
}

func newBackends(errorChan chan<- error) []backend {
	backends := make([]backend, 10)

	for i := range backends {
		serveMux := http.NewServeMux()
		serveMux.Handle("/", requestID(instrument(serviceTime(unreliableHandler(rand.Intn(5)+1)))))
		server := newServer(serveMux)
		listener, err := newListener("0")

		if err != nil {
			panic(err)
		}

		go func() {
			errorChan <- server.Serve(listener)
		}()

		backends[i] = backend{
			server:  server,
			address: listener.Addr().String(),
			name:    backendName(i),
		}
	}

	return backends
}

func backendName(i int) string {
	nameLen := len(backendServiceNames)
	return fmt.Sprintf("%s_%d", backendServiceNames[i%nameLen], i%nameLen)
}

func newListener(port string) (net.Listener, error) {
	return net.Listen("tcp", "0.0.0.0:"+port)
}

func newServer(serveMux http.Handler) *http.Server {
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      serveMux,
	}
}
