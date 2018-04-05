package web

import (
	"fmt"
	"time"
)

//go:generate moq -out mock_connection_pool.go . ConnectionPool
// ConnectionPool defines the interface for http ConnectionPools
type ConnectionPool interface {
	GetConnection(bindAddr string, port int) (Connection, error)
}

// HTTPConnectionPool is a concrete implementation of ConnectionPool
type HTTPConnectionPool struct {
	healthCheckInterval time.Duration
	healthCheckMax      int
	connectionFactory   func(bindAddr string, port int) Connection
	connections         map[string]Connection
}

func NewHTTPConnectionPool() *HTTPConnectionPool {
	return &HTTPConnectionPool{
		healthCheckInterval: 500 * time.Millisecond,
		healthCheckMax:      10,
		connectionFactory:   createConnection,
		connections:         make(map[string]Connection),
	}
}

func createConnection(bindAddr string, port int) Connection {
	return NewHTTPConnection(bindAddr, port)
}

// GetConnection return a http server from the pool if one exists, or creates a new http server and starts
// listening.
// This method needs a complete update to take into account https
func (h *HTTPConnectionPool) GetConnection(bindAddr string, port int) (Connection, error) {
	key := fmt.Sprintf("%s_%d", bindAddr, port)
	if c := h.connections[key]; c != nil {
		return c, nil
	}

	// Need to add TLS setup
	s := h.connectionFactory(bindAddr, port)
	errChan := make(chan error)
	startedChan := make(chan bool)

	// need to see if returns an error
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		var lastError error
		for try := 0; try < h.healthCheckMax; try++ {

			// perform a health check
			err := s.CheckHealth()
			if err == nil {
				startedChan <- true
				return
			}

			lastError = err
			time.Sleep(h.healthCheckInterval) // wait before running another health check
		}

		errChan <- fmt.Errorf("Error starting server: %s", lastError.Error())
	}()

	select {
	case e := <-errChan:
		return nil, e
	case <-startedChan:
		h.connections[key] = s
		return s, nil
	}
}
