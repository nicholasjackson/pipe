package nats

import (
	"testing"

	"github.com/matryer/is"
)

func TestTypeEqualsNatsQueue(t *testing.T) {
	is := is.New(t)

	p := StreamingProvider{}

	is.Equal("nats_queue", p.Type())
}
