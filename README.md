A silly spike to try to understand what different monitoring options give us.

You almost certainly need something different.

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
