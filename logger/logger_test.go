package logger

import (
	"testing"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/pipe/providers"
)

func TestProviderMessagePublishedAcceptsArgs(t *testing.T) {
	l := New(hclog.New(&hclog.LoggerOptions{Level: hclog.Debug}), nil)

	mockedProvider := &providers.ProviderMock{
		NameFunc: func() string {
			return "test_provider"
		},
		TypeFunc: func() string {
			return "test_type"
		},
	}

	msg := &providers.Message{}

	args := []interface{}{"queue", "test_queue"}
	l.ProviderMessagePublished(mockedProvider, msg, args...)
}
