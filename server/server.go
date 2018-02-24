package server

import (
	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/faas-nats/config"
	"github.com/nicholasjackson/faas-nats/providers"
)

// PipeServer is the main server which configures the providers and starts listening for messages
type PipeServer struct {
	config    config.Config
	logger    hclog.Logger
	statsd    *statsd.Client
	listeners map[string]<-chan *providers.Message
}

// New creates a new PipeServer
func New(c config.Config, l hclog.Logger, s *statsd.Client) *PipeServer {
	return &PipeServer{
		config:    c,
		logger:    l,
		statsd:    s,
		listeners: make(map[string]<-chan *providers.Message),
	}
}

// Listen starts listening for messages
func (p *PipeServer) Listen() {
	for _, i := range p.config.Inputs {
		i.Setup(p.config.ConnectionPools[i.Type()], p.logger, p.statsd)
	}

	for _, i := range p.config.Outputs {
		i.Setup(p.config.ConnectionPools[i.Type()], p.logger, p.statsd)
	}

	// listen for messages
	for _, i := range p.config.Inputs {
		c, err := i.Listen()
		if err != nil {
			p.logger.Error("Unable to listen for input", err)
		}

		p.listeners["abc"] = c
	}
}

// Stop listening for messages and shutdown connections
func (p *PipeServer) Stop() {
	// stop the providers
}
