package web

import (
	"testing"

	"github.com/matryer/is"
)

func TestGetsExistingConnectionFromPool(t *testing.T) {
	is := is.New(t)
	hp := HTTPConnectionPool{
		connections: make(map[string]Connection),
	}
	server := &HTTPConnection{}
	hp.connections["localhost_19999"] = server

	c, err := hp.GetConnection("localhost", 19999)

	is.NoErr(err)       // should not have returned an error
	is.Equal(server, c) // should have returned server from pool
}

func TestCreatesNewConnectionFromEmptyPool(t *testing.T) {
	is := is.New(t)
	hp := HTTPConnectionPool{}

	c, err := hp.GetConnection("localhost", 19999)

	is.NoErr(err)                         // should not have returned an error
	is.True(c != nil)                     // should have created a new server
	is.Equal("localhost:19999", c.Addr()) // should have set the correct address
}
