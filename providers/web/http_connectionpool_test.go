package web

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/matryer/is"
)

func TestCreatesNewConnectionFromEmptyPool(t *testing.T) {
	is := is.New(t)
	hp := NewHTTPConnectionPool()

	port := rand.Intn(10000) + 50000 // generate random port between 50000 - 60000
	c, err := hp.GetConnection("localhost", port)

	is.NoErr(err)                                         // should not have returned an error
	is.True(c != nil)                                     // should have created a new server
	is.Equal(fmt.Sprintf("localhost:%d", port), c.Addr()) // should have set the correct address
}

func TestGetsExistingConnectionFromPool(t *testing.T) {
	is := is.New(t)
	hp := NewHTTPConnectionPool()
	port := rand.Intn(10000) + 50000 // generate random port between 50000 - 60000

	c1, err := hp.GetConnection("localhost", port)
	is.NoErr(err) // should not have returned an error

	c2, err := hp.GetConnection("localhost", port)
	is.NoErr(err) // should not have returned an error

	is.Equal(c1, c2) // connection 1 and 2 should be the same object
}

func TestReturnsErrorOnNewConnectionFailedHealthCheck(t *testing.T) {
	t.Log("Not implemented")
}
