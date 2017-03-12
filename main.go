package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type key int

const (
	requestIDKey key = 0

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

const (
	indexHTML = `<!DOCTYPE html>
<html>
	<head>
		<title>Welcome to my service</title>
		<style type="text/css">
			#footer {
				border-top: 10px solid #005ea5;
			    background-color: #dee0e2;
			}
			#footer ul {
				list-style: none;
			}
			#footer ul li {
    			display: inline-block;
    			margin: 0 15px 15px 0;
			}
			#overview p {
				margin: 0 25px 0 25px;
			}
			.floated-inner-block {
				margin: 0 25px;
			}
			.homepage-top {
    			background: #005ea5;
    			color: #fff;
			}
			.homepage-top h1 {
				font-family: Arial, sans-serif;
    			font-size: 32px;
    			line-height: 1.09375;
    			text-transform: none;
    			font-size-adjust: 0.5;
    			font-weight: bold;
    			padding: 25px 0 15px;
			}
			.values-list ul {
				list-style: none;
    			padding: 0 25px;
			}
			.visuallyhidden {
 			   position: absolute;
    			left: -9999em;
			}
			p {
				font-family: Arial, sans-serif;
    			font-size: 16px;
				line-height: 1.25;
    			font-weight: 400;
    			text-transform: none;
			}
		</style>
	</head>
	<body>
		<header class="homepage-top">
			<div class="floated-inner-block">
				<h1>Welcome!</h1>
				<p>A simple app using for examining telemetry options.</p>
			</div>
		</header>
		<main>
			<section id="overview" aria-labelledby="overview-label">
				<h2 id="overview-label" class="visuallyhidden">Overview</h2>
				<p>This is a toy application which makes calls to upstream services.</p>
				<p>The upstream services might fail, or take a while to respond. This gives us "interesting" data to capture and then report on.</p>
			</section>
			<section id="responses" aria-labelledby="responses-label">
				<h2 id="responses-label" class="visuallyhidden">Responses</h2>
				<div class="values-list">
					<ul>
					{{range .}}
						<li>
							<code>{{.Key}}</code> : {{.Value}}
						</li>
					{{end}}
					</ul>
				</div>
			</section>
		</main>
		<footer id="footer">
			<div class="footer-meta">
				<h2 class="visuallyhidden">Support links</h2>
				<ul>
					<li><a href="https://github.com/jabley/monitoring-spike">Source</a></li>
					<li>Built by <a href="https://twitter.com/jabley">James Abley</a></li>
				</ul>
			</div>
		</footer>
	</body>
</html>
`
)

type backend struct {
	server  *http.Server
	address string
	name    string
}

// KeyValue makes the ENV vars into a first-class data structure
type KeyValue struct {
	Key   string
	Value string
}

// KeyValues is a shorter way of referencing an array
type KeyValues []*KeyValue

var (
	tmpl                = template.Must(template.New("index.html").Parse(indexHTML))
	backendServiceNames = []string{
		navigationService,
		contentService,
		searchService,
		productService,
		priceService,
		shipppingService,
		identityService,
		customerService,
		basketService,
		orderService,
	}

	homePageServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		basketService:     true,
	}
	productListingServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		basketService:     true,
	}
	productDetailServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		basketService:     true,
	}
	categoryListingServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		basketService:     true,
	}
	categoryDetailServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		basketService:     true,
	}
	searchServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		identityService:   true,
	}
	accountServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		identityService:   true,
	}
	checkoutServices = map[string]bool{
		navigationService: true,
		contentService:    true,
		searchService:     true,
		productService:    true,
		priceService:      true,
		customerService:   true,
		basketService:     true,
	}
)

func main() {
	flag.Parse()

	port := getDefaultConfig("PORT", "8080")

	errorChan := make(chan error, 1)

	backends := newBackends(errorChan)

	srv := newMainServer(backends)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	quit := make(chan struct{})
	go monitorProcess(quit)

	go listenAndServe(port, srv, errorChan)

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
			for _, b := range backends {
				b.server.Shutdown(ctx)
			}
			quit <- struct{}{}
			os.Exit(0)
		}
	}
}

func listenAndServe(port string, server *http.Server, errorChan chan<- error) {
	listener, err := newListener(port)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- server.Serve(listener)
}

func getDefaultConfig(name, fallback string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	return fallback
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

func monitorProcess(quit chan struct{}) {
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

// getMemStats returns a non-nill *runtime.MemStats. This is a mildly expensive call, so don't hammer it.
func getMemStats() *runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return &m
}
