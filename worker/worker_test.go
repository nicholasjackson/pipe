package worker

import (
	"testing"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
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
		CallFunctionFunc: func(name string, query string, payload []byte) ([]byte, error) {
			return returnPayload, returnError
		},
	}

	logger := hclog.New(&hclog.LoggerOptions{Level: hclog.LevelFromString("DEBUG")})
	stats, _ := statsd.New("")

	return is.New(t), NewNatsWorker(mockedNatsConnection, mockedClient, stats, logger), mockedNatsConnection, mockedClient
}

func createMessage(data []byte) *stan.Msg {
	msg := &stan.Msg{}
	msg.Data = data
	msg.Timestamp = time.Now().UnixNano()

	return msg
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

func TestDoesNotRegistersMessageListenerWhenInvalidExpiration(t *testing.T) {
	is, nw, mc, _ := setupWorkerTests(t)

	c := config.Config{
		Functions: []config.Function{
			config.Function{
				Name:       "test1",
				Message:    "tests.message",
				Expiration: "pie",
			},
		},
	}

	nw.RegisterMessageListeners(c)

	is.Equal(0, len(mc.QueueSubscribeCalls())) // expected 1 call to subscribe
}

func TestWorkerCallsFunctionWithRawMessage(t *testing.T) {
	is, nw, mc, cl := setupWorkerTests(t)

	c := config.Config{
		Gateway: "http://myserver.com",
		Functions: []config.Function{
			config.Function{
				Name:    "test1",
				Query:   "test=yes",
				Message: "tests.message",
			},
		},
	}

	nw.RegisterMessageListeners(c)
	f := mc.QueueSubscribeCalls()[0].Cb

	msg := createMessage([]byte("data"))
	f(msg) // call the function

	is.Equal(1, len(cl.CallFunctionCalls()))                      // expected 1 call to function
	is.Equal("data", string(cl.CallFunctionCalls()[0].Payload))   // expected raw payload to be passed
	is.Equal("test=yes", string(cl.CallFunctionCalls()[0].Query)) // expected query to be set
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

	msg := createMessage([]byte(`{ "name": "nic" }`))
	f(msg) // call the function

	is.Equal(1, len(cl.CallFunctionCalls()))                                     // expected 1 call to function
	is.Equal(`{ "nicsname": "nic" }`, string(cl.CallFunctionCalls()[0].Payload)) // expected processed payload to be passed
}

func TestWorkerIgnoresExpiredMessage(t *testing.T) {
	is, nw, mc, cl := setupWorkerTests(t)

	c := config.Config{
		Gateway: "http://myserver.com",
		Functions: []config.Function{
			config.Function{
				Name:       "test1",
				Query:      "test=yes",
				Message:    "tests.message",
				Expiration: "1us",
			},
		},
	}

	nw.RegisterMessageListeners(c)
	f := mc.QueueSubscribeCalls()[0].Cb

	msg := createMessage([]byte("data"))
	f(msg) // call the function

	is.Equal(0, len(cl.CallFunctionCalls())) // expected 0 calls to function
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

	msg := createMessage(nil)
	f(msg) // call the function

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

	msg := createMessage(nil)
	f(msg) // call the function

	is.Equal(`{ "nicsname": "nic" }`, string(mc.PublishCalls()[0].Data)) // expected template to have been transformed
}
