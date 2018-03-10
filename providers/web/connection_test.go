package web

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/matryer/is"
)

func setupHTTPConnection(t *testing.T) (*is.I, Connection, func()) {
	is := is.New(t)
	c := NewHTTPConnection("localhost", 18999)

	go c.ListenAndServe()
	time.Sleep(10 * time.Millisecond)

	return is, c, func() {
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
		c.Shutdown(ctx)
	}
}

func TestConnectionListensAtPath(t *testing.T) {
	is, c, cleanup := setupHTTPConnection(t)
	defer cleanup()

	var called bool
	err := c.ListenPath("/", http.MethodGet, func(rw http.ResponseWriter, r *http.Request) {
		called = true
	})

	is.NoErr(err) // ListenPath should not have returned an error

	httpC := http.DefaultClient
	httpC.Timeout = 1 * time.Second
	_, err = httpC.Get("http://localhost:18999")

	is.NoErr(err)   // calling endpoint should not return an error
	is.True(called) // should have called the handler
}

func TestConnectionNotListensToGetMethodWhenMethodPost(t *testing.T) {
	is, c, cleanup := setupHTTPConnection(t)
	defer cleanup()

	var called bool
	err := c.ListenPath("/", http.MethodGet, func(rw http.ResponseWriter, r *http.Request) {
		called = true
	})

	is.NoErr(err) // ListenPath should not have returned an error

	httpC := http.DefaultClient
	httpC.Timeout = 1 * time.Second
	_, err = httpC.Post("http://localhost:18999", "text/plain", bytes.NewBuffer([]byte{}))

	is.NoErr(err)           // calling endpoint should not return an error
	is.Equal(false, called) // should not have called the handler
}
