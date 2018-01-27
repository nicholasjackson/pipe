package worker

import (
	"time"

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
		nw.logger.Info("Registering event", "name", f.Name, "event", f.Message)
		nw.stats.Incr("worker.event.register", []string{"name:" + f.Name, "message:" + f.Message}, 1)

		exp := 48 * time.Hour
		if f.Expiration != "" {
			var err error
			exp, err = time.ParseDuration(f.Expiration)
			if err != nil {
				nw.logger.Error("Invalid duration for function", f.Name, f.Expiration)
				nw.stats.Incr("worker.register.failed", []string{"message:" + f.Message}, 1)
				continue
			}
		}

		func(f config.Function, expiration time.Duration) {
			s, err := nw.conn.QueueSubscribe(f.Message, "queue."+f.Name, func(m *stan.Msg) {
				nw.handleMessage(f, m, expiration)
			})

			if err != nil {
				nw.logger.Error("Error registering queue", "error", err.Error())
				nw.stats.Incr("worker.register.failed", []string{"message:" + f.Message}, 1)

				return
			}

			nw.subs = append(nw.subs, &s)
		}(f, exp)
	}
}

func (nw *NatsWorker) handleMessage(f config.Function, m *stan.Msg, expiration time.Duration) {
	nw.logger.Info("Handle event", "subject", m.Subject, "subscription", f.Name, "id", m.CRC32, "redelivered", m.Redelivered, "size", m.Size()/1000)
	nw.stats.Incr("worker.event.handle", []string{"message:" + f.Message}, 1)
	nw.logger.Debug("Event Data", "subscription", f.Name, "id", m.CRC32, "redelivered", m.Redelivered, "data", string(m.Data))

	// check expiration
	if time.Now().Sub(time.Unix(0, m.Timestamp)) > expiration {
		nw.logger.Info("Message expired", "subject", m.Subject, "timestamp", m.Timestamp, "expiration", expiration)
		nw.stats.Incr("worker.event.expired", []string{"message:" + f.Message}, 1)

		return
	}

	data, err := nw.processInputTemplate(f, m.Data)
	if err != nil {
		return
	}

	resp, err := nw.callFunction(f, data)
	if err != nil {
		return
	}

	// do we need to publish a success_message
	for _, m := range f.SuccessMessages {

		out, err := nw.processOutputTemplate(f, m, resp)
		if err != nil {
			return
		}

		nw.publishMessage(f, m, out)
	}

	return
}

func (nw *NatsWorker) processInputTemplate(f config.Function, data []byte) ([]byte, error) {
	// do we have a transformation template
	if f.InputTemplate != "" {
		functionData, err := nw.parser.Parse(f.InputTemplate, data)
		if err != nil {
			nw.logger.Error("Error processing intput template", "subscription", f.Name, "error", err)
			nw.stats.Incr("worker.event.error.inputtemplate", []string{"message:" + f.Message}, 1)

			return nil, err
		}

		nw.logger.Debug("Transformed input template", "subscription", f.Name, "template", f.InputTemplate, "data", data)
		return functionData, err
	}

	return data, nil
}

func (nw *NatsWorker) processOutputTemplate(f config.Function, s config.SuccessMessage, data []byte) ([]byte, error) {
	if s.OutputTemplate != "" {

		temp, err := nw.parser.Parse(s.OutputTemplate, data)
		if err != nil {
			nw.logger.Error("Error processing output template", "subscription", f.Name, "template", s.OutputTemplate, "error", err)
			nw.stats.Incr(
				"worker.event.error.outputtemplate",
				[]string{"message:" + f.Message},
				1,
			)
			return nil, err
		}

		nw.logger.Debug("Transformed output template", "subscription", f.Name, "template", s.OutputTemplate, "data", data)
		return temp, err
	}

	return data, nil
}

func (nw *NatsWorker) callFunction(f config.Function, payload []byte) ([]byte, error) {
	nw.logger.Info("Calling function", "subscription", f.Name, "function", f.FunctionName)
	nw.logger.Debug("Function payload", "subscription", f.Name, "function", f.FunctionName, "payload", payload)

	resp, err := nw.client.CallFunction(f.FunctionName, f.Query, payload)
	if err != nil {
		nw.stats.Incr("worker.event.error.functioncall", []string{
			"message:" + f.Message,
			"function" + f.FunctionName,
		}, 1)
		nw.logger.Error("Error calling function", "subscription", f.Name, "function", f.FunctionName, "error", err)

		return nil, err
	}

	nw.logger.Debug("Function response", "subscription", f.Name, "function", f.FunctionName, "response", string(resp))
	return resp, nil
}

func (nw *NatsWorker) publishMessage(f config.Function, s config.SuccessMessage, payload []byte) error {

	nw.logger.Info("Publishing message", "subscription", f.Name, "message", s.Name)
	nw.logger.Debug("Publishing message", "subscription", f.Name, "message", s.Name, "payload", payload)

	nw.stats.Incr(
		"worker.event.sendoutput",
		[]string{
			"message:" + f.Message,
			"output:" + s.Name,
		},
		1,
	)

	return nw.conn.Publish(s.Name, payload)
}
