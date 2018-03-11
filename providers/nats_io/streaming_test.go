package nats

import (
	"testing"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/matryer/is"
	stan "github.com/nats-io/go-nats-streaming"
	"github.com/nicholasjackson/pipe/providers"
)

func setupStreamingProvider(t *testing.T, direction string) (*is.I, *StreamingProvider, *ConnectionMock, *SubscriptionMock, *ConnectionPoolMock) {
	is := is.New(t)

	mockedSubscription := &SubscriptionMock{
		CloseFunc: func() error {
			return nil
		},
		UnsubscribeFunc: func() error {
			return nil
		},
	}

	mockedConnection := &ConnectionMock{
		PublishFunc: func(subj string, data []byte) error {
			return nil
		},
		QueueSubscribeFunc: func(
			subject string,
			qgroup string,
			cb stan.MsgHandler,
			opts ...stan.SubscriptionOption) (stan.Subscription, error) {
			return mockedSubscription, nil
		},
	}

	p := &StreamingProvider{
		direction: direction,
		Queue:     "testqueue",
		name:      "testprovider",
	}

	mockedConnectionPool := &ConnectionPoolMock{
		GetConnectionFunc: func(server string, clusterID string) (Connection, error) {
			return mockedConnection, nil
		},
	}

	stats, _ := statsd.New("http://localhost:8125")
	p.Setup(mockedConnectionPool, hclog.Default(), stats)

	return is, p, mockedConnection, mockedSubscription, mockedConnectionPool
}

func TestTypeEqualsNatsQueue(t *testing.T) {
	is, p, _, _, _ := setupStreamingProvider(t, providers.DirectionInput)

	is.Equal("nats_queue", p.Type())
}

func TestAssignsConnectionFromPoolOnSetup(t *testing.T) {
	is, p, cm, _, _ := setupStreamingProvider(t, providers.DirectionInput)

	is.Equal(p.connection, cm) // should have set the connection to the connection mock
}

func TestListenRegistersToListenForMessagesOnAQueue(t *testing.T) {
	is, p, cm, _, _ := setupStreamingProvider(t, providers.DirectionInput)
	p.Listen()

	is.Equal(1, len(cm.QueueSubscribeCalls()))                       // should have subscribed to a queue
	is.Equal(p.Queue, cm.QueueSubscribeCalls()[0].Subject)           // should have subscribed to queuename
	is.Equal(p.Queue+"-"+p.name, cm.QueueSubscribeCalls()[0].Qgroup) // should have created a queuegroup
}

func TestDoesNotListenRegistersToListenForMessagesOnAQueueWhenOutput(t *testing.T) {
	is, p, cm, _, _ := setupStreamingProvider(t, providers.DirectionOutput)
	p.Listen()

	is.Equal(0, len(cm.QueueSubscribeCalls())) // should have subscribed to a queue
}

func TestNewMessagesOnAQueueAddMessageToTheListenChannel(t *testing.T) {
	is, p, _, _, _ := setupStreamingProvider(t, providers.DirectionInput)
	msgs, err := p.Listen()

	is.NoErr(err)

	go func() {
		m := &stan.Msg{}
		m.Data = []byte("abc")
		m.Redelivered = true
		m.Sequence = 1
		m.Timestamp = 1234141
		p.messageHandler(m)
	}()

	select {
	case m := <-msgs:
		is.Equal("abc", string(m.Data))       // message data should be equal
		is.True(m.Redelivered)                // message should have redelivered set
		is.Equal(uint64(1), m.Sequence)       // message should have sequence set
		is.Equal(int64(1234141), m.Timestamp) // message should have timestamp set
	case <-time.After(3 * time.Second):
		is.Fail() // message received timeout
	}
}

func TestSendAddsAMessageToTheOutboundQueue(t *testing.T) {
	is, p, cm, _, _ := setupStreamingProvider(t, providers.DirectionInput)

	p.Publish([]byte("1233"))

	is.Equal(1, len(cm.PublishCalls()))                 // should have called publish once
	is.Equal(p.Queue, cm.PublishCalls()[0].Subj)        // should have set the subject
	is.Equal("1233", string(cm.PublishCalls()[0].Data)) // should have set the payload
}

func TestStopStopsListeningAndCancelsTheQueue(t *testing.T) {
	is, p, _, ms, _ := setupStreamingProvider(t, providers.DirectionInput)
	p.Listen()
	p.Stop()

	is.Equal(1, len(ms.UnsubscribeCalls())) // should have unsubscribed
	is.Equal(1, len(ms.CloseCalls()))       // should have closed the subscription
}
