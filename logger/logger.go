package logger

import (
	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/pipe/pipe"
	"github.com/nicholasjackson/pipe/providers"
)

// Logger proposes a standard interface for provider logging and metrics
type Logger interface {
	ServerNoPipesConfigured(providers.Provider)
	ServerNewMessageReceived(*pipe.Pipe, providers.Message)
	ServerHandleMessageDuration(*pipe.Pipe)
	ServerHandleMessageExpired(*pipe.Pipe, providers.Message)
	ServerActionPublish(*pipe.Pipe, providers.Message)
	ServerActionPublishFailed(*pipe.Pipe, providers.Message)
	ServerActionPublishSuccess(*pipe.Pipe, providers.Message)
	ServerSuccessPublish(*pipe.Pipe, providers.Message)
	ServerSuccessPublishFailed(*pipe.Pipe, providers.Message)
	ServerSuccessPublishSuccess(*pipe.Pipe, providers.Message)

	ServerTemplateProcess(pipe.Action, []byte)
	ServerTemplateProcessFail(pipe.Action, []byte)
	ServerTemplateProcessSuccess(pipe.Action, []byte)

	ProviderConnectionFailed(providers.Provider)
	ProviderConnectionCreated(providers.Provider)
	ProviderSubcriptionFailed(providers.Provider)
	ProviderSubcriptionCreated(providers.Provider)
	ProviderMessagePublished(providers.Provider, providers.Message)
}

// LoggerImpl is a concrete implementation of the Logger interface
type LoggerImpl struct {
	logger hclog.Logger
	stats  *statsd.Client
}

func NewLogger(l hclog.Logger, s *statsd.Client) Logger {
	return &LoggerImpl{}
}

func (l *LoggerImpl) ServerNoPipesConfigured(providers.Provider)                {}
func (l *LoggerImpl) ServerNewMessageReceived(*pipe.Pipe, providers.Message)    {}
func (l *LoggerImpl) ServerHandleMessageDuration(*pipe.Pipe)                    {}
func (l *LoggerImpl) ServerHandleMessageExpired(*pipe.Pipe, providers.Message)  {}
func (l *LoggerImpl) ServerActionPublish(*pipe.Pipe, providers.Message)         {}
func (l *LoggerImpl) ServerActionPublishFailed(*pipe.Pipe, providers.Message)   {}
func (l *LoggerImpl) ServerActionPublishSuccess(*pipe.Pipe, providers.Message)  {}
func (l *LoggerImpl) ServerSuccessPublish(*pipe.Pipe, providers.Message)        {}
func (l *LoggerImpl) ServerSuccessPublishFailed(*pipe.Pipe, providers.Message)  {}
func (l *LoggerImpl) ServerSuccessPublishSuccess(*pipe.Pipe, providers.Message) {}

func (l *LoggerImpl) ServerTemplateProcess(pipe.Action, []byte)        {}
func (l *LoggerImpl) ServerTemplateProcessFail(pipe.Action, []byte)    {}
func (l *LoggerImpl) ServerTemplateProcessSuccess(pipe.Action, []byte) {}

func (l *LoggerImpl) ProviderConnectionFailed(providers.Provider)                    {}
func (l *LoggerImpl) ProviderConnectionCreated(providers.Provider)                   {}
func (l *LoggerImpl) ProviderSubcriptionFailed(providers.Provider)                   {}
func (l *LoggerImpl) ProviderSubcriptionCreated(providers.Provider)                  {}
func (l *LoggerImpl) ProviderMessagePublished(providers.Provider, providers.Message) {}
