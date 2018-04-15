package web

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/nicholasjackson/pipe/logger"
)

func setupPoolTests(t *testing.T) (*is.I, *HTTPConnectionPool, *ConnectionMock) {
	cm := &ConnectionMock{
		AddrFunc: func() string {
			panic("TODO: mock out the Addr method")
		},
		CheckHealthFunc: func() error {
			return nil
		},
		ListenAndServeFunc: func() error {
			return nil
		},
		ListenPathFunc: func(path string, method string, handler http.HandlerFunc) error {
			return nil
		},
		ShutdownFunc: func(ctx context.Context) {
		},
	}

	is := is.New(t)
	p := NewHTTPConnectionPool()

	// speed up the test execution by reducing the backoff for health checks
	p.healthCheckInterval = 1 * time.Millisecond

	// replace the connection factory with one returning our mock
	p.connectionFactory = func(httpAddr string, port int, logger logger.Logger) Connection {
		return cm
	}

	return is, p, cm
}

func TestCreatesNewConnectionFromEmptyPool(t *testing.T) {
	is, p, cm := setupPoolTests(t)

	c, err := p.GetConnection("localhost", 8080, nil)

	is.NoErr(err)                              // should not have returned an error
	is.Equal(c, cm)                            // should have created a new server
	is.Equal(1, len(cm.ListenAndServeCalls())) // should have called listen and serve
}

func TestGetsExistingConnectionFromPool(t *testing.T) {
	is, p, cm := setupPoolTests(t)

	c1, err := p.GetConnection("localhost", 8080, nil)
	is.NoErr(err) // should not have returned an error

	c2, err := p.GetConnection("localhost", 8080, nil)
	is.NoErr(err) // should not have returned an error

	is.True(c1 != nil)                         // connection should not be nil
	is.Equal(c1, cm)                           // connection 1 should be the same object returned from the factory
	is.Equal(c1, c2)                           // connection 1 and 2 should be the same object
	is.Equal(1, len(cm.ListenAndServeCalls())) // should have called listen and serve
}

func TestReturnsErrorIfBindNotPossible(t *testing.T) {
	is, p, cm := setupPoolTests(t)
	cm.ListenAndServeFunc = func() error {
		return fmt.Errorf("Unable to start")
	}
	cm.CheckHealthFunc = func() error {
		return fmt.Errorf("Unhealthy")
	}

	_, err := p.GetConnection("localhost", 8080, nil)

	is.Equal("Unable to start", err.Error()) // should have returned an error
}

func TestReturnsErrorOnNewConnectionFailedHealthCheck(t *testing.T) {
	is, p, cm := setupPoolTests(t)
	cm.CheckHealthFunc = func() error {
		return fmt.Errorf("Unhealthy")
	}

	_, err := p.GetConnection("localhost", 8080, nil)

	is.True(err != nil)                                       // should have returned an error
	is.Equal("Error starting server: Unhealthy", err.Error()) // should have returned an error
}

func TestCallsHealthCheckEqualTohealthCheckMax(t *testing.T) {
	is, p, cm := setupPoolTests(t)
	cm.CheckHealthFunc = func() error {
		return fmt.Errorf("Unhealthy")
	}

	p.GetConnection("localhost", 8080, nil)

	is.Equal(10, len(cm.CheckHealthCalls())) // should have called the health check 10 times
}
