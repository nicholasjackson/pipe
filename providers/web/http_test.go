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
	"github.com/nicholasjackson/pipe/logger"
	"github.com/nicholasjackson/pipe/providers"
)

var serverResponse []byte

func setupHTTPProvider(t *testing.T, direction string) (*is.I, *HTTPProvider, *ConnectionMock, *ConnectionPoolMock, func()) {
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
		GetConnectionFunc: func(server string, port int, log logger.Logger) (Connection, error) {
			return mockedConnection, nil
		},
	}

	stats, _ := statsd.New("http://localhost:8125")
	log := hclog.Default()

	p := &HTTPProvider{
		direction: direction,
		Protocol:  u.Scheme,
		Server:    u.Hostname(),
		Port:      port,
		Path:      u.Path,
		pool:      mockedConnectionPool,
		log:       logger.New(log, stats),
	}
	p.Setup()

	return is, p, mockedConnection, mockedConnectionPool, func() {
		httptest.Close()
	}
}

func TestNewSetsDefaults(t *testing.T) {
	is := is.New(t)

	p := NewHTTPProvider("test", "output", nil, nil)

	is.Equal("POST", p.Method)   // should have set the method to post
	is.Equal(80, p.Port)         // should have set the default port to 80
	is.Equal("/", p.Path)        // should have set the default path to /
	is.Equal("http", p.Protocol) // should have set the default protocol to http
}

func TestSetupWithInboundTypeDoesGetsConnection(t *testing.T) {
	is, _, _, mcp, cleanup := setupHTTPProvider(t, providers.DirectionInput)
	defer cleanup()

	is.Equal(1, len(mcp.GetConnectionCalls())) // should have retrieved a connection from the connection pool
}

func TestSetupWithOutboundTypeDoesNotGetConnection(t *testing.T) {
	is, _, _, mcp, cleanup := setupHTTPProvider(t, providers.DirectionOutput)
	defer cleanup()

	is.Equal(0, len(mcp.GetConnectionCalls())) // should not have retrieved a connection from the connection pool
}

func TestPublishCallsEndpointWithData(t *testing.T) {
	is, p, _, _, cleanup := setupHTTPProvider(t, providers.DirectionInput)
	defer cleanup()
	payload := providers.Message{Data: []byte("test data")}

	_, err := p.Publish(payload)

	is.NoErr(err)                                          // should not have reuturned an error
	is.Equal(string(payload.Data), string(serverResponse)) // should have sent the correct payload
}

func TestPublishCallsEndpointAndReturnsBody(t *testing.T) {
	is, p, _, _, cleanup := setupHTTPProvider(t, providers.DirectionInput)
	defer cleanup()
	payload := providers.Message{Data: []byte("test data")}

	data, err := p.Publish(payload)

	is.NoErr(err)                     // should not have reuturned an error
	is.Equal("ok", string(data.Data)) // should have sent the correct payload
}

func TestListenWithOutboundTypeDoesNotCallListenPath(t *testing.T) {
	is, p, mc, _, cleanup := setupHTTPProvider(t, providers.DirectionOutput)
	p.direction = "output"
	defer cleanup()

	msgs, err := p.Listen()
	is.NoErr(err)        // should not have returned an error
	is.True(msgs == nil) // should have returned a nil message channel

	is.Equal(0, len(mc.ListenPathCalls())) // should not call listen path for output providers
}

func TestListenReturnsEvents(t *testing.T) {
	is, p, mc, _, cleanup := setupHTTPProvider(t, providers.DirectionInput)
	defer cleanup()

	contentType := "application/json"
	msgs, err := p.Listen()
	is.NoErr(err) // should not have returned an error

	go func() {
		rw := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte("abc")))
		r.Header.Set("content-type", contentType)

		h := mc.ListenPathCalls()[0].Handler
		h.ServeHTTP(rw, r)
	}()

	select {
	case m := <-msgs:
		is.Equal("abc", string(m.Data))                     // message data should be equal
		is.Equal(false, m.Redelivered)                      // message should have redelivered set
		is.Equal(uint64(1), m.Sequence)                     // message should have sequence set
		is.True(time.Now().UnixNano()-m.Timestamp < 100000) // message should have timestamp set
		is.True(len(m.ID) > 1)                              // message should have a message id
		is.Equal(m.ContentType, contentType)                // message should have content type
	case <-time.After(3 * time.Second):
		is.Fail() // message received timeout
	}
}
