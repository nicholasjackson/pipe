package nats

import (
	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/faas-nats/providers"
)

type StreamingProvider struct {
	Name      string
	Server    string    `hcl:"server"`
	ClusterID string    `hcl:"cluster_id"`
	Queue     string    `hcl:"queue"`
	AuthBasic AuthBasic `hcl:"auth_basic,block"`
	AuthMTLS  AuthMTLS  `hcl:"auth_mtls,block"`
}

type AuthBasic struct {
	User     string `hcl:"user"`
	Password string `hcl:"password"`
}

type AuthMTLS struct {
	TLSClientKey    string `hcl:"tls_client_key"`
	TLSClientCert   string `hcl:"tls_client_cert"`
	TLSClientCACert string `hcl:"tls_client_cacert"`
}

func (sp *StreamingProvider) Type() string {
	return "nats_queue"
}

func (sp *StreamingProvider) Setup(cp providers.ConnectionPool, logger hclog.Logger, stats *statsd.Client) error {
	pool := cp.(*StreamingConnectionPool)

	_, err := pool.GetConnection(sp.Server, sp.ClusterID)
	if err != nil {
		stats.Incr("connection.nats.failed", nil, 1)
		logger.Error("Unable to connect to nats server", "error", err)
	}
	return nil
}

func (sp *StreamingProvider) Listen() (<-chan *providers.Message, error) {
	return nil, nil
}

func (sp *StreamingProvider) Stop() error {
	return nil
}