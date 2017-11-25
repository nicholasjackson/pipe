package worker

import "github.com/nats-io/nats"

//go:generate moq -out mock_conn_test.go . NatsConnection

// NatsConnection  defines the behaviour required for the Nats connection
type NatsConnection interface {
	QueueSubscribe(subj, queue string, cb nats.MsgHandler) (*nats.Subscription, error)
	Publish(subj string, data []byte) error
}
