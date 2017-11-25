package worker

import (
	"log"

	"github.com/nats-io/nats"
	"github.com/nicholasjackson/faas-nats/client"
	"github.com/nicholasjackson/faas-nats/config"
	"github.com/nicholasjackson/faas-nats/template"
)

// NatsWorker is responsilbe for receiving requests and forwarding to nats
type NatsWorker struct {
	conn   NatsConnection
	client client.Client
	subs   []*nats.Subscription
	parser *template.Parser
}

// NewNatsWorker creates a new worker
func NewNatsWorker(nc NatsConnection, c client.Client) *NatsWorker {
	return &NatsWorker{
		conn:   nc,
		client: c,
		subs:   make([]*nats.Subscription, 0),
		parser: &template.Parser{},
	}
}

// RegisterMessageListeners registers the messasge listeners
func (nw *NatsWorker) RegisterMessageListeners(c config.Config) {
	for _, f := range c.Functions {
		log.Println("Registering event", f.Message)

		s, err := nw.conn.QueueSubscribe(f.Message, "queue."+f.Name, func(m *nats.Msg) {
			log.Println("Handle event:", m.Subject)

			functionData := m.Data

			// do we have a transformation template
			if f.Templates.InputTemplate != "" {
				var err error
				functionData, err = nw.parser.Parse(f.Templates.InputTemplate, m.Data)
				if err != nil {
					log.Println("Error processing intput template:", err)
					return
				}
			}

			log.Printf("Calling function: %s, with payload %s\n", f.FunctionName, functionData)
			resp, err := nw.client.CallFunction(f.FunctionName, functionData)
			if err != nil {
				log.Println("Error calling function:", err)
				return
			}

			// do we need to publish a success_message
			if f.SuccessMessage != "" {
				returnData := resp

				if f.Templates.OutputTemplate != "" {
					var err error
					returnData, err = nw.parser.Parse(f.Templates.OutputTemplate, resp)
					if err != nil {
						log.Println("Error processing output template:", err)
						return
					}
				}

				log.Printf("Publishing message: %s with payload %s\n", f.SuccessMessage, returnData)
				nw.conn.Publish(f.SuccessMessage, returnData)
			}
		})

		if err != nil {
			log.Println("Error registering queue")
			return
		}

		nw.subs = append(nw.subs, s)
	}
}
