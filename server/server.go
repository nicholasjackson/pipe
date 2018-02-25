package server

import (
	"time"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/pipe/config"
	"github.com/nicholasjackson/pipe/pipe"
	"github.com/nicholasjackson/pipe/providers"
	"github.com/nicholasjackson/pipe/template"
)

// PipeServer is the main server which configures the providers and starts listening for messages
type PipeServer struct {
	config config.Config
	logger hclog.Logger
	statsd *statsd.Client
	parser *template.Parser
}

// New creates a new PipeServer
func New(c config.Config, l hclog.Logger, s *statsd.Client) *PipeServer {

	return &PipeServer{
		config: c,
		logger: l,
		statsd: s,
		parser: &template.Parser{},
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

	// setup listeners
	for _, i := range p.config.Inputs {
		go p.listen(i)
	}
}

// Stop listening for messages and shutdown connections
func (p *PipeServer) Stop() {
	// stop the providers
}

func (p *PipeServer) listen(i providers.Provider) {
	c, err := i.Listen()
	if err != nil {
		p.logger.Error("Unable to listen for input", err)
		return
	}

	for m := range c {
		p.logger.Info("recieved message")

		for _, pi := range p.getPipesByInputProvider(i) {
			p.handleMessage(pi, m)
		}
	}
}

func (p *PipeServer) getPipesByInputProvider(i providers.Provider) []*pipe.Pipe {
	var pipes []*pipe.Pipe

	for _, pi := range p.config.Pipes {
		if pi.Input == i.Name() {
			pipes = append(pipes, pi)
		}
	}

	return pipes
}

func (p *PipeServer) handleMessage(pi *pipe.Pipe, m *providers.Message) {
	if time.Now().Sub(time.Unix(0, m.Timestamp)) > pi.ExpirationDuration {
		p.logger.Info("Message expired", "subject", pi.Name, "timestamp", m.Timestamp, "expiration", pi.ExpirationDuration)
		p.statsd.Incr("handler.message.expired", []string{"pipe:" + pi.Name}, 1)

		return
	}

	// transform data if necessary
	data, err := p.processInputTemplate(pi.Action, m.Data)
	if err != nil {
		return
	}

	pi.Action.OutputProvider.Publish(data)
}

func (p *PipeServer) processInputTemplate(a pipe.Action, data []byte) ([]byte, error) {
	// do we have a transformation template
	if a.Template != "" {
		functionData, err := p.parser.Parse(a.Template, data)
		if err != nil {
			p.logger.Error("Error processing input template", "output", a.Output, "error", err)
			p.statsd.Incr("handler.error.inputtemplate", []string{"output:" + a.Output}, 1)

			return nil, err
		}

		p.logger.Debug("Transformed input template", "output", a.Output, "template", a.Template, "data", data)
		return functionData, err
	}

	return data, nil
}
