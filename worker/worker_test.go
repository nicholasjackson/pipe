package worker

import (
	"testing"

	"github.com/matryer/is"
	stan "github.com/nats-io/go-nats-streaming"
	"github.com/nicholasjackson/faas-nats/client"
	"github.com/nicholasjackson/faas-nats/config"
)

var returnPayload []byte
var returnError error

func setupWorkerTests(t *testing.T) (*is.I, *NatsWorker, *NatsConnectionMock, *client.ClientMock) {
	mockedNatsConnection := &NatsConnectionMock{
		QueueSubscribeFunc: func(subject string, qgroup string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
			return nil, nil
		},
		PublishFunc: func(subj string, data []byte) error {
			return nil
		},
	}

	mockedClient := &client.ClientMock{
		CallFunctionFunc: func(name string, payload []byte) ([]byte, error) {
			return returnPayload, returnError
		},
	}

	return is.New(t), NewNatsWorker(mockedNatsConnection, mockedClient), mockedNatsConnection, mockedClient
}

func TestRegistersNMessageListeners(t *testing.T) {
	is, nw, mc, _ := setupWorkerTests(t)

	c := config.Config{
		Functions: []config.Function{
			config.Function{
				Name:    "test1",
				Message: "tests.message",
			},
			config.Function{
				Name:    "test2",
				Message: "tests.message",
			},
		},
	}

	nw.RegisterMessageListeners(c)

	is.Equal(2, len(mc.QueueSubscribeCalls())) // expected 2 call to subscribe
}

func TestRegistersMessageListenerWithCorrectDetails(t *testing.T) {
	is, nw, mc, _ := setupWorkerTests(t)

	c := config.Config{
		Functions: []config.Function{
			config.Function{
				Name:    "test1",
				Message: "tests.message",
			},
		},
	}

	nw.RegisterMessageListeners(c)

	is.Equal(1, len(mc.QueueSubscribeCalls()))                     // expected 1 call to subscribe
	is.Equal("queue.test1", mc.QueueSubscribeCalls()[0].Qgroup)    // expected queue to be set to correct value
	is.Equal("tests.message", mc.QueueSubscribeCalls()[0].Subject) // expected subject to be set to correct value
}

func TestWorkerCallsFunctionWithRawMessage(t *testing.T) {
	is, nw, mc, cl := setupWorkerTests(t)

	c := config.Config{
		Functions: []config.Function{
			config.Function{
				Name:    "test1",
				Message: "tests.message",
			},
		},
	}

	nw.RegisterMessageListeners(c)
	f := mc.QueueSubscribeCalls()[0].Cb

	msg := stan.Msg{}
	msg.Data = []byte("data")
	f(&msg) // call the function

	is.Equal(1, len(cl.CallFunctionCalls()))                    // expected 1 call to function
	is.Equal("data", string(cl.CallFunctionCalls()[0].Payload)) // expected raw payload to be passed
}

func TestWorkerCallsFunctionTransformingMessage(t *testing.T) {
	is, nw, mc, cl := setupWorkerTests(t)

	c := config.Config{
		Functions: []config.Function{
			config.Function{
				Name:    "test1",
				Message: "tests.message",
				Templates: config.Templates{
					InputTemplate: `{ "nicsname": "{{ .JSON.name }}" }`,
				},
			},
		},
	}

	nw.RegisterMessageListeners(c)
	f := mc.QueueSubscribeCalls()[0].Cb

	msg := stan.Msg{}
	msg.Data = []byte(`{ "name": "nic" }`)
	f(&msg) // call the function

	is.Equal(1, len(cl.CallFunctionCalls()))                                     // expected 1 call to function
	is.Equal(`{ "nicsname": "nic" }`, string(cl.CallFunctionCalls()[0].Payload)) // expected processed payload to be passed
}

func TestWorkerPublishesEventPostFunctionCall(t *testing.T) {
	is, nw, mc, _ := setupWorkerTests(t)

	c := config.Config{
		Functions: []config.Function{
			config.Function{
				Name:           "test1",
				Message:        "tests.message",
				SuccessMessage: "tests.message.success",
			},
		},
	}

	nw.RegisterMessageListeners(c)
	f := mc.QueueSubscribeCalls()[0].Cb

	msg := stan.Msg{}
	f(&msg) // call the function

	is.Equal(1, len(mc.PublishCalls()))
}

func TestWorkerPublishesEventPostFunctionCallTransformingPayload(t *testing.T) {
	is, nw, mc, _ := setupWorkerTests(t)

	c := config.Config{
		Functions: []config.Function{
			config.Function{
				Name:           "test1",
				Message:        "tests.message",
				SuccessMessage: "tests.message.success",
				Templates: config.Templates{
					OutputTemplate: `{ "nicsname": "{{ .JSON.name }}" }`,
				},
			},
		},
	}

	returnPayload = []byte(`{ "name": "nic" }`)

	nw.RegisterMessageListeners(c)
	f := mc.QueueSubscribeCalls()[0].Cb

	msg := stan.Msg{}
	f(&msg) // call the function

	is.Equal(`{ "nicsname": "nic" }`, string(mc.PublishCalls()[0].Data))
}
