package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx := newContextWithRequestID(r.Context(), r)
		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}

func newContextWithRequestID(ctx context.Context, r *http.Request) context.Context {
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = generateRandomID()
	}

	return context.WithValue(ctx, requestIDKey, reqID)
}

func serviceTime(next http.Handler) http.Handler {
	record := func(r *http.Request, duration time.Duration) {
		// TODO(jabley): send data to a metrics gathering service
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer record(r, time.Now().Sub(start))
		next.ServeHTTP(rw, r)
	})
}

func instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx := newInstrumentedContext(r.Context())
		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}

func newInstrumentedContext(ctx context.Context) context.Context {
	// TODO(jabley): add metrics gathering objects to the request context.
	return ctx
}

func mainHandler(backends []backend) http.Handler {
	tr := &http.Transport{
		ResponseHeaderTimeout: 2 * time.Second,
	}
	client := &http.Client{Transport: tr}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Frontend received request\n")

		results := make(chan KeyValue, len(backends))

		var wg sync.WaitGroup

		for _, b := range backends {
			wg.Add(1)

			go func(b backend, results chan<- KeyValue) {
				defer wg.Done()
				// TODO(jabley): capture the response time
				// start := time.Now()
				// defer doSomething(b, time.Since(start))
				fetch(client, b, results)
			}(b, results)
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
	})
}

func fetch(client *http.Client, b backend, results chan<- KeyValue) {
	fmt.Printf("Sending request to backend %s\n", b.name)

	res, err := client.Get("http://" + b.address)

	fmt.Printf("Received response from backend %s\n", b.name)

	if err != nil {
		results <- KeyValue{b.name, err.Error()}
	} else {
		defer res.Body.Close()
		results <- KeyValue{b.name, res.Status}
	}
}

func statusHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
		w.WriteHeader(http.StatusOK)
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		json.NewEncoder(w).Encode(mem)
	})
}

func unreliableHandler(percentageFailures int) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}