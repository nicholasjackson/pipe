package config

import (
	"errors"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/nicholasjackson/faas-nats/pipe"
	"github.com/nicholasjackson/faas-nats/providers"
	nats "github.com/nicholasjackson/faas-nats/providers/nats_io"
)

type Config struct {
	Inputs          map[string]providers.Provider
	Outputs         map[string]providers.Provider
	Pipes           map[string]*pipe.Pipe
	ConnectionPools map[string]providers.ConnectionPool
}

func ParseHCLFile(file string) (Config, error) {
	parser := hclparse.NewParser()

	config := Config{
		ConnectionPools: make(map[string]providers.ConnectionPool),
		Inputs:          make(map[string]providers.Provider),
		Outputs:         make(map[string]providers.Provider),
		Pipes:           make(map[string]*pipe.Pipe),
	}

	f, diag := parser.ParseHCLFile(file)
	if diag.HasErrors() {
		return config, errors.New(diag.Error())
	}

	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return config, errors.New("Error getting body")
	}

	b := body.Blocks[0]

	switch b.Type {

	case "input":
		if err := processInput(&config, b); err != nil {
			return config, err
		}

	case "pipe":
		if err := processPipe(&config, b); err != nil {
			return config, err
		}
	}

	return config, nil
}

func processInput(c *Config, b *hclsyntax.Block) error {
	var i providers.Provider

	switch b.Labels[0] {
	case "nats_queue":
		i = &nats.StreamingProvider{}
		if c.ConnectionPools["nats_queue"] == nil {
			c.ConnectionPools["nats_queue"] = &nats.StreamingConnectionPool{}
		}
	}

	if err := decodeBody(b, i); err != nil {
		return err
	}

	c.Inputs[b.Labels[1]] = i

	return nil
}

func processPipe(c *Config, b *hclsyntax.Block) error {
	p := pipe.Pipe{}

	if err := decodeBody(b, &p); err != nil {
		return err
	}

	c.Pipes[b.Labels[0]] = &p

	return nil
}

func decodeBody(b *hclsyntax.Block, p interface{}) error {
	diag := gohcl.DecodeBody(b.Body, nil, p)
	if diag.HasErrors() {
		return errors.New(diag.Error())
	}

	return nil
}
