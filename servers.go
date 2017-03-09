package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"
)

func newMainServer(backends []backend) *http.Server {
	tr := &http.Transport{
		ResponseHeaderTimeout: 2 * time.Second,
	}
	client := &http.Client{Transport: tr}

	serveMux := http.NewServeMux()
	serveMux.Handle("/", requestID(instrument(serviceTime(mainHandler(client, filterBackends(backends, homePageServices))))))
	serveMux.Handle("/products", requestID(instrument(serviceTime(productListing(client, filterBackends(backends, productListingServices))))))
	serveMux.Handle("/products/", requestID(instrument(serviceTime(productDetail(client, filterBackends(backends, productDetailServices))))))
	serveMux.Handle("/categories", requestID(instrument(serviceTime(categoryListing(client, filterBackends(backends, categoryListingServices))))))
	serveMux.Handle("/categories/", requestID(instrument(serviceTime(categoryDetail(client, filterBackends(backends, categoryDetailServices))))))
	serveMux.Handle("/search", requestID(instrument(serviceTime(search(client, filterBackends(backends, searchServices))))))
	serveMux.Handle("/account", requestID(instrument(serviceTime(account(client, filterBackends(backends, accountServices))))))
	serveMux.Handle("/checkout", requestID(instrument(serviceTime(checkout(client, filterBackends(backends, checkoutServices))))))

	return newServer(serveMux)
}

func filterBackends(backends []backend, desired map[string]bool) []backend {
	res := make([]backend, 0)

	for _, b := range backends {
		if _, ok := desired[b.name]; ok {
			res = append(res, b)
		}
	}

	return res
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
	return fmt.Sprintf("%s", backendServiceNames[i%nameLen])
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
