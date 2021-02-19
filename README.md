# Limit Requests

This is a simple demo how to limit HTTP requests per sender (distinguished by
its IPv4) per time frame. It uses Go's concurrency features (goroutines,
channels, mutexes).

## Demo

For demo purposes, the timeout has to be set to five seconds.

Run it:

    $ go run limit_requests.go

Test it:

    $ curl http://localhost:8080 &
    $ curl http://localhost:8080 &
    $ curl http://localhost:8080 &
    $ curl http://localhost:8080 &

Output:

    OK, 1 requests served
    OK, 2 requests served
    timeout, 1 requests timed out
    timeout, 2 requests timed out

Or test it using the demo script (one request per second):

    $ ./spam.sh

Output (non-deterministic):

    OK, 1 requests served
    OK, 2 requests served
    timeout, 1 requests timed out
    OK, 3 requests served
    timeout, 2 requests timed out
    OK, 4 requests served
    timeout, 3 requests timed out
    timeout, 4 requests timed out
    OK, 5 requests served
    timeout, 5 requests timed out

## User Experience

This request limiter has the following impact on the client (assuming a
server-side timeout of five seconds):

- If the client only sends one requests every five seconds or less, the requests
  will be served immediately (no delay).
- If the client sends a second requests after three seconds, it has to wait two
  seconds for the second request.
- If the client sends three requests at once, the first request will be served
  immediately, the second (or third) will be served after five seconds, and the
  third (or second) will run into a timeout.

Different timeouts can be used for different endpoints. For example, `GET`
endpoints are not limited at all. `POST` endpoints that create new entries are
only allowed once in ten seconds. `DELETE` requests are only allowed once per
minute.
