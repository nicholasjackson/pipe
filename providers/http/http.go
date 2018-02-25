package http

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/pipe/providers"
)

type HTTPProvider struct {
	name      string
	Protocol  string               `hcl:"protocol,optional"` // default to http
	Server    string               `hcl:"server"`
	Port      int                  `hcl:"port,optional"` // default to 80
	Path      string               `hcl:"path,optional"` // default to /
	TLS       *providers.TLS       `hcl:"tls_config,block"`
	AuthBasic *providers.AuthBasic `hcl:"auth_basic,block"`
	AuthMTLS  *providers.AuthMTLS  `hcl:"auth_mtls,block"`

	stats      *statsd.Client
	logger     hclog.Logger
	msgChannel chan *providers.Message
}

func (sp *HTTPProvider) Name() string {
	return sp.name
}
func (h *HTTPProvider) Type() string {
	return "http"
}

func (h *HTTPProvider) Setup(cp providers.ConnectionPool, log hclog.Logger, stats *statsd.Client) error {
	h.stats = stats
	h.logger = log

	return nil
}

func (h *HTTPProvider) Listen() (<-chan *providers.Message, error) {
	return nil, nil
}

func (h *HTTPProvider) Publish(d []byte) error {
	url := fmt.Sprintf("%s://%s:%d%s", h.Protocol, h.Server, h.Port, h.Path)
	resp, err := http.Post(url, "text/plain", bytes.NewReader(d))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Got error code %d", resp.StatusCode)
	}

	return nil
}

func (h *HTTPProvider) Stop() error {
	return nil
}
