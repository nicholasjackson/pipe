package nats

import stan "github.com/nats-io/go-nats-streaming"

//go:generate moq -out mock_connection.go . Connection
// Connection  defines the behaviour required for the Nats connection
type Connection interface {
	QueueSubscribe(
		subject,
		qgroup string,
		cb stan.MsgHandler,
		opts ...stan.SubscriptionOption) (stan.Subscription, error)
	Publish(subj string, data []byte) error
}
