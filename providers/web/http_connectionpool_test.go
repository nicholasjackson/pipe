package web

import (
	"testing"

	"github.com/matryer/is"
)

func TestCreatesNewConnectionFromEmptyPool(t *testing.T) {
	is := is.New(t)
	hp := NewHTTPConnectionPool()

	c, err := hp.GetConnection("localhost", 19999)

	is.NoErr(err)                         // should not have returned an error
	is.True(c != nil)                     // should have created a new server
	is.Equal("localhost:19999", c.Addr()) // should have set the correct address
}

func TestGetsExistingConnectionFromPool(t *testing.T) {
	is := is.New(t)
	hp := NewHTTPConnectionPool()

	c1, err := hp.GetConnection("localhost", 19999)
	is.NoErr(err) // should not have returned an error

	c2, err := hp.GetConnection("localhost", 19999)
	is.NoErr(err) // should not have returned an error

	is.Equal(c1, c2) // connection 1 and 2 should be the same object
}

func TestReturnsErrorOnNewConnectionFailedHealthCheck(t *testing.T) {
	t.Log("Not implemented")
}
