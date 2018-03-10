package web

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/matryer/is"
)

var serverResponse []byte

func setupHTTPProvider(t *testing.T) (*is.I, *HTTPProvider, *ConnectionMock, func()) {
	is := is.New(t)

	httptest := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			d, err := ioutil.ReadAll(r.Body)
			is.NoErr(err) // expected no error from http body

			serverResponse = d
			rw.Write([]byte("ok"))
		}),
	)

	u, _ := url.Parse(httptest.URL)
	port, _ := strconv.Atoi(u.Port())

	mockedConnection := &ConnectionMock{
		AddrFunc: func() string {
			panic("TODO: mock out the Addr method")
		},
		ListenAndServeFunc: func() error {
			panic("TODO: mock out the ListenAndServe method")
		},
		ListenPathFunc: func(path string, method string, handler http.HandlerFunc) error {
			return nil
		},
		ShutdownFunc: func(ctx context.Context) {
			panic("TODO: mock out the Shutdown method")
		},
	}

	mockedConnectionPool := &ConnectionPoolMock{
		GetConnectionFunc: func(server string, port int) (Connection, error) {
			return mockedConnection, nil
		},
	}

	p := &HTTPProvider{
		Protocol: u.Scheme,
		Server:   u.Hostname(),
		Port:     port,
		Path:     u.Path,
	}

	stats, _ := statsd.New("http://localhost:8125")
	p.Setup(mockedConnectionPool, hclog.Default(), stats)

	is.Equal(1, len(mockedConnectionPool.GetConnectionCalls())) // should have retrieved a connection from the connection pool

	return is, p, mockedConnection, func() {
		httptest.Close()
	}
}

func TestPublishCallsEndpointWithData(t *testing.T) {
	is, p, _, cleanup := setupHTTPProvider(t)
	defer cleanup()
	payload := []byte("test data")

	_, err := p.Publish(payload)

	is.NoErr(err)                     // should not have reuturned an error
	is.Equal(payload, serverResponse) // should have sent the correct payload
}

func TestPublishCallsEndpointAndReturnsBody(t *testing.T) {
	is, p, _, cleanup := setupHTTPProvider(t)
	defer cleanup()
	payload := []byte("test data")

	data, err := p.Publish(payload)

	is.NoErr(err)                // should not have reuturned an error
	is.Equal("ok", string(data)) // should have sent the correct payload
}

func TestListenReturnsEvents(t *testing.T) {
	is, p, mc, cleanup := setupHTTPProvider(t)
	defer cleanup()

	msgs, err := p.Listen()
	is.NoErr(err) // should not have returned an error

	go func() {
		rw := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", bytes.NewBuffer([]byte("abc")))
		h := mc.ListenPathCalls()[0].Handler
		h.ServeHTTP(rw, r)
	}()

	select {
	case m := <-msgs:
		is.Equal("abc", string(m.Data))                    // message data should be equal
		is.Equal(false, m.Redelivered)                     // message should have redelivered set
		is.Equal(uint64(1), m.Sequence)                    // message should have sequence set
		is.True(time.Now().UnixNano()-m.Timestamp < 10000) // message should have timestamp set
	case <-time.After(3 * time.Second):
		is.Fail() // message received timeout
	}
}
