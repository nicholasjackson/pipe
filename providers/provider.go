package providers

import (
	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
)

//go:generate moq -out mock_connectionpool.go . ConnectionPool
type ConnectionPool interface {
}

const (
	DirectionInput  = "input"
	DirectionOutput = "output"
)

//go:generate moq -out mock_provider.go . Provider

// Provider defines a generic interface than an input or an output must implement
type Provider interface {
	Name() string
	Type() string                                                          // Type returns the type of the provider
	Direction() string                                                     // Direction returns input or output
	Setup(cp ConnectionPool, log hclog.Logger, stats *statsd.Client) error // Setup to initalize any connection for the provider
	Listen() (<-chan *Message, error)                                      // Listen for messages
	Publish(Message) (Message, error)                                      // Publish a message to the outbound provider
	Stop() error                                                           // Stop listening for messages
}
