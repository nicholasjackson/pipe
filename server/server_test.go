package server

import (
	"fmt"
	"testing"

	"time"

	"github.com/matryer/is"
	"github.com/nicholasjackson/pipe/config"
	"github.com/nicholasjackson/pipe/pipe"
	"github.com/nicholasjackson/pipe/providers"
)

func setup(t *testing.T) (*is.I, *serverTest) {
	is := is.New(t)

	testElements := &serverTest{
		inputChan: make(chan *providers.Message),
	}
	m := createMocks(testElements)

	pipe := pipe.Pipe{
		Name:               "test_pipe",
		Input:              "mock_input",
		InputProvider:      m.mockedInputProvider,
		Expiration:         "5s",
		ExpirationDuration: 5 * time.Second,

		Action: pipe.Action{
			Output:         "mock_output",
			OutputProvider: m.mockedOutputProvider,
		},

		OnSuccess: []pipe.Action{
			pipe.Action{
				Output:         "mock_success",
				OutputProvider: m.mockedSuccessFailProvider,
			},
		},
		OnFail: []pipe.Action{
			pipe.Action{
				Output:         "mock_fail",
				OutputProvider: m.mockedSuccessFailProvider,
			},
		},
	}

	c := config.New(m.mockedLogger)
	c.Outputs["mock_output"] = m.mockedOutputProvider
	c.Outputs["mock_success_fail"] = m.mockedSuccessFailProvider
	c.Inputs["mock_input"] = m.mockedInputProvider
	c.Pipes["test_pipe"] = &pipe
	c.ConnectionPools["mock_provider"] = m.mockedConnectionPool

	p := New(c, m.mockedLogger)
	testElements.config = c
	testElements.pipeServer = p

	return is, testElements
}

func TestListenSetsUpInputProviders(t *testing.T) {
	is, te := setup(t)

	te.pipeServer.Listen()

	input := te.config.Inputs["mock_input"].(*providers.ProviderMock)
	is.Equal(1, len(input.SetupCalls())) // should have called setup on the input provider
}

func TestListenSetsUpOutputProviders(t *testing.T) {
	is, te := setup(t)

	te.pipeServer.Listen()

	output := te.config.Outputs["mock_output"].(*providers.ProviderMock)
	is.Equal(1, len(output.SetupCalls())) // should have called setup on the inputs provider
}

func TestListenListensForInputProviderMessages(t *testing.T) {
	is, te := setup(t)

	te.pipeServer.Listen()

	input := te.config.Inputs["mock_input"].(*providers.ProviderMock)
	is.Equal(1, len(input.ListenCalls())) // should be listening for messages
}

func TestListenCallsActionWhenMessageReceived(t *testing.T) {
	is, te := setup(t)

	te.pipeServer.Listen()

	te.inputChan <- &providers.Message{Timestamp: time.Now().UnixNano()}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := te.config.Outputs["mock_output"].(*providers.ProviderMock)
	is.Equal(1, len(output.PublishCalls())) // should send a message to the output
}

func TestListenSetsParentIDForActionMessages(t *testing.T) {
	is, te := setup(t)

	te.pipeServer.Listen()

	m := providers.NewMessage()
	m.Timestamp = time.Now().UnixNano()
	m.Data = []byte(`{ "name": "nic" }`)

	te.inputChan <- &m
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := te.config.Outputs["mock_output"].(*providers.ProviderMock)
	is.Equal(m.ID, output.PublishCalls()[0].In1.ParentID) // should set the parent id when sending success
}

func TestListenIgnoresExpiredMessage(t *testing.T) {
	is, te := setup(t)
	te.config.Pipes["test_pipe"].ExpirationDuration = 1 * time.Hour

	te.pipeServer.Listen()

	te.inputChan <- &providers.Message{Timestamp: int64(time.Now().Nanosecond()) - (10 * time.Hour).Nanoseconds()}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := te.config.Outputs["mock_output"].(*providers.ProviderMock)
	is.Equal(0, len(output.PublishCalls())) // should have ignored the message
}

func TestListenCallsActionTransformingMessage(t *testing.T) {
	is, te := setup(t)
	te.config.Pipes["test_pipe"].Action.Template = `{ "nicsname": "{{ .JSON.name }}" }`

	te.pipeServer.Listen()

	te.inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := te.config.Outputs["mock_output"].(*providers.ProviderMock)
	is.Equal(1, len(output.PublishCalls()))                                      // expected 1 call to function
	is.Equal(`{ "nicsname": "nic" }`, string(output.PublishCalls()[0].In1.Data)) // expected processed payload to be passed
}

func TestListenPublishesSuccessEventPostAction(t *testing.T) {
	is, te := setup(t)

	te.pipeServer.Listen()

	te.inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := te.config.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(1, len(output.PublishCalls()))                                  // expected 2 call to function
	is.Equal(`{ "name": "nic" }`, string(output.PublishCalls()[0].In1.Data)) // expected processed payload to be passed
}

func TestListenPublishesMultipleSuccessEventsPostAction(t *testing.T) {
	is, te := setup(t)
	te.config.Pipes["test_pipe"].OnSuccess = append(te.config.Pipes["test_pipe"].OnSuccess, te.config.Pipes["test_pipe"].OnSuccess[0])

	te.pipeServer.Listen()

	te.inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := te.config.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(2, len(output.PublishCalls())) // expected 1 call to function
}

func TestListenSetsParentIDForSuccessMessages(t *testing.T) {
	is, te := setup(t)
	te.config.Pipes["test_pipe"].OnSuccess = append(te.config.Pipes["test_pipe"].OnSuccess, te.config.Pipes["test_pipe"].OnSuccess[0])

	te.pipeServer.Listen()

	m := providers.NewMessage()
	m.Timestamp = time.Now().UnixNano()
	m.Data = []byte(`{ "name": "nic" }`)

	te.inputChan <- &m
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := te.config.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(m.ID, output.PublishCalls()[0].In1.ParentID) // should set the parent id when sending success
}

func TestListenTransformsSuccessEventPostAction(t *testing.T) {
	is, te := setup(t)
	te.config.Pipes["test_pipe"].OnSuccess[0].Template = `{ "nicsname": "{{ .JSON.name }}" }`

	te.pipeServer.Listen()

	te.inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := te.config.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(1, len(output.PublishCalls()))                                      // expected 1 call to function
	is.Equal(`{ "nicsname": "nic" }`, string(output.PublishCalls()[0].In1.Data)) // expected processed payload to be passed
}

func TestListenPublishesFailEventPostAction(t *testing.T) {
	is, te := setup(t)
	te.mockedOutputProvider.PublishFunc = func(in1 providers.Message) (providers.Message, error) {
		return providers.Message{}, fmt.Errorf("boom")
	}

	te.pipeServer.Listen()

	te.inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := te.config.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(1, len(output.PublishCalls()))                                  // expected 1 call to function
	is.Equal(`{ "name": "nic" }`, string(output.PublishCalls()[0].In1.Data)) // expected processed payload to be passed
}

func TestListenPublishesMultipleFailEventsPostAction(t *testing.T) {
	is, te := setup(t)
	te.mockedOutputProvider.PublishFunc = func(in1 providers.Message) (providers.Message, error) {
		return providers.Message{}, fmt.Errorf("boom")
	}
	te.config.Pipes["test_pipe"].OnFail = append(te.config.Pipes["test_pipe"].OnFail, te.config.Pipes["test_pipe"].OnFail[0])

	te.pipeServer.Listen()

	te.inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := te.config.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(2, len(output.PublishCalls())) // expected 1 call to function
}

func TestListenTransformsFailEventPostAction(t *testing.T) {
	is, te := setup(t)
	te.mockedOutputProvider.PublishFunc = func(in1 providers.Message) (providers.Message, error) {
		return providers.Message{}, fmt.Errorf("boom")
	}
	te.config.Pipes["test_pipe"].OnFail[0].Template = `{ "nicsname": "{{ .JSON.name }}" }`

	te.pipeServer.Listen()

	te.inputChan <- &providers.Message{
		Timestamp: time.Now().UnixNano(),
		Data:      []byte(`{ "name": "nic" }`),
	}
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := te.config.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(1, len(output.PublishCalls()))                                      // expected 1 call to function
	is.Equal(`{ "nicsname": "nic" }`, string(output.PublishCalls()[0].In1.Data)) // expected processed payload to be passed
}

func TestListenSetsParentIDForFailMessages(t *testing.T) {
	is, te := setup(t)
	te.mockedOutputProvider.PublishFunc = func(in1 providers.Message) (providers.Message, error) {
		return providers.Message{}, fmt.Errorf("boom")
	}
	te.config.Pipes["test_pipe"].OnFail = append(te.config.Pipes["test_pipe"].OnFail, te.config.Pipes["test_pipe"].OnFail[0])

	te.pipeServer.Listen()

	m := providers.NewMessage()
	m.Timestamp = time.Now().UnixNano()
	m.Data = []byte(`{ "name": "nic" }`)

	te.inputChan <- &m
	time.Sleep(20 * time.Millisecond) // wait for message to be recieved

	output := te.config.Outputs["mock_success_fail"].(*providers.ProviderMock)
	is.Equal(m.ID, output.PublishCalls()[0].In1.ParentID) // should set the parent id when sending success
}
