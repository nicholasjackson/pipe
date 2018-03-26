package logger

import (
	"time"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/pipe/pipe"
	"github.com/nicholasjackson/pipe/providers"
)

//go:generate moq -out mock_logger.go . Logger
// Logger proposes a standard interface for provider logging and metrics
type Logger interface {
	GetLogger() hclog.Logger
	GetStatsD() *statsd.Client

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

	ProviderConnectionFailed(*providers.Provider)
	ProviderConnectionCreated(*providers.Provider)
	ProviderSubcriptionFailed(*providers.Provider)
	ProviderSubcriptionCreated(*providers.Provider)
	ProviderMessagePublished(*providers.Provider, *providers.Message)
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
	return &LoggerImpl{}
}

// GetLogger returns the assigned logger
func (l *LoggerImpl) GetLogger() hclog.Logger {
	return l.logger
}

// GetStatsD returns the assigned statsd client
func (l *LoggerImpl) GetStatsD() *statsd.Client {
	return l.stats
}

// ServerUnableToListen writes an error message to the log and statsd
func (l *LoggerImpl) ServerUnableToListen(p providers.Provider, err error) {
	l.logger.Error("Unable to listen for input", "error", err, "provider", p.Name(), "type", p.Type())
	l.stats.Incr("listen.error", []string{"provider:" + p.Name()}, 1)
}

// ServerNoPipesConfigured writes an error message to the log and statsd
func (l *LoggerImpl) ServerNoPipesConfigured(p providers.Provider) {
	l.logger.Info("No pipes configured to handle this message", "provider", p.Name(), "type", p.Type())
	l.stats.Incr("listen.nopipes.configured", []string{"provider:" + p.Name()}, 1)
}

// ServerNewMessageReceivedStart logs that a new message has been received, user should call the returned function
// Stop to ensure timing data is submitted to the logs i.e: defer p.logger.ServerStartNewMessageReceived(p, m).Stop()
func (l *LoggerImpl) ServerNewMessageReceivedStart(pi *pipe.Pipe, m *providers.Message) *LoggerTiming {
	l.logger.Info("Recieved message", "pipe", pi.Name)
	l.logger.Debug("Message data", "message", m)

	// time the length of the message handling
	st := time.Now()
	return &LoggerTiming{
		Stop: func() {
			dur := time.Now().Sub(st)
			l.stats.Timing("handler.message.called", dur, []string{"pipe:" + pi.Name}, 1)
			l.logger.Info("Finished processing message", "pipe", pi.Name, "time", dur)
		},
	}
}

// ServerHandleMessageExpired logs that a message has expired and will not be handled
func (l *LoggerImpl) ServerHandleMessageExpired(pi *pipe.Pipe, m *providers.Message) {
	l.logger.Info("Message expired", "pipe", pi.Name, "timestamp", m.Timestamp, "expiration", pi.ExpirationDuration)
	l.stats.Incr("handler.message.expired", []string{"pipe:" + pi.Name}, 1)
}

// ServerActionPublish logs that a message has been published for a defined action
func (l *LoggerImpl) ServerActionPublish(pi *pipe.Pipe, m *providers.Message) {
	l.logger.Info("Publish message action", "pipe", pi.Name, "output", pi.Action.Output)
	l.stats.Incr("handler.message.action.publish", []string{"pipe:" + pi.Name}, 1)
}

// ServerActionPublishFailed logs that publishing a message to an action has failed
func (l *LoggerImpl) ServerActionPublishFailed(pi *pipe.Pipe, m *providers.Message, err error) {
	l.logger.Error("Publish message action failed", "pipe", pi.Name, "error", err, "message", m)
	l.stats.Incr("handler.message.action.publish.failed", []string{"pipe:" + pi.Name}, 1)
}

// ServerActionPublishSuccess logs that publishing a message was succcessful
func (l *LoggerImpl) ServerActionPublishSuccess(pi *pipe.Pipe, m *providers.Message) {
	l.logger.Info("Publish message action succeded", "pipe", pi.Name, "output", pi.Action.Output)
	l.logger.Debug("Message data", "message", m)

	l.stats.Incr("handler.message.action.publish.success", []string{"pipe:" + pi.Name}, 1)
}

// ServerSuccessPublish logs that a success message will be published
func (l *LoggerImpl) ServerSuccessPublish(pi *pipe.Pipe, a *pipe.Action, m *providers.Message) {
	l.logger.Info("Attempt process success action", "pipe", pi.Name, "output", a.Output)
	l.stats.Incr("handler.message.success.publish.called", []string{"pipe:" + pi.Name, "output:" + a.Output}, 1)
}

// ServerSuccessPublishFailed logs that the publishing of a success message has failed
func (l *LoggerImpl) ServerSuccessPublishFailed(pi *pipe.Pipe, a *pipe.Action, m *providers.Message, err error) {
	l.logger.Error("Publish success action failed", "pipe", pi.Name, "output", a.Output, "error", err, "message", m)
	l.stats.Incr("handler.message.success.publish.failed", []string{"pipe:" + pi.Name, "output:" + a.Output}, 1)
}

// ServerSuccessPublishSuccess logs that the publishing of the success message has returned without error
func (l *LoggerImpl) ServerSuccessPublishSuccess(pi *pipe.Pipe, a *pipe.Action, m *providers.Message) {
	l.logger.Info("Publish success action succeded", "pipe", pi.Name, "output", pi.Action.Output)
	l.stats.Incr("handler.message.success.publish.success", []string{"pipe:" + pi.Name, "output:" + a.Output}, 1)
}

// ServerFailPublish logs that a success message will be published
func (l *LoggerImpl) ServerFailPublish(pi *pipe.Pipe, a *pipe.Action, m *providers.Message) {
	l.logger.Info("Attempt process fail action", "pipe", pi.Name, "output", a.Output)
	l.stats.Incr("handler.message.fail.publish.called", []string{"pipe:" + pi.Name, "output:" + a.Output}, 1)
}

// ServerFailPublishFailed logs that the publishing of a success message has failed
func (l *LoggerImpl) ServerFailPublishFailed(pi *pipe.Pipe, a *pipe.Action, m *providers.Message, err error) {
	l.logger.Error("Publish fail action failed", "pipe", pi.Name, "output", a.Output, "error", err, "message", m)
	l.stats.Incr("handler.message.fail.publish.failed", []string{"pipe:" + pi.Name, "output:" + a.Output}, 1)
}

// ServerFailPublishSuccess logs that the publishing of the success message has returned without error
func (l *LoggerImpl) ServerFailPublishSuccess(pi *pipe.Pipe, a *pipe.Action, m *providers.Message) {
	l.logger.Info("Publish fail action succeded", "pipe", pi.Name, "output", pi.Action.Output)
	l.stats.Incr("handler.message.fail.publish.success", []string{"pipe:" + pi.Name, "output:" + a.Output}, 1)
}

// ServerTemplateProcess logs that the system will start processing a template, since this method times the execution
// of the process the user must called the returned Stop function e.g: defer p.logger.ServerTemplateProcessStart.Stop()
func (l *LoggerImpl) ServerTemplateProcessStart(a *pipe.Action, data []byte) *LoggerTiming {
	l.logger.Info("Transform output template", "output", a.Output, "template", a.Template)
	l.logger.Debug("Transform output template", "output", a.Output, "template", a.Template, "data", data)

	// time the length of the message handling
	st := time.Now()
	return &LoggerTiming{
		Stop: func() {
			dur := time.Now().Sub(st)
			l.stats.Timing("template.process.called", dur, []string{"output:" + a.Output}, 1)
		},
	}
}

// ServerTemplateProcessFail logs that a template has failed to process
func (l *LoggerImpl) ServerTemplateProcessFail(a *pipe.Action, data []byte, err error) {
	l.logger.Error("Error processing output template", "output", a.Output, "error", err)
	l.stats.Incr("handler.message.template.failed", []string{"output:" + a.Output}, 1)
}

// ServerTemplateProcessSuccess logs that a template has processed successfully
func (l *LoggerImpl) ServerTemplateProcessSuccess(a *pipe.Action, data []byte) {
	l.stats.Incr("handler.message.template.success", []string{"output:" + a.Output}, 1)
	l.logger.Debug("Transformed input template", "output", a.Output, "template")
}

func (l *LoggerImpl) ProviderConnectionFailed(*providers.Provider)                     {}
func (l *LoggerImpl) ProviderConnectionCreated(*providers.Provider)                    {}
func (l *LoggerImpl) ProviderSubcriptionFailed(*providers.Provider)                    {}
func (l *LoggerImpl) ProviderSubcriptionCreated(*providers.Provider)                   {}
func (l *LoggerImpl) ProviderMessagePublished(*providers.Provider, *providers.Message) {}
