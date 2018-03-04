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
	// time the length of the message handling
	defer func(st time.Time) {
		p.statsd.Timing("handler.message.called", time.Now().Sub(st), []string{"pipe:" + pi.Name}, 1)
	}(time.Now())

	// ensure we do not process expired messages
	if time.Now().Sub(time.Unix(0, m.Timestamp)) > pi.ExpirationDuration {
		p.logger.Info("Message expired", "pipe", pi.Name, "timestamp", m.Timestamp, "expiration", pi.ExpirationDuration)
		p.statsd.Incr("handler.message.expired", []string{"pipe:" + pi.Name}, 1)

		return
	}

	// transform data if necessary
	data, err := p.processOutputTemplate(pi.Action, m.Data)
	if err != nil {
		return
	}

	p.logger.Info("Publish message action", "pipe", pi.Name, "output", pi.Action.Output)
	p.statsd.Incr("handler.message.action.publish", []string{"pipe:" + pi.Name}, 1)

	_, err = pi.Action.OutputProvider.Publish(data)
	if err != nil {
		p.logger.Error("Publish message action failed", "pipe", pi.Name, "error", err, "data", data)
		p.statsd.Incr("handler.message.action.publish.failed", []string{"pipe:" + pi.Name}, 1)

		p.publishFail(pi, m)
		return
	}

	p.logger.Info("Publish message action succeded", "pipe", pi.Name, "output", pi.Action.Output)
	p.statsd.Incr("handler.message.action.publish.success", []string{"pipe:" + pi.Name}, 1)

	p.publishSuccess(pi, m)
}

func (p *PipeServer) publishSuccess(pi *pipe.Pipe, m *providers.Message) {
	// process success messages
	for _, a := range pi.OnSuccess {
		// transform data if necessary
		p.logger.Info("Attempt process success action", "pipe", pi.Name, "output", a.Output)

		data, err := p.processOutputTemplate(a, m.Data)
		if err != nil {
			continue
		}

		_, err = a.OutputProvider.Publish(data)
		if err != nil {
			p.logger.Error("Publish success action failed", "pipe", pi.Name, "output", a.Output, "error", err, "data", data)
			p.statsd.Incr("handler.message.success.publish.failed", []string{"pipe:" + pi.Name, "output:" + a.Output}, 1)
			continue
		}

		p.logger.Info("Publish success action succeded", "pipe", pi.Name, "output", pi.Action.Output)
		p.statsd.Incr("handler.message.success.publish.success", []string{"pipe:" + pi.Name, "output:" + a.Output}, 1)
	}
}

func (p *PipeServer) publishFail(pi *pipe.Pipe, m *providers.Message) {
	// process success messages
	for _, a := range pi.OnFail {
		// transform data if necessary
		p.logger.Info("Attempt process fail action", "pipe", pi.Name, "output", a.Output)

		data, err := p.processOutputTemplate(a, m.Data)
		if err != nil {
			continue
		}

		a.OutputProvider.Publish(data)
		if err != nil {
			p.logger.Error("Publish success action failed", "pipe", pi.Name, "output", a.Output, "error", err, "data", data)
			p.statsd.Incr("handler.message.success.publish.failed", []string{"pipe:" + pi.Name, "output:" + a.Output}, 1)
			continue
		}

		p.logger.Info("Publish success action succeded", "pipe", pi.Name, "output", pi.Action.Output)
		p.statsd.Incr("handler.message.success.publish.success", []string{"pipe:" + pi.Name, "output:" + a.Output}, 1)
	}
}

func (p *PipeServer) processOutputTemplate(a pipe.Action, data []byte) ([]byte, error) {
	// do we have a transformation template
	if a.Template != "" {
		p.logger.Debug("Transform output template", "output", a.Output, "template", a.Template, "data", data)

		functionData, err := p.parser.Parse(a.Template, data)
		if err != nil {
			p.logger.Error("Error processing output template", "output", a.Output, "error", err)
			p.statsd.Incr("handler.message.template.failed", []string{"output:" + a.Output}, 1)

			return nil, err
		}

		p.statsd.Incr("handler.message.template.success", []string{"output:" + a.Output}, 1)
		p.logger.Debug("Transformed input template", "output", a.Output, "template")
		return functionData, err
	}

	return data, nil
}
