package web

import (
	"fmt"
	"net/http"
	"time"
)

//go:generate moq -out mock_connection_pool.go . ConnectionPool
// ConnectionPool defines the interface for http ConnectionPools
type ConnectionPool interface {
	GetConnection(bindAddr string, port int) (Connection, error)
}

// HTTPConnectionPool is a concrete implementation of ConnectionPool
type HTTPConnectionPool struct {
	connections map[string]Connection
}

func NewHTTPConnectionPool() *HTTPConnectionPool {
	return &HTTPConnectionPool{
		connections: make(map[string]Connection),
	}
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
	s := NewHTTPConnection(bindAddr, port)
	errChan := make(chan error)
	startedChan := make(chan bool)

	// need to see if returns an error
	go func() {
		err := s.ListenAndServe()
		errChan <- err
	}()

	go func() {
		var lastError error
		for try := 0; try < 10; try++ {

			// perform a health check
			resp, err := http.Get(fmt.Sprintf("http://%s:%d/_health", bindAddr, port))
			if err == nil && resp.StatusCode == http.StatusOK {
				startedChan <- true
				return
			}

			// if the health check returns anything other than 200 raise an error
			if resp.StatusCode != http.StatusOK {
				err = fmt.Errorf("Expected status %d from health check, got status %d", http.StatusOK, resp.StatusCode)
			}

			lastError = err
			time.Sleep(500 * time.Millisecond)
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
