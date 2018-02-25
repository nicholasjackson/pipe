package server

import (
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

func setup(t *testing.T) (*is.I, config.Config, *PipeServer) {
	is := is.New(t)

	inputChan = make(chan *providers.Message)

	mockedInputProvider := &providers.ProviderMock{
		ListenFunc: func() (<-chan *providers.Message, error) {
			return inputChan, nil
		},
		PublishFunc: func(in1 []byte) error {
			return nil
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
		PublishFunc: func(in1 []byte) error {
			return nil
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
				OutputProvider: mockedOutputProvider,
			},
		},
	}

	c := config.New()
	c.Outputs["mock_output"] = mockedOutputProvider
	c.Inputs["mock_input"] = mockedInputProvider
	c.Pipes["test_pipe"] = &pipe
	c.ConnectionPools["mock_provider"] = mockedConnectionPool

	s, _ := statsd.New("localhost:8125")
	p := New(c, hclog.Default(), s)

	return is, c, p
}

func TestListenSetsUpInputProviders(t *testing.T) {
	is, c, p := setup(t)

	p.Listen()

	input := c.Inputs["mock_input"].(*providers.ProviderMock)
	is.Equal(1, len(input.SetupCalls()))                                   // should have called setup on the input provider
	is.Equal(c.ConnectionPools["mock_provider"], input.SetupCalls()[0].Cp) // should have setup the inputs
}

func TestListenSetsUpOutputProviders(t *testing.T) {
	is, c, p := setup(t)

	p.Listen()

	output := c.Outputs["mock_output"].(*providers.ProviderMock)
	is.Equal(1, len(output.SetupCalls()))                                   // should have called setup on the inputs provider
	is.Equal(c.ConnectionPools["mock_provider"], output.SetupCalls()[0].Cp) // should have setup the outputs
}

func TestListenListensForInputProviderMessages(t *testing.T) {
	is, c, p := setup(t)

	p.Listen()
	time.Sleep(20 * time.Millisecond) // wait for setup

	input := c.Inputs["mock_input"].(*providers.ProviderMock)
	is.Equal(1, len(input.ListenCalls())) // should be listening for messages
}

func TestListenCallsActionWhenMessageReceived(t *testing.T) {
	is, c, p := setup(t)

	p.Listen()
	time.Sleep(20 * time.Millisecond) // wait for setup

	inputChan <- &providers.Message{Timestamp: time.Now().UnixNano()}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := c.Outputs["mock_output"].(*providers.ProviderMock)
	is.Equal(1, len(output.PublishCalls())) // should send a message to the output
}

func TestListenIgnoresExpiredMessage(t *testing.T) {
	is, c, p := setup(t)
	c.Pipes["test_pipe"].ExpirationDuration = 1 * time.Hour

	p.Listen()
	time.Sleep(20 * time.Millisecond) // wait for setup

	inputChan <- &providers.Message{Timestamp: int64(time.Now().Nanosecond()) - (10 * time.Hour).Nanoseconds()}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := c.Outputs["mock_output"].(*providers.ProviderMock)
	is.Equal(0, len(output.PublishCalls())) // should have ignored the message
}

func TestWorkerCallsFunctionTransformingMessage(t *testing.T) {
	is, c, p := setup(t)
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
