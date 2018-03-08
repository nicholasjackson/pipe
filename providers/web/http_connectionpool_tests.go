package web

import (
	"net/http"
	"testing"

	"github.com/matryer/is"
)

func TestGetsConnectionFromPool(t *testing.T) {
	is := is.New(t)
	hp := HTTPConnectionPool{}
	server := &http.Server{}
	hp.connections["localhost_8080"] = server

	c, err := hp.GetConnection("localhost", 8080)

	is.NoErr(err)       // should not have returned an error
	is.Equal(server, c) // should have returned server from pool
}
