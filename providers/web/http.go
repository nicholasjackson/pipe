package web

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/nicholasjackson/pipe/logger"
	"github.com/nicholasjackson/pipe/providers"
)

type HTTPProvider struct {
	name      string
	direction string
	Protocol  string               `hcl:"protocol,optional"` // default to http
	Server    string               `hcl:"server"`
	Port      int                  `hcl:"port,optional"`   // default to 80
	Method    string               `hcl:"method,optional"` // default to POST
	Path      string               `hcl:"path,optional"`   // default to /
	TLS       *providers.TLS       `hcl:"tls_config,block"`
	AuthBasic *providers.AuthBasic `hcl:"auth_basic,block"`
	AuthMTLS  *providers.AuthMTLS  `hcl:"auth_mtls,block"`

	pool       ConnectionPool
	connection Connection
	log        logger.Logger
	msgChannel chan *providers.Message
}

func NewHTTPProvider(
	name, direction string,
	cp ConnectionPool, l logger.Logger) *HTTPProvider {

	return &HTTPProvider{
		name:      name,
		direction: direction,
		log:       l,
		pool:      cp,
		Method:    "POST",
		Path:      "/",
		Port:      80,
		Protocol:  "http",
	}
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

func (h *HTTPProvider) Setup() error {
	// do not get a connection from the pool if this is an output provider
	if h.direction == providers.DirectionOutput {
		return nil
	}

	h.msgChannel = make(chan *providers.Message, 1)
	c, err := h.pool.GetConnection(h.Server, h.Port)
	if err != nil {
		h.log.ProviderConnectionFailed(h, err)
		return err
	}

	h.log.ProviderConnectionCreated(h)
	h.connection = c

	return nil
}

func (h *HTTPProvider) Listen() (<-chan *providers.Message, error) {
	// do not listen if this is an ouput provider
	if h.direction == providers.DirectionOutput {
		return nil, nil
	}

	// should be configureable
	err := h.connection.ListenPath(h.Path, "POST", h.messageHandler)

	if err != nil {
		h.log.ProviderSubcriptionFailed(h, err)
		return nil, err
	}

	h.log.ProviderSubcriptionCreated(h)
	return h.msgChannel, err
}

func (h *HTTPProvider) Publish(msg providers.Message) (providers.Message, error) {
	h.log.ProviderMessagePublished(h, &msg)

	url := fmt.Sprintf("%s://%s:%d%s", h.Protocol, h.Server, h.Port, h.Path)
	resp, err := http.Post(url, "text/plain", bytes.NewReader(msg.Data))
	if err != nil {
		return providers.Message{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return providers.Message{}, fmt.Errorf("Got error code %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return providers.Message{}, err
	}

	return providers.Message{Data: data}, err
}

func (h *HTTPProvider) Stop() error {
	return nil
}

func (h *HTTPProvider) messageHandler(rw http.ResponseWriter, r *http.Request) {
	m := providers.NewMessage()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	m.ContentType = r.Header.Get("content-type")
	m.Data = data
	m.Redelivered = false
	m.Sequence = 1
	m.Timestamp = time.Now().UnixNano()

	h.msgChannel <- &m
}
