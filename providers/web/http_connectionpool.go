package web

import (
	"fmt"
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

// GetConnection return a http server from the pool if one exists, or creates a new http server and starts
// listening.
// This method needs a complete update to take into account https
func (h *HTTPConnectionPool) GetConnection(bindAddr string, port int) (Connection, error) {
	key := fmt.Sprintf("%s_%d", bindAddr, port)
	if c := h.connections[key]; c != nil {
		return c, nil
	}

	s := NewHTTPConnection(bindAddr, port)

	go s.ListenAndServe()

	return s, nil
}
