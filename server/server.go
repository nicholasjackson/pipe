package server

import (
	"sync"
	"time"

	"github.com/nicholasjackson/pipe/config"
	"github.com/nicholasjackson/pipe/logger"
	"github.com/nicholasjackson/pipe/pipe"
	"github.com/nicholasjackson/pipe/providers"
	"github.com/nicholasjackson/pipe/template"
)

// PipeServer is the main server which configures the providers and starts listening for messages
type PipeServer struct {
	config       *config.Config
	logger       logger.Logger
	parser       *template.Parser
	startupGroup sync.WaitGroup
}

// New creates a new PipeServer
func New(c *config.Config, l logger.Logger) *PipeServer {

	return &PipeServer{
		config: c,
		logger: l,
		parser: &template.Parser{},
	}
}

// Listen starts listening for messages
func (p *PipeServer) Listen() {
	// setup the wait group
	p.startupGroup = sync.WaitGroup{}
	p.startupGroup.Add(len(p.config.Inputs))

	for _, i := range p.config.Inputs {
		i.Setup()
	}

	for _, i := range p.config.Outputs {
		i.Setup()
	}

	// setup listeners
	for _, i := range p.config.Inputs {
		go p.listen(i)
	}

	// do not return untill all inputs are listening
	p.startupGroup.Wait()
}

// Stop listening for messages and shutdown connections
func (p *PipeServer) Stop() {
	// stop the providers
}

func (p *PipeServer) listen(i providers.Provider) {
	c, err := i.Listen()
	if err != nil {
		p.logger.ServerUnableToListen(i, err)
		return
	}

	// decrement the startup group now this input is listening
	p.startupGroup.Done()

	for m := range c {
		pipes := p.getPipesByInputProvider(i)
		if len(pipes) < 1 {
			p.logger.ServerNoPipesConfigured(i)
		}

		for _, pi := range pipes {
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
	defer p.logger.ServerNewMessageReceivedStart(pi, m).Stop()

	// ensure we do not process expired messages
	if time.Now().Sub(time.Unix(0, m.Timestamp)) > pi.ExpirationDuration {
		p.logger.ServerHandleMessageExpired(pi, m)
		return
	}

	// transform data if necessary
	data, err := p.processOutputTemplate(pi.Action, m.Data)
	if err != nil {
		return
	}
	msg := providers.NewMessage()
	msg.Data = data
	msg.ParentID = m.ID

	p.logger.ServerActionPublish(pi, m)

	_, err = pi.Action.OutputProvider.Publish(msg)
	if err != nil {
		p.logger.ServerActionPublishFailed(pi, m, err)
		p.publishFail(pi, m)

		return
	}

	p.logger.ServerActionPublishSuccess(pi, m)

	p.publishSuccess(pi, m)
}

func (p *PipeServer) publishSuccess(pi *pipe.Pipe, m *providers.Message) {
	// process success messages
	for _, a := range pi.OnSuccess {
		p.logger.ServerSuccessPublish(pi, &a, m)

		// transform data if necessary
		data, err := p.processOutputTemplate(a, m.Data)
		if err != nil {
			continue
		}
		msg := providers.NewMessage()
		msg.ParentID = m.ID
		msg.Data = data

		_, err = a.OutputProvider.Publish(msg)
		if err != nil {
			p.logger.ServerSuccessPublishFailed(pi, &a, m, err)
			continue
		}

		p.logger.ServerSuccessPublishSuccess(pi, &a, m)
	}
}

func (p *PipeServer) publishFail(pi *pipe.Pipe, m *providers.Message) {
	// process success messages
	for _, a := range pi.OnFail {
		p.logger.ServerFailPublish(pi, &a, m)

		// transform data if necessary
		data, err := p.processOutputTemplate(a, m.Data)
		if err != nil {
			continue
		}
		msg := providers.NewMessage()
		msg.ParentID = m.ID
		msg.Data = data

		a.OutputProvider.Publish(msg)
		if err != nil {
			p.logger.ServerFailPublishFailed(pi, &a, m, err)
			continue
		}

		p.logger.ServerFailPublishSuccess(pi, &a, m)
	}
}

func (p *PipeServer) processOutputTemplate(a pipe.Action, data []byte) ([]byte, error) {
	// do we have a transformation template
	if a.Template != "" {
		defer p.logger.ServerTemplateProcessStart(&a, data).Stop()

		functionData, err := p.parser.Parse(a.Template, data)
		if err != nil {
			p.logger.ServerTemplateProcessFail(&a, data, err)
			return nil, err
		}

		return functionData, nil
	}

	return data, nil
}
