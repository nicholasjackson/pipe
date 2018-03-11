package web

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/pipe/providers"
)

type HTTPProvider struct {
	name      string
	direction string
	Protocol  string               `hcl:"protocol,optional"` // default to http
	Server    string               `hcl:"server"`
	Port      int                  `hcl:"port,optional"` // default to 80
	Path      string               `hcl:"path,optional"` // default to /
	TLS       *providers.TLS       `hcl:"tls_config,block"`
	AuthBasic *providers.AuthBasic `hcl:"auth_basic,block"`
	AuthMTLS  *providers.AuthMTLS  `hcl:"auth_mtls,block"`

	connection Connection
	stats      *statsd.Client
	logger     hclog.Logger
	msgChannel chan *providers.Message
}

func NewHTTPProvider(name string, direction string) *HTTPProvider {
	return &HTTPProvider{name: name, direction: direction}
}

func (sp *HTTPProvider) Name() string {
	return sp.name
}

func (h *HTTPProvider) Type() string {
	return "http"
}

func (h *HTTPProvider) Direction() string {
	return h.direction
}

func (h *HTTPProvider) Setup(cp providers.ConnectionPool, log hclog.Logger, stats *statsd.Client) error {
	h.stats = stats
	h.logger = log

	if h.direction == providers.DirectionOutput {
		return nil
	}

	h.msgChannel = make(chan *providers.Message, 1)
	pool := cp.(ConnectionPool)
	c, err := pool.GetConnection(h.Server, h.Port)
	if err != nil {
		h.stats.Incr("connection.http.failed", nil, 1)
		h.logger.Error("Unable to create http server", "error", err)
		return err
	}

	h.stats.Incr("connection.http.created", nil, 1)
	h.logger.Info("Created http connection for", "server", h.Server, "protocol", h.Protocol, "port", h.Port, "path", h.Path)
	h.connection = c

	return nil
}

func (h *HTTPProvider) Listen() (<-chan *providers.Message, error) {
	if h.direction == providers.DirectionOutput {
		return nil, nil
	}

	err := h.connection.ListenPath(h.Path, "GET", h.messageHandler)

	if err != nil {
		h.stats.Incr("listen.http.failed", nil, 1)
		h.logger.Error("Failed to create http listener for", "server", h.Server, "protocol", h.Protocol, "port", h.Port, "path", h.Path)
		return nil, err
	}

	h.stats.Incr("listen.http.created", nil, 1)
	h.logger.Debug("Created http listener for", "server", h.Server, "protocol", h.Protocol, "port", h.Port, "path", h.Path)

	return h.msgChannel, err
}

func (h *HTTPProvider) Publish(d []byte) ([]byte, error) {
	h.logger.Debug("Publishing message for", "server", h.Server, "protocol", h.Protocol, "port", h.Port, "path", h.Path)
	h.stats.Incr("publish.http.call", nil, 1)

	url := fmt.Sprintf("%s://%s:%d%s", h.Protocol, h.Server, h.Port, h.Path)
	resp, err := http.Post(url, "text/plain", bytes.NewReader(d))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Got error code %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, err
}

func (h *HTTPProvider) Stop() error {
	return nil
}

func (h *HTTPProvider) messageHandler(rw http.ResponseWriter, r *http.Request) {
	m := providers.Message{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	m.Data = data
	m.Redelivered = false
	m.Sequence = 1
	m.Timestamp = time.Now().UnixNano()

	h.msgChannel <- &m
}
