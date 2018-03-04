package server

import (
	"fmt"
	"testing"

	"time"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/matryer/is"
	"github.com/nicholasjackson/pipe/config"
	"github.com/nicholasjackson/pipe/pipe"
	"github.com/nicholasjackson/pipe/providers"
)

var inputChan chan *providers.Message

func setup(t *testing.T, actionError error) (*is.I, config.Config, *PipeServer) {
	is := is.New(t)

	inputChan = make(chan *providers.Message)

	mockedInputProvider := &providers.ProviderMock{
		ListenFunc: func() (<-chan *providers.Message, error) {
			return inputChan, nil
		},
		PublishFunc: func(in1 []byte) ([]byte, error) {
			return nil, nil
		},
		SetupFunc: func(cp providers.ConnectionPool, log hclog.Logger, stats *statsd.Client) error {
			return nil
		},
		StopFunc: func() error {
			panic("TODO: mock out the Stop method")
		},
		TypeFunc: func() string {
			return "mock_provider"
		},
		NameFunc: func() string {
			return "mock_input"
		},
	}

	mockedOutputProvider := &providers.ProviderMock{
		ListenFunc: func() (<-chan *providers.Message, error) {
			panic("TODO: mock out the Listen method")
		},
		PublishFunc: func(in1 []byte) ([]byte, error) {
			return nil, actionError
		},
		SetupFunc: func(cp providers.ConnectionPool, log hclog.Logger, stats *statsd.Client) error {
			return nil
		},
		StopFunc: func() error {
			panic("TODO: mock out the Stop method")
		},
		TypeFunc: func() string {
			return "mock_provider"
		},
	}

	mockedSuccessFailProvider := &providers.ProviderMock{
		ListenFunc: func() (<-chan *providers.Message, error) {
			panic("TODO: mock out the Listen method")
		},
		PublishFunc: func(in1 []byte) ([]byte, error) {
			return nil, nil
		},
		SetupFunc: func(cp providers.ConnectionPool, log hclog.Logger, stats *statsd.Client) error {
			return nil
		},
		StopFunc: func() error {
			panic("TODO: mock out the Stop method")
		},
		TypeFunc: func() string {
			return "mock_provider"
		},
	}

	mockedConnectionPool := &providers.ConnectionPoolMock{}

	pipe := pipe.Pipe{
		Name:               "test_pipe",
		Input:              "mock_input",
		InputProvider:      mockedInputProvider,
		Expiration:         "5s",
		ExpirationDuration: 5 * time.Second,

		Action: pipe.Action{
			Output:         "mock_output",
			OutputProvider: mockedOutputProvider,
		},

		OnSuccess: []pipe.Action{
			pipe.Action{
				Output:         "mock_success",
				OutputProvider: mockedSuccessFailProvider,
			},
		},
		OnFail: []pipe.Action{
			pipe.Action{
				Output:         "mock_fail",
				OutputProvider: mockedSuccessFailProvider,
			},
		},
	}

	c := config.New()
	c.Outputs["mock_output"] = mockedOutputProvider
	c.Outputs["mock_success_fail"] = mockedSuccessFailProvider
	c.Inputs["mock_input"] = mockedInputProvider
	c.Pipes["test_pipe"] = &pipe
	c.ConnectionPools["mock_provider"] = mockedConnectionPool

	s, _ := statsd.New("localhost:8125")
	p := New(c, hclog.Default(), s)

	return is, c, p
}

func TestListenSetsUpInputProviders(t *testing.T) {
	is, c, p := setup(t, nil)

	p.Listen()

	input := c.Inputs["mock_input"].(*providers.ProviderMock)
	is.Equal(1, len(input.SetupCalls()))                                   // should have called setup on the input provider
	is.Equal(c.ConnectionPools["mock_provider"], input.SetupCalls()[0].Cp) // should have setup the inputs
}

func TestListenSetsUpOutputProviders(t *testing.T) {
	is, c, p := setup(t, nil)

	p.Listen()

	output := c.Outputs["mock_output"].(*providers.ProviderMock)
	is.Equal(1, len(output.SetupCalls()))                                   // should have called setup on the inputs provider
	is.Equal(c.ConnectionPools["mock_provider"], output.SetupCalls()[0].Cp) // should have setup the outputs
}

func TestListenListensForInputProviderMessages(t *testing.T) {
	is, c, p := setup(t, nil)

	p.Listen()
	time.Sleep(20 * time.Millisecond) // wait for setup

	input := c.Inputs["mock_input"].(*providers.ProviderMock)
	is.Equal(1, len(input.ListenCalls())) // should be listening for messages
}

func TestListenCallsActionWhenMessageReceived(t *testing.T) {
	is, c, p := setup(t, nil)

	p.Listen()
	time.Sleep(20 * time.Millisecond) // wait for setup

	inputChan <- &providers.Message{Timestamp: time.Now().UnixNano()}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := c.Outputs["mock_output"].(*providers.ProviderMock)
	is.Equal(1, len(output.PublishCalls())) // should send a message to the output
}

func TestListenIgnoresExpiredMessage(t *testing.T) {
	is, c, p := setup(t, nil)
	c.Pipes["test_pipe"].ExpirationDuration = 1 * time.Hour

	p.Listen()
	time.Sleep(20 * time.Millisecond) // wait for setup

	inputChan <- &providers.Message{Timestamp: int64(time.Now().Nanosecond()) - (10 * time.Hour).Nanoseconds()}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := c.Outputs["mock_output"].(*providers.ProviderMock)
	is.Equal(0, len(output.PublishCalls())) // should have ignored the message
}

func TestListenCallsActionTransformingMessage(t *testing.T) {
	is, c, p := setup(t, nil)
	c.Pipes["test_pipe"].Action.Template = `{ "nicsname": "{{ .JSON.name }}" }`

	p.Listen()
	time.Sleep(20 * time.Millisecond) // wait for setup

	inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := c.Outputs["mock_output"].(*providers.ProviderMock)
	is.Equal(1, len(output.PublishCalls()))                                 // expected 1 call to function
	is.Equal(`{ "nicsname": "nic" }`, string(output.PublishCalls()[0].In1)) // expected processed payload to be passed
}

func TestListenPublishesSuccessEventPostAction(t *testing.T) {
	is, c, p := setup(t, nil)

	p.Listen()
	time.Sleep(20 * time.Millisecond) // wait for setup

	inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := c.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(1, len(output.PublishCalls()))                             // expected 2 call to function
	is.Equal(`{ "name": "nic" }`, string(output.PublishCalls()[0].In1)) // expected processed payload to be passed
}

func TestListenPublishesMultipleSuccessEventsPostAction(t *testing.T) {
	is, c, p := setup(t, nil)
	c.Pipes["test_pipe"].OnSuccess = append(c.Pipes["test_pipe"].OnSuccess, c.Pipes["test_pipe"].OnSuccess[0])

	p.Listen()
	time.Sleep(20 * time.Millisecond) // wait for setup

	inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := c.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(2, len(output.PublishCalls())) // expected 1 call to function
}

func TestListenTransformsSuccessEventPostAction(t *testing.T) {
	is, c, p := setup(t, nil)
	c.Pipes["test_pipe"].OnSuccess[0].Template = `{ "nicsname": "{{ .JSON.name }}" }`

	p.Listen()
	time.Sleep(20 * time.Millisecond) // wait for setup

	inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := c.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(1, len(output.PublishCalls()))                                 // expected 1 call to function
	is.Equal(`{ "nicsname": "nic" }`, string(output.PublishCalls()[0].In1)) // expected processed payload to be passed
}

func TestListenPublishesFailEventPostAction(t *testing.T) {
	is, c, p := setup(t, fmt.Errorf("boom"))

	p.Listen()
	time.Sleep(20 * time.Millisecond) // wait for setup

	inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := c.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(1, len(output.PublishCalls()))                             // expected 1 call to function
	is.Equal(`{ "name": "nic" }`, string(output.PublishCalls()[0].In1)) // expected processed payload to be passed
}

func TestListenPublishesMultipleFailEventsPostAction(t *testing.T) {
	is, c, p := setup(t, fmt.Errorf("boom"))
	c.Pipes["test_pipe"].OnFail = append(c.Pipes["test_pipe"].OnFail, c.Pipes["test_pipe"].OnFail[0])

	p.Listen()
	time.Sleep(20 * time.Millisecond) // wait for setup

	inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := c.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(2, len(output.PublishCalls())) // expected 1 call to function
}

func TestListenTransformsFailEventPostAction(t *testing.T) {
	is, c, p := setup(t, fmt.Errorf("boom"))
	c.Pipes["test_pipe"].OnFail[0].Template = `{ "nicsname": "{{ .JSON.name }}" }`

	p.Listen()
	time.Sleep(20 * time.Millisecond) // wait for setup

	inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := c.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(1, len(output.PublishCalls()))                                 // expected 1 call to function
	is.Equal(`{ "nicsname": "nic" }`, string(output.PublishCalls()[0].In1)) // expected processed payload to be passed
}
