package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net/http"
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

func mainHandler(client *http.Client, backends []backend) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		process(client, "/home", backends, rw, r)
	})
}

func process(client *http.Client, path string, backends []backend, rw http.ResponseWriter, r *http.Request) {
	results := make(chan KeyValue, len(backends))

	var wg sync.WaitGroup

	for _, b := range backends {
		wg.Add(1)

		go func(b backend, results chan<- KeyValue) {
			defer wg.Done()
			// TODO(jabley): capture the response time
			// start := time.Now()
			// defer doSomething(b, time.Since(start))
			fetch(client, path, b, results)
		}(b, results)
	}

	wg.Wait()

	values := make([]KeyValue, len(backends))
	for i := range values {
		values[i] = <-results
	}

	rw.Header().Set("Content-Type", "text/html; charset=UTF-8")
	rw.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")

	if err := tmpl.Execute(rw, values); err != nil {
		panic(err)
	}
}

func fetch(client *http.Client, path string, b backend, results chan<- KeyValue) {
	URL := "http://" + b.address + path
	res, err := client.Get(URL)

	if err != nil {
		results <- KeyValue{b.name, err.Error()}
	} else {
		defer res.Body.Close()
		results <- KeyValue{b.name, res.Status}
	}
}

func productListing(client *http.Client, backends []backend) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		process(client, "/products", backends, rw, r)
	})
}

func productDetail(client *http.Client, backends []backend) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		process(client, "/products/"+hash(r.URL.Path), backends, rw, r)
	})
}

func categoryListing(client *http.Client, backends []backend) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		process(client, "/categories", backends, rw, r)
	})
}

func categoryDetail(client *http.Client, backends []backend) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		process(client, "/categories/"+hash(r.URL.Path), backends, rw, r)
	})
}

func search(client *http.Client, backends []backend) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		process(client, "/search?q="+hash(r.URL.Path), backends, rw, r)
	})
}

func account(client *http.Client, backends []backend) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		process(client, "/account", backends, rw, r)
	})
}

func checkout(client *http.Client, backends []backend) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		process(client, "/checkout", backends, rw, r)
	})
}

func hash(s string) string {
	return fmt.Sprintf("%d", hashAsUint(s))
}

func hashAsUint(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// predictableResponseTime gives a broadly similar response time for a given URL.
// This is used to fake the processing time to talk to a database, etc.
// For a given set of URLs, we want predictable behaviour. This is to show that
// certain customers / etc are slow. We should be able to see in a monitoring
// that requests for certain resources are slow.
func predictableResponseTime(r *http.Request) {
	crc := hashAsUint(r.URL.Path)
	if crc%5 == 0 {
		// perturb the response time for this one in a repeatable fashion
		time.Sleep(time.Duration(rand.Intn(200)+200) * time.Millisecond)
	}

	// This is our fake normal service time
	time.Sleep(time.Duration(time.Duration(rand.Intn(20)) * time.Millisecond))
}

func unreliableHandler(percentageFailures int) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		predictableResponseTime(r)
		if rand.Intn(100) < percentageFailures {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{
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
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(`{
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
