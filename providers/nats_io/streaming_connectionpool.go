package nats

import (
	"fmt"
	"time"

	stan "github.com/nats-io/go-nats-streaming"
)

// StreamingConnectionPool implements a connection pool for Nats Streaming
type StreamingConnectionPool struct {
	connections map[string]Connection
}

func NewStreamingConnectionPool() *StreamingConnectionPool {
	return &StreamingConnectionPool{
		connections: make(map[string]Connection),
	}
}

// GetConnection returns a connection from the pool, if one does not exist it creates it
func (scp *StreamingConnectionPool) GetConnection(server, clusterID string) (Connection, error) {
	clientID := fmt.Sprintf("%s-%d", "pipe", time.Now().UnixNano())
	key := fmt.Sprintf("%s-%s", server, clusterID)

	if nc, ok := scp.connections[key]; ok {
		return nc, nil
	}

	nc, err := stan.Connect(clusterID, clientID, stan.NatsURL(server))
	if err != nil {
		return nil, err
	}
	scp.connections[key] = nc
	return nc, nil
}
