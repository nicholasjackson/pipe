package logger

import (
	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/pipe/pipe"
	"github.com/nicholasjackson/pipe/providers"
)

// ServerLogger defines the interface for logging methods for the server
type ServerLogger interface {
	ServerUnableToListen(providers.Provider, error)
	ServerNoPipesConfigured(providers.Provider)
	ServerNewMessageReceivedStart(*pipe.Pipe, *providers.Message) *LoggerTiming
	ServerHandleMessageExpired(*pipe.Pipe, *providers.Message)
	ServerActionPublish(*pipe.Pipe, *providers.Message)
	ServerActionPublishFailed(*pipe.Pipe, *providers.Message, error)
	ServerActionPublishSuccess(*pipe.Pipe, *providers.Message)
	ServerSuccessPublish(*pipe.Pipe, *pipe.Action, *providers.Message)
	ServerSuccessPublishFailed(*pipe.Pipe, *pipe.Action, *providers.Message, error)
	ServerSuccessPublishSuccess(*pipe.Pipe, *pipe.Action, *providers.Message)
	ServerFailPublish(*pipe.Pipe, *pipe.Action, *providers.Message)
	ServerFailPublishFailed(*pipe.Pipe, *pipe.Action, *providers.Message, error)
	ServerFailPublishSuccess(*pipe.Pipe, *pipe.Action, *providers.Message)

	ServerTemplateProcessStart(*pipe.Action, []byte) *LoggerTiming
	ServerTemplateProcessFail(*pipe.Action, []byte, error)
	ServerTemplateProcessSuccess(*pipe.Action, []byte)
}

// ProviderLogger defines logging methods for a provider
type ProviderLogger interface {
	ProviderConnectionFailed(providers.Provider, error)
	ProviderConnectionCreated(providers.Provider)
	ProviderSubcriptionFailed(providers.Provider, error)
	ProviderSubcriptionCreated(providers.Provider)
	ProviderMessagePublished(providers.Provider, *providers.Message, ...interface{})
}

//go:generate moq -out mock_logger.go . Logger
// Logger proposes a standard interface for provider logging and metrics
type Logger interface {
	GetLogger() hclog.Logger
	GetStatsD() *statsd.Client

	ServerLogger
	ProviderLogger
}

type LoggerTiming struct {
	Stop func()
}

// LoggerImpl is a concrete implementation of the Logger interface
type LoggerImpl struct {
	logger hclog.Logger
	stats  *statsd.Client
}

// New creates a new logger from the given logger and statsd client
func New(l hclog.Logger, s *statsd.Client) Logger {
	return &LoggerImpl{
		logger: l,
		stats:  s,
	}
}

// GetLogger returns the assigned logger
func (l *LoggerImpl) GetLogger() hclog.Logger {
	return l.logger
}

// GetStatsD returns the assigned statsd client
func (l *LoggerImpl) GetStatsD() *statsd.Client {
	return l.stats
}
