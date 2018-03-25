package nats

import (
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	stan "github.com/nats-io/go-nats-streaming"
	"github.com/nicholasjackson/pipe/providers"
)

type StreamingProvider struct {
	name      string
	direction string

	Server    string               `hcl:"server"`
	ClusterID string               `hcl:"cluster_id"`
	Queue     string               `hcl:"queue"`
	AuthBasic *providers.AuthBasic `hcl:"auth_basic,block"`
	AuthMTLS  *providers.AuthMTLS  `hcl:"auth_mtls,block"`

	connection   Connection
	subscription stan.Subscription
	stats        *statsd.Client
	logger       hclog.Logger
	msgChannel   chan *providers.Message
}

func NewStreamingProvider(name, direction string) *StreamingProvider {
	return &StreamingProvider{name: name, direction: direction}
}

func (sp *StreamingProvider) Type() string {
	return "nats_queue"
}

func (sp *StreamingProvider) Name() string {
	return sp.name
}

func (sp *StreamingProvider) Direction() string {
	return sp.direction
}

func (sp *StreamingProvider) Setup(cp providers.ConnectionPool, logger hclog.Logger, stats *statsd.Client) error {
	pool := cp.(ConnectionPool)

	conn, err := pool.GetConnection(sp.Server, sp.ClusterID)
	if err != nil {
		stats.Incr("connection.nats.failed", nil, 1)
		logger.Error("Unable to connect to nats server", "error", err)
		return err
	}

	stats.Incr("connection.nats.created", nil, 1)
	logger.Debug("Created connection for", sp.Server, sp.ClusterID)

	sp.connection = conn
	sp.stats = stats
	sp.logger = logger

	return nil
}

func (sp *StreamingProvider) Listen() (<-chan *providers.Message, error) {
	// only listen if this is an input provider
	if sp.direction == providers.DirectionOutput {
		return nil, nil
	}

	qGroup := fmt.Sprintf("%s-%s", sp.Queue, sp.name)
	subscription, err := sp.connection.QueueSubscribe(sp.Queue, qGroup, sp.messageHandler)
	if err != nil {
		sp.stats.Incr("subscription.nats.failed", nil, 1)
		sp.logger.Error("Failed to create subscription for", sp.Queue)
		return nil, err
	}

	sp.stats.Incr("subscription.nats.created", nil, 1)
	sp.logger.Debug("Created subscription for", "queue", sp.Queue)
	sp.subscription = subscription

	sp.msgChannel = make(chan *providers.Message)

	return sp.msgChannel, nil
}

// Publish a message to the configured outbound queue
func (sp *StreamingProvider) Publish(msg providers.Message) (providers.Message, error) {
	sp.logger.Debug("Publishing message", "id", msg.ID, "parentid", msg.ParentID, "name", sp.name, "subject", sp.Queue)
	sp.stats.Incr("publish.nats.call", nil, 1)
	return providers.Message{}, sp.connection.Publish(sp.Queue, msg.Data)
}

func (sp *StreamingProvider) Stop() error {
	sp.subscription.Close()
	sp.subscription.Unsubscribe()

	return nil
}

func (sp *StreamingProvider) messageHandler(msg *stan.Msg) {
	m := providers.NewMessage()

	m.Data = msg.Data
	m.Redelivered = msg.Redelivered
	m.Timestamp = msg.Timestamp
	m.Sequence = msg.Sequence

	sp.msgChannel <- &m
}
