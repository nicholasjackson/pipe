package web

import (
	"fmt"
	"net/http"
)

//go:generate moq -out mock_connection_pool.go . ConnectionPool
// ConnectionPool defines the interface for http ConnectionPools
type ConnectionPool interface {
	GetConnection(server string, port int) (*http.Server, error)
}

// HTTPConnectionPool is a concrete implementation of ConnectionPool
type HTTPConnectionPool struct {
	connections map[string]*http.Server
}

func (h *HTTPConnectionPool) GetConnection(server string, port int) (*http.Server, error) {
	key := fmt.Sprintf("%s_%d", server, port)
	if c := h.connections[key]; c != nil {
		return c, nil
	}

	return nil, nil
}
