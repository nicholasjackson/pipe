package nats

import (
	"fmt"

	stan "github.com/nats-io/go-nats-streaming"
	"github.com/nicholasjackson/pipe/logger"
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
	log          logger.Logger
	pool         ConnectionPool
	msgChannel   chan *providers.Message
}

func NewStreamingProvider(
	name, direction string,
	cp ConnectionPool,
	l logger.Logger) *StreamingProvider {
	return &StreamingProvider{
		name:      name,
		direction: direction,
		pool:      cp,
		log:       l,
	}
}

func (sp *StreamingProvider) Name() string {
	return sp.name
}

func (sp *StreamingProvider) Type() string {
	return "nats_queue"
}

func (sp *StreamingProvider) Direction() string {
	return sp.direction
}

func (sp *StreamingProvider) Setup() error {
	conn, err := sp.pool.GetConnection(sp.Server, sp.ClusterID)
	if err != nil {
		sp.log.ProviderConnectionFailed(sp, err)
		return err
	}

	sp.log.ProviderConnectionCreated(sp)
	sp.connection = conn

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
		sp.log.ProviderSubcriptionFailed(sp, err)
		return nil, err
	}

	sp.log.ProviderSubcriptionCreated(sp)
	sp.subscription = subscription

	sp.msgChannel = make(chan *providers.Message)

	return sp.msgChannel, nil
}

// Publish a message to the configured outbound queue
func (sp *StreamingProvider) Publish(msg providers.Message) (providers.Message, error) {
	sp.log.ProviderMessagePublished(sp, &msg, []interface{}{"queue", sp.Queue}...)

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
