package worker

//go:generate moq -out mock_conn_test.go . NatsConnection
import stan "github.com/nats-io/go-nats-streaming"

// NatsConnection  defines the behaviour required for the Nats connection
type NatsConnection interface {
	QueueSubscribe(subject, qgroup string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error)
	Publish(subj string, data []byte) error
}
