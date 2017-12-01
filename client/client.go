package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
)

//go:generate moq -out mock_client.go . Client

// Client defines an interface for a function client
type Client interface {
	CallFunction(name string, payload []byte) ([]byte, error)
}

// Impl is a simple client for calling functions
type Impl struct {
	gateway string
	stats   *statsd.Client
	logger  hclog.Logger
}

// NewClient creates a new client
func NewClient(gateway string, stats *statsd.Client, logger hclog.Logger) *Impl {
	return &Impl{gateway, stats, logger}
}

// CallFunction calls the function and returns the response
func (c *Impl) CallFunction(name string, payload []byte) ([]byte, error) {
	startTime := time.Now()
	defer func(st time.Time) {
		dur := time.Now().Sub(st)
		c.stats.Timing("gateway.call.time", dur, []string{"function:" + name}, 1)
	}(startTime)

	c.stats.Incr("gateway.call.do", []string{"function:" + name}, 1)

	resp, err := http.Post(c.gateway+"/function/"+name, "", bytes.NewBuffer(payload))
	if err != nil {
		c.stats.Incr("gateway.call.failed", []string{"function:" + name}, 1)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.stats.Incr("gateway.call.success", []string{"function:" + name}, 1)
		return nil, fmt.Errorf("Invalid repsponse code from function: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}
