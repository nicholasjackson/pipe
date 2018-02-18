package worker

/*
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
*/
