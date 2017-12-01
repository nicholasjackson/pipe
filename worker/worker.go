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
			nw.logger.Info("Handle event", "subject", m.Subject)
			nw.stats.Incr("worker.event.handle", []string{"message:" + f.Message}, 1)

			functionData := m.Data
			nw.logger.Debug("Event Data", "data", string(m.Data))

			// do we have a transformation template
			if f.Templates.InputTemplate != "" {
				var err error
				functionData, err = nw.parser.Parse(f.Templates.InputTemplate, m.Data)
				if err != nil {
					nw.logger.Error("Error processing intput template", "error", err)
					nw.stats.Incr("worker.event.error.inputtemplate", []string{"message:" + f.Message}, 1)

					return
				}
			}

			nw.logger.Debug("Calling function", "function", f.FunctionName, "payload", functionData)

			resp, err := nw.client.CallFunction(f.FunctionName, functionData)
			if err != nil {
				nw.stats.Incr("worker.event.error.functioncall", []string{
					"message:" + f.Message,
					"function" + f.FunctionName,
				}, 1)
				nw.logger.Error("Error calling function", "error", err)

				return
			}

			nw.logger.Debug("Function response", "response", string(resp))

			// do we need to publish a success_message
			if f.SuccessMessage != "" {
				returnData := resp

				if f.Templates.OutputTemplate != "" {

					var err error
					returnData, err = nw.parser.Parse(f.Templates.OutputTemplate, resp)
					if err != nil {
						nw.logger.Error("Error processing output template", "error", err)
						nw.stats.Incr(
							"worker.event.error.outputtemplate",
							[]string{"message:" + f.Message},
							1,
						)

						return
					}
				}

				nw.logger.Debug("Publishing message", "message", f.SuccessMessage, "payload", returnData)

				nw.stats.Incr(
					"worker.event.sendoutput",
					[]string{
						"message:" + f.Message,
						"output:" + f.SuccessMessage,
					},
					1,
				)
				nw.conn.Publish(f.SuccessMessage, returnData)
			}
		})

		if err != nil {
			nw.logger.Error("Error registering queue", "error", err.Error())
			nw.stats.Incr("worker.register.failed", []string{"message:" + f.Message}, 1)

			return
		}

		nw.subs = append(nw.subs, &s)
	}
}
