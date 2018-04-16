package logger

import (
	"github.com/kr/pretty"
	"github.com/nicholasjackson/pipe/providers"
)

func (l *LoggerImpl) ProviderConnectionFailed(p providers.Provider, err error) {
	l.stats.Incr("provider.connection.failed", []string{"provider:" + p.Name(), "type:" + p.Type()}, 1)
	l.logger.Error("Unable to connect to nats server", "error", err, "provider", p.Name(), "type", p.Type())
}

func (l *LoggerImpl) ProviderConnectionCreated(p providers.Provider) {
	l.stats.Incr("provider.connection.created", []string{"provider:" + p.Name(), "type:" + p.Type()}, 1)
	l.logger.Info("Created connection for", "provider", p.Name(), "type", p.Type())
}

func (l *LoggerImpl) ProviderSubcriptionFailed(p providers.Provider, err error) {
	l.stats.Incr("provider.subscription.failed", []string{"provider:" + p.Name(), "type:" + p.Type()}, 1)
	l.logger.Error("Failed to create subscription", "error", err, "provider", p.Name(), "type", p.Type())
}

func (l *LoggerImpl) ProviderSubcriptionCreated(p providers.Provider) {
	l.stats.Incr("provider.subscription.created", []string{"provider:" + p.Name(), "type:" + p.Type()}, 1)
	l.logger.Info("Created subscription for", "provider", p.Name(), "type", p.Type())
}

func (l *LoggerImpl) ProviderMessagePublished(p providers.Provider, m *providers.Message, args ...interface{}) {
	l.stats.Incr("provider.publish.call", []string{"provider:" + p.Name(), "type:" + p.Type()}, 1)
	l.logger.Info("Publishing message", "id", m.ID, "parentid", m.ParentID, "provider", p.Name(), "type", p.Type(), args)

	l.logger.Debug("Publishing message", "id", m.ID, "parentid", m.ParentID, "provider", p.Name(), "type", p.Type(), "message", pretty.Sprint(m), "data", string(m.Data))
}
