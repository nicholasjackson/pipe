package config

import (
	"errors"
	"path"
	"path/filepath"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/nicholasjackson/faas-nats/pipe"
	"github.com/nicholasjackson/faas-nats/providers"
	"github.com/nicholasjackson/faas-nats/providers/http"
	nats "github.com/nicholasjackson/faas-nats/providers/nats_io"
)

type Config struct {
	Inputs          map[string]providers.Provider
	Outputs         map[string]providers.Provider
	Pipes           map[string]*pipe.Pipe
	ConnectionPools map[string]providers.ConnectionPool
}

func New() Config {
	return Config{
		ConnectionPools: make(map[string]providers.ConnectionPool),
		Inputs:          make(map[string]providers.Provider),
		Outputs:         make(map[string]providers.Provider),
		Pipes:           make(map[string]*pipe.Pipe),
	}
}

func ParseFolder(folder string) (Config, error) {
	c := New()

	files, err := filepath.Glob(path.Join(folder, "**/*.hcl"))
	if err != nil {
		return c, err
	}

	for _, f := range files {
		conf, err := ParseHCLFile(f)
		if err != nil {
			return c, err
		}

		for k, v := range conf.Pipes {
			c.Pipes[k] = v
		}
		for k, v := range conf.Inputs {
			c.Inputs[k] = v
		}
		for k, v := range conf.Outputs {
			c.Outputs[k] = v
		}
		for k, v := range conf.ConnectionPools {
			c.ConnectionPools[k] = v
		}
	}

	return c, nil
}

func ParseHCLFile(file string) (Config, error) {
	parser := hclparse.NewParser()
	config := New()

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
		fallthrough
	case "output":
		if err := processBody(&config, b); err != nil {
			return config, err
		}

	case "pipe":
		if err := processPipe(&config, b); err != nil {
			return config, err
		}
	}

	return config, nil
}

func processBody(c *Config, b *hclsyntax.Block) error {
	var i providers.Provider

	switch b.Labels[0] {
	case "nats_queue":
		i = &nats.StreamingProvider{}
		if c.ConnectionPools["nats_queue"] == nil {
			c.ConnectionPools["nats_queue"] = &nats.StreamingConnectionPool{}
		}
	case "http":
		i = &http.HTTPProvider{}
		if c.ConnectionPools["http"] == nil {
			c.ConnectionPools["http"] = &http.HTTPConnectionPool{}
		}
	}

	if err := decodeBody(b, i); err != nil {
		return err
	}

	switch b.Type {
	case "input":
		c.Inputs[b.Labels[1]] = i
	case "output":
		c.Outputs[b.Labels[1]] = i
	}

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
