package worker

import (
	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	stan "github.com/nats-io/go-nats-streaming"
	"github.com/nicholasjackson/faas-nats/client"
	"github.com/nicholasjackson/faas-nats/config"
	"github.com/nicholasjackson/faas-nats/template"
)

// NatsWorker is responsilbe for receiving requests and forwarding to nats
type NatsWorker struct {
	conn   NatsConnection
	client client.Client
	stats  *statsd.Client
	logger hclog.Logger
	subs   []*stan.Subscription
	parser *template.Parser
}

// NewNatsWorker creates a new worker
func NewNatsWorker(
	nc NatsConnection,
	c client.Client,
	stats *statsd.Client,
	logger hclog.Logger,
) *NatsWorker {

	template.Logger = logger.Named("templates")

	return &NatsWorker{
		conn:   nc,
		client: c,
		logger: logger,
		stats:  stats,
		subs:   make([]*stan.Subscription, 0),
		parser: &template.Parser{},
	}
}

// RegisterMessageListeners registers the messasge listeners
func (nw *NatsWorker) RegisterMessageListeners(c config.Config) {
	for _, f := range c.Functions {
		nw.logger.Info("Registering event", "event", f.Message)
		nw.stats.Incr("worker.event.register", []string{"message:" + f.Message}, 1)

		s, err := nw.conn.QueueSubscribe(f.Message, "queue."+f.Name, func(m *stan.Msg) {
			nw.handleMessage(f, m)
		})

		if err != nil {
			nw.logger.Error("Error registering queue", "error", err.Error())
			nw.stats.Incr("worker.register.failed", []string{"message:" + f.Message}, 1)

			return
		}

		nw.subs = append(nw.subs, &s)
	}
}

func (nw *NatsWorker) handleMessage(f config.Function, m *stan.Msg) {
	nw.logger.Info("Handle event", "subject", m.Subject)
	nw.stats.Incr("worker.event.handle", []string{"message:" + f.Message}, 1)

	nw.logger.Trace("Event Data", "data", string(m.Data))

	data, err := nw.processInputTemplate(f, m.Data)
	if err != nil {
		return
	}

	resp, err := nw.callFunction(f, data)
	if err != nil {
		return
	}

	// do we need to publish a success_message
	if f.SuccessMessage == "" {
		return
	}

	out, err := nw.processOutputTemplate(f, resp)
	if err != nil {
		return
	}

	nw.publishMessage(f, out)
}

func (nw *NatsWorker) processInputTemplate(f config.Function, data []byte) ([]byte, error) {
	// do we have a transformation template
	if f.Templates.InputTemplate != "" {
		functionData, err := nw.parser.Parse(f.Templates.InputTemplate, data)
		if err != nil {
			nw.logger.Error("Error processing intput template", "error", err)
			nw.stats.Incr("worker.event.error.inputtemplate", []string{"message:" + f.Message}, 1)

			return nil, err
		}

		nw.logger.Debug("Transformed input template", "template", f.Templates.OutputTemplate, "data", data)
		return functionData, err
	}

	return data, nil
}

func (nw *NatsWorker) processOutputTemplate(f config.Function, data []byte) ([]byte, error) {
	if f.Templates.OutputTemplate != "" {

		temp, err := nw.parser.Parse(f.Templates.OutputTemplate, data)
		if err != nil {
			nw.logger.Error("Error processing output template", "error", err)
			nw.stats.Incr(
				"worker.event.error.outputtemplate",
				[]string{"message:" + f.Message},
				1,
			)
			return nil, err
		}

		nw.logger.Debug("Transformed output template", "template", f.Templates.OutputTemplate, "data", data)
		return temp, err
	}

	return data, nil
}

func (nw *NatsWorker) callFunction(f config.Function, payload []byte) ([]byte, error) {
	nw.logger.Info("Calling function", "function", f.FunctionName)
	nw.logger.Debug("Function payload", "function", f.FunctionName, "payload", payload)

	resp, err := nw.client.CallFunction(f.FunctionName, f.Query, payload)
	if err != nil {
		nw.stats.Incr("worker.event.error.functioncall", []string{
			"message:" + f.Message,
			"function" + f.FunctionName,
		}, 1)
		nw.logger.Error("Error calling function", "error", err)

		return nil, err
	}

	nw.logger.Debug("Function response", "response", string(resp))
	return resp, nil
}

func (nw *NatsWorker) publishMessage(f config.Function, payload []byte) error {

	nw.logger.Info("Publishing message", "message", f.SuccessMessage)
	nw.logger.Debug("Publishing message", "message", f.SuccessMessage, "payload", payload)

	nw.stats.Incr(
		"worker.event.sendoutput",
		[]string{
			"message:" + f.Message,
			"output:" + f.SuccessMessage,
		},
		1,
	)

	return nw.conn.Publish(f.SuccessMessage, payload)
}
