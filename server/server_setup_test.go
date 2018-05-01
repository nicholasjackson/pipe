package server

import (
	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/pipe/config"
	"github.com/nicholasjackson/pipe/logger"
	"github.com/nicholasjackson/pipe/pipe"
	"github.com/nicholasjackson/pipe/providers"
)

type serverTest struct {
	config                    *config.Config
	pipeServer                *PipeServer
	inputChan                 chan providers.Message
	mockedInputProvider       *providers.ProviderMock
	mockedOutputProvider      *providers.ProviderMock
	mockedSuccessFailProvider *providers.ProviderMock
	mockedConnectionPool      *providers.ConnectionPoolMock
	mockedLogger              *logger.LoggerMock
}

func createMocks(m *serverTest) *serverTest {
	m.mockedInputProvider = &providers.ProviderMock{
		ListenFunc: func() (<-chan providers.Message, error) {
			return m.inputChan, nil
		},
		PublishFunc: func(in1 providers.Message) (providers.Message, error) {
			return providers.Message{ID: "abc123"}, nil
		},
		SetupFunc: func() error {
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

	m.mockedOutputProvider = &providers.ProviderMock{
		ListenFunc: func() (<-chan providers.Message, error) {
			panic("TODO: mock out the Listen method")
		},
		PublishFunc: func(in1 providers.Message) (providers.Message, error) {
			return providers.Message{ID: "abc123"}, nil
		},
		SetupFunc: func() error {
			return nil
		},
		StopFunc: func() error {
			panic("TODO: mock out the Stop method")
		},
		TypeFunc: func() string {
			return "mock_provider"
		},
	}

	m.mockedSuccessFailProvider = &providers.ProviderMock{
		ListenFunc: func() (<-chan providers.Message, error) {
			panic("TODO: mock out the Listen method")
		},
		PublishFunc: func(in1 providers.Message) (providers.Message, error) {
			return providers.Message{ID: "abc123"}, nil
		},
		SetupFunc: func() error {
			return nil
		},
		StopFunc: func() error {
			panic("TODO: mock out the Stop method")
		},
		TypeFunc: func() string {
			return "mock_provider"
		},
	}

	m.mockedConnectionPool = &providers.ConnectionPoolMock{}

	// make and configure a mocked Logger
	m.mockedLogger = &logger.LoggerMock{
		GetLoggerFunc: func() hclog.Logger {
			return nil
		},
		GetStatsDFunc: func() *statsd.Client {
			return nil
		},
		ProviderConnectionCreatedFunc: func(in1 providers.Provider) {
		},
		ProviderConnectionFailedFunc: func(in1 providers.Provider, in2 error) {
		},
		ProviderMessagePublishedFunc: func(in1 providers.Provider, in2 providers.Message, in3 ...interface{}) {
		},
		ProviderSubcriptionCreatedFunc: func(in1 providers.Provider) {
		},
		ProviderSubcriptionFailedFunc: func(in1 providers.Provider, in2 error) {
		},
		ServerActionPublishFunc: func(in1 *pipe.Pipe, in2 providers.Message) {
		},
		ServerActionPublishFailedFunc: func(in1 *pipe.Pipe, in2 providers.Message, in3 error) {
		},
		ServerActionPublishSuccessFunc: func(in1 *pipe.Pipe, in2 providers.Message) {
		},
		ServerFailPublishFunc: func(in1 *pipe.Pipe, in2 *pipe.Action, in3 providers.Message) {
		},
		ServerFailPublishFailedFunc: func(in1 *pipe.Pipe, in2 *pipe.Action, in3 providers.Message, in4 error) {
		},
		ServerFailPublishSuccessFunc: func(in1 *pipe.Pipe, in2 *pipe.Action, in3 providers.Message) {
		},
		ServerHandleMessageExpiredFunc: func(in1 *pipe.Pipe, in2 providers.Message) {
		},
		ServerNewMessageReceivedStartFunc: func(in1 *pipe.Pipe, in2 providers.Message) *logger.LoggerTiming {
			return &logger.LoggerTiming{
				Stop: func() {},
			}
		},
		ServerNoPipesConfiguredFunc: func(in1 providers.Provider) {
		},
		ServerSuccessPublishFunc: func(in1 *pipe.Pipe, in2 *pipe.Action, in3 providers.Message) {
		},
		ServerSuccessPublishFailedFunc: func(in1 *pipe.Pipe, in2 *pipe.Action, in3 providers.Message, in4 error) {
		},
		ServerSuccessPublishSuccessFunc: func(in1 *pipe.Pipe, in2 *pipe.Action, in3 providers.Message) {
		},
		ServerTemplateProcessFailFunc: func(in1 *pipe.Action, in2 []byte, in3 error) {
		},
		ServerTemplateProcessStartFunc: func(in1 *pipe.Action, in2 []byte) *logger.LoggerTiming {
			return &logger.LoggerTiming{
				Stop: func() {},
			}
		},
		ServerTemplateProcessSuccessFunc: func(in1 *pipe.Action, in2 []byte) {
		},
		ServerUnableToListenFunc: func(in1 providers.Provider, in2 error) {
		},
	}

	return m
}
