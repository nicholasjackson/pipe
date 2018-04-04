package web

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/matryer/is"
)

func setupHTTPConnection(t *testing.T) (*is.I, Connection, int, func()) {
	is := is.New(t)
	port := rand.Intn(10000) + 50000 // generate random port between 50000 - 60000
	c := NewHTTPConnection("localhost", port)

	go c.ListenAndServe()
	time.Sleep(10 * time.Millisecond)

	return is, c, port, func() {
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
		c.Shutdown(ctx)
	}
}

func TestConnectionListensAtPath(t *testing.T) {
	is, c, p, cleanup := setupHTTPConnection(t)
	defer cleanup()

	var called bool
	err := c.ListenPath("/", http.MethodGet, func(rw http.ResponseWriter, r *http.Request) {
		called = true
	})

	is.NoErr(err) // ListenPath should not have returned an error

	httpC := http.DefaultClient
	httpC.Timeout = 1 * time.Second
	_, err = httpC.Get(fmt.Sprintf("http://localhost:%d", p))

	is.NoErr(err)   // calling endpoint should not return an error
	is.True(called) // should have called the handler
}

func TestConnectionNotListensToGetMethodWhenMethodPost(t *testing.T) {
	is, c, p, cleanup := setupHTTPConnection(t)
	defer cleanup()

	var called bool
	err := c.ListenPath("/", http.MethodGet, func(rw http.ResponseWriter, r *http.Request) {
		called = true
	})

	is.NoErr(err) // ListenPath should not have returned an error

	httpC := http.DefaultClient
	httpC.Timeout = 1 * time.Second
	_, err = httpC.Post(fmt.Sprintf("http://localhost:%d", p), "text/plain", bytes.NewBuffer([]byte{}))

	is.NoErr(err)           // calling endpoint should not return an error
	is.Equal(false, called) // should not have called the handler
}

func TestConnectionCreatesHealthHandler(t *testing.T) {
	is, _, p, cleanup := setupHTTPConnection(t)
	defer cleanup()

	httpC := http.DefaultClient
	httpC.Timeout = 1 * time.Second
	resp, err := httpC.Post(fmt.Sprintf("http://localhost:%d/_health", p), "text/plain", bytes.NewBuffer([]byte{}))

	is.NoErr(err)                            // calling endpoint should not return an error
	is.Equal(http.StatusOK, resp.StatusCode) // should have retuend status 200 for health
}
