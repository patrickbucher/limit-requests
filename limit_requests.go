package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	muUsers sync.Mutex
	users   = make(map[string]chan struct{})

	timeout = 5 * time.Second
)

// Wait ensures that only one request per client is served within the given
// timeout. For every client, which is distinguished by its IPv4 address, a
// token is produced once per timeout. The token is given to one of the waiting
// requests, and a new token is produced thereafter. A request either acquires
// a token within the given timeout, and the request is returned; or the
// request runs out of time, and an error is returned.
func Wait(r *http.Request, timeout time.Duration) (*http.Request, error) {
	// timeout after given time
	timeoutChan := make(chan struct{})
	go func() {
		time.Sleep(timeout)
		timeoutChan <- struct{}{}
	}()

	// every user has a channel that gets tokens
	user := ip4(r)
	muUsers.Lock()
	tokenChan, ok := users[user]
	if !ok {
		tokenChan = make(chan struct{})
		go func() {
			// the first token is served immediately
			tokenChan <- struct{}{}
		}()
		users[user] = tokenChan
	}
	muUsers.Unlock()

	// wait for timeout or token
	select {
	case <-tokenChan:
		// token acquired: request can be served, new token be spawned
		go func() {
			time.Sleep(timeout)
			tokenChan <- struct{}{}
		}()
		return r, nil
	case <-timeoutChan:
		// timeout: do not serve the request
		return nil, fmt.Errorf("one request per %v allowed", timeout)
	}
}

var (
	serveCounter sync.Mutex
	served       int

	timeoutCounter sync.Mutex
	timedOut       int
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r, err := Wait(r, timeout)
		if err != nil {
			fmt.Println(err)
			// NOTE: use http.StatusTooManyRequests; body message just for demo purposes
			timeoutCounter.Lock()
			timedOut++
			t := timedOut
			timeoutCounter.Unlock()
			w.Write([]byte(fmt.Sprintf("timeout, %d requests timed out\n", t)))
			return
		}
		fmt.Println("OK")
		serveCounter.Lock()
		served++
		c := served
		serveCounter.Unlock()
		w.Write([]byte(fmt.Sprintf("OK, %d requests served\n", c)))
	})
	http.ListenAndServe("0.0.0.0:8080", nil)
}

func ip4(r *http.Request) string {
	if !strings.Contains(r.RemoteAddr, ":") {
		return r.RemoteAddr
	}
	fields := strings.Split(r.RemoteAddr, ":")
	if len(fields) < 2 {
		return r.RemoteAddr
	}
	return fields[0]
}
