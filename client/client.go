package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

//go:generate moq -out mock_client.go . Client

// Client defines an interface for a function client
type Client interface {
	CallFunction(name string, payload []byte) ([]byte, error)
}

// Impl is a simple client for calling functions
type Impl struct {
	gateway string
}

// NewClient creates a new client
func NewClient(gateway string) *Impl {
	return &Impl{gateway}
}

// CallFunction calls the function and returns the response
func (c *Impl) CallFunction(name string, payload []byte) ([]byte, error) {
	resp, err := http.Post(c.gateway+"/function/"+name, "", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Invalid repsponse code from function: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}
