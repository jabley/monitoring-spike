package servers

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	navigationService = "navigation"
	contentService    = "content"
	searchService     = "search"
	productService    = "product"
	priceService      = "price"
	shipppingService  = "shipping"
	identityService   = "identity"
	customerService   = "customer"
	basketService     = "basket"
	orderService      = "order"
)

// NewFrontendServer creates a new frontend server
func NewFrontendServer(backends []Backend, logger *zap.Logger) *http.Server {
	tr := &http.Transport{
		ResponseHeaderTimeout: 2 * time.Second,
	}
	client := &http.Client{Transport: tr}

	serveMux := http.NewServeMux()
	serveMux.Handle("/", requestID(instrument(serviceTime("frontend", mainHandler(client, filterBackends(backends, homePageServices))), logger)))
	serveMux.Handle("/products", requestID(instrument(serviceTime("frontend", productListing(client, filterBackends(backends, productListingServices))), logger)))
	serveMux.Handle("/products/", requestID(instrument(serviceTime("frontend", productDetail(client, filterBackends(backends, productDetailServices))), logger)))
	serveMux.Handle("/categories", requestID(instrument(serviceTime("frontend", categoryListing(client, filterBackends(backends, categoryListingServices))), logger)))
	serveMux.Handle("/categories/", requestID(instrument(serviceTime("frontend", categoryDetail(client, filterBackends(backends, categoryDetailServices))), logger)))
	serveMux.Handle("/search", requestID(instrument(serviceTime("frontend", search(client, filterBackends(backends, searchServices))), logger)))
	serveMux.Handle("/account", requestID(instrument(serviceTime("frontend", account(client, filterBackends(backends, accountServices))), logger)))
	serveMux.Handle("/checkout", requestID(instrument(serviceTime("frontend", checkout(client, filterBackends(backends, checkoutServices))), logger)))

	return NewServer(serveMux)
}

// NewBackendServer returns a new backend server
func NewBackendServer(name string, logger *zap.Logger) *http.Server {
	serveMux := http.NewServeMux()
	serveMux.Handle("/", requestID(instrument(serviceTime(name, unreliableHandler(rand.Intn(5)+1, name)), logger)))
	return NewServer(serveMux)
}

func filterBackends(backends []Backend, desired map[string]bool) []Backend {
	var res []Backend

	for _, b := range backends {
		if desired[b.Name] {
			res = append(res, b)
		}
	}

	return res
}

// NewBackends creates new backends
func NewBackends(errorChan chan<- error) []Backend {
	backends := make([]Backend, 10)

	for i := range backends {
		name := backendName(i)

		backends[i] = Backend{
			Name:    name,
			Address: name + ":8080",
		}
	}

	return backends
}

func backendName(i int) string {
	nameLen := len(backendServiceNames)
	return fmt.Sprintf("%s", backendServiceNames[i%nameLen])
}

// NewListener creates a new listener on the specified port.
func NewListener(port string) (net.Listener, error) {
	return net.Listen("tcp", "0.0.0.0:"+port)
}

// ListenAndServe listens on the specified port.
func ListenAndServe(port string, server *http.Server, errorChan chan<- error) {
	listener, err := NewListener(port)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- server.Serve(listener)
}

// NewServer creates a new http.Server
func NewServer(serveMux http.Handler) *http.Server {
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      serveMux,
	}
}

func generateRandomID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	// Horrible magic numbers due to UUID spec in RFC4122
	b[6] = (b[6] & 0xF) | (byte(4) << 4)
	b[8] = (b[8] | 0x40) & 0x7F
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
