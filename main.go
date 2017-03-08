package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
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
}

// KeyValue makes the ENV vars into a first-class data structure
type KeyValue struct {
	Key   string
	Value string
}

// KeyValues is a shorter way of referencing an array
type KeyValues []*KeyValue

var (
	tmpl = template.Must(template.New("index.html").Parse(indexHTML))
)

func main() {
	flag.Parse()

	port := getDefaultConfig("PORT", "8080")

	errorChan := make(chan error, 1)

	backends := newBackends(errorChan)

	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/", mainHandler(backends))
	serveMux.HandleFunc("/_status", statusHandler)

	srv := newServer(serveMux)

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		listener, err := newListener(port)
		if err != nil {
			errorChan <- err
			return
		}
		errorChan <- srv.Serve(listener)
	}()

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
			os.Exit(0)
		}
	}
}

func newBackends(errorChan chan<- error) []backend {
	backends := make([]backend, 10)

	for i := range backends {
		serveMux := http.NewServeMux()
		serveMux.HandleFunc("/", unreliableHandler(rand.Intn(5)+1))
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
		}
	}

	return backends
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

func getDefaultConfig(name, fallback string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	return fallback
}

func mainHandler(backends []backend) http.HandlerFunc {
	client := &http.Client{}

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Frontend received request\n")

		results := make(chan KeyValue, len(backends))

		var wg sync.WaitGroup

		for _, b := range backends {
			wg.Add(1)

			go func(address string, results chan<- KeyValue) {
				defer wg.Done()

				fmt.Printf("Sending request to backend %s\n", address)

				res, err := client.Get("http://" + address)

				fmt.Printf("Received response from backend %s\n", address)

				if err != nil {
					results <- KeyValue{address, err.Error()}
				} else {
					defer res.Body.Close()
					results <- KeyValue{address, res.Status}
				}
			}(b.address, results)
		}

		wg.Wait()

		values := make([]KeyValue, len(backends))
		for i := range values {
			values[i] = <-results
		}

		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")

		if err := tmpl.Execute(w, values); err != nil {
			panic(err)
		}
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	json.NewEncoder(w).Encode(mem)
}

func unreliableHandler(percentageFailures int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Backend received request\n")

		if rand.Intn(100) < percentageFailures {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
  "errors": [
    {
      "status": "400",
      "source": { "pointer": "/data/attributes/first-name" },
      "title":  "Invalid Attribute",
      "detail": "First name must contain at least three characters."
    }
  ]
}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
  "data": [{
    "type": "articles",
    "id": "1",
    "attributes": {
      "title": "JSON API paints my bikeshed!",
      "body": "The shortest article. Ever.",
      "created": "2015-05-22T14:56:29.000Z",
      "updated": "2015-05-22T14:56:28.000Z"
    },
    "relationships": {
      "author": {
        "data": {"id": "42", "type": "people"}
      }
    }
  }],
  "included": [
    {
      "type": "people",
      "id": "42",
      "attributes": {
        "name": "John",
        "age": 80,
        "gender": "male"
      }
    }
  ]
}`))
		}
	}
}

func getKeyValues() KeyValues {
	result := make(KeyValues, 2)
	result[0] = &KeyValue{"PORT", os.Getenv("PORT")}
	result[1] = &KeyValue{"PROVIDER", os.Getenv("PROVIDER")}
	return result
}

func newKeyValue(kv string) *KeyValue {
	s := strings.Split(kv, "=")
	return &KeyValue{Key: s[0], Value: s[1]}
}
