A spike to try to understand what different monitoring options give us.

**You almost certainly need something different. Please stop and understand your
context.**

However, the premise is simple:

A server that speaks HTTP, and talks HTTP to upstream dependencies. As part of
doing that, it should capture:

1. A counter that is incremented for every request received
1. A counter that is incremented for every successful response
1. A counter that is incremented for every error response
1. A histogram of service time for successful responses
1. A histogram of service time for error responses
1. A gauge of the number of in-flight requests (this is a nice-to-have. Easier
   in stateful processes like Python, the JVM, Go etc. )

## Dependencies

- This is written in Go. Either you'll need Go, or Docker
- It uses Make for building things. That might be handy
- [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) is used to format and rearrange imports – `go get golang.org/x/tools/cmd/goimports`
- [vegeta](https://github.com/tsenart/vegeta) is used for sending a load of traffic to the server – `brew install vegeta` or similar

## Running it

You'll want to run the server, and throw some traffic at it so that some metrics
are created in the system you're playing with.

### Running the server

You'll need to run the server. You can do this in a couple of ways:

1. Using Go:
```sh
$ make
$ env FOO=your-api-key ./monitoring-spike &
```
1. Using Docker
```sh
$ make dockerise
$ docker run -p 8080:8080 -e FOO=your-api-key -itd jabley/monitoring-spike
```

### Generating some load

```sh
$ make report
```

This will run `vegeta` and hopefully create some metrics in the system that
you're evaluating.

## PS

Look at the branches!
