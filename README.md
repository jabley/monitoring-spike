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

- This is written in Go.
- It uses Make for building things. That might be handy
- It uses Docker Compose to run the applications
- [vegeta](https://github.com/tsenart/vegeta) is used for sending a load of traffic to the server – `brew install vegeta` or similar

## Running it

You'll want to run the applications, and throw some traffic at it so that some
metrics are created in the system you're playing with.

```sh
$ DURATION=30s make report
```

This will:

* use Docker to build everything
* use locally installed `vegeta` to send some traffic to the applications for 30 seconds
* print out a report of how the frontend behaved

## PS

Look at the branches!
