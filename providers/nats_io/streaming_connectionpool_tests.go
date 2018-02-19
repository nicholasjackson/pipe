package nats

import (
	"testing"

	"github.com/matryer/is"
	stan "github.com/nats-io/go-nats-streaming"
)

func TestGetsConnectionFromPool(t *testing.T) {
	is := is.New(t)
	cp := StreamingConnectionPool{}
	mockedConnection := &ConnectionMock{
		PublishFunc: func(subj string, data []byte) error {
			panic("TODO: mock out the Publish method")
		},
		QueueSubscribeFunc: func(subject string, qgroup string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
			panic("TODO: mock out the QueueSubscribe method")
		},
	}
	cp.connections["test-testid"] = mockedConnection

	nc, err := cp.GetConnection("test", "testid")

	is.NoErr(err)
	is.Equal(mockedConnection, nc) // should have returned the connection from the cache
}

func TestCreatesConnectionWhenOneDoesNotExistInPool(t *testing.T) {
	is := is.New(t)
	cp := StreamingConnectionPool{}

	nc, err := cp.GetConnection("test", "testid")

	is.NoErr(err)
	is.True(nc != nil) // should have returned a connection
}
