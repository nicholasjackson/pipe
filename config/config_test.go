package config

import (
	"testing"

	"github.com/matryer/is"
	nats "github.com/nicholasjackson/faas-nats/providers/nats_io"
)

func TestParsesConfigPipeHCL(t *testing.T) {
	is := is.New(t)

	c, err := ParseHCLFile("../test_fixtures/pipe/standard.hcl")

	is.NoErr(err)             // error should have been nil
	is.Equal(1, len(c.Pipes)) // should have returned one pipe
}

func TestParsesConfigPipeHCLNoFail(t *testing.T) {
	is := is.New(t)

	c, err := ParseHCLFile("../test_fixtures/pipe/no_fail.hcl")

	is.NoErr(err)                                     // error should have been nil
	is.Equal(1, len(c.Pipes))                         // should have returned one pipe
	is.Equal(0, len(c.Pipes["process_image"].OnFail)) // should have returned 0 fail blocks
}

func TestParsesConfigPipeHCLNoSuccess(t *testing.T) {
	is := is.New(t)

	c, err := ParseHCLFile("../test_fixtures/pipe/no_success.hcl")

	is.NoErr(err)                                        // error should have been nil
	is.Equal(1, len(c.Pipes))                            // should have returned one pipe
	is.Equal(0, len(c.Pipes["process_image"].OnSuccess)) // should have returned 0 success blocks
}

func TestParsesConfigNatsProviderHCL(t *testing.T) {
	is := is.New(t)

	c, err := ParseHCLFile("../test_fixtures/providers/nats_io_input.hcl")

	is.NoErr(err)              // error should have been nil
	is.Equal(1, len(c.Inputs)) // should have returned one input provider

	p, ok := c.Inputs["nats_messages_in"].(*nats.StreamingProvider)
	is.True(ok) // should have returned a StreamingProvider

	is.Equal("nats://myserver.com", p.Server) // should have set the server to a valid value
	is.Equal("abc123", p.ClusterID)           // should have set the cluster id to a valid value
	is.Equal("mymessagequeue", p.Queue)       // should have set the messagequeue to a valid value

	is.Equal("xxx", p.AuthBasic.User)     // should have set basic auth user
	is.Equal("xxx", p.AuthBasic.Password) // should have set basic auth password

	is.Equal("cacert", p.AuthMTLS.TLSClientCert)     // should have set mtls auth client cert
	is.Equal("cakey", p.AuthMTLS.TLSClientKey)       // should have set mtls auth client key
	is.Equal("caclient", p.AuthMTLS.TLSClientCACert) // should have set mtls auth ca cert

	is.True(c.ConnectionPools["nats_queue"] != nil) // should have created a connection pool

	_, ok = c.ConnectionPools["nats_queue"].(*nats.StreamingConnectionPool)
	is.True(ok) // should have create a connection pool of type nats.StreamingConnectionPool
}
