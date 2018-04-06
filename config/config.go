package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/nicholasjackson/pipe/logger"
	"github.com/nicholasjackson/pipe/pipe"
	"github.com/nicholasjackson/pipe/providers"
	nats "github.com/nicholasjackson/pipe/providers/nats_io"
	"github.com/nicholasjackson/pipe/providers/web"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var ctx *hcl.EvalContext

type Config struct {
	Inputs          map[string]providers.Provider
	Outputs         map[string]providers.Provider
	ConnectionPools map[string]providers.ConnectionPool
	Pipes           map[string]*pipe.Pipe
	log             logger.Logger
}

func New(l logger.Logger) *Config {
	ctx = buildContext()

	return &Config{
		ConnectionPools: make(map[string]providers.ConnectionPool),
		Inputs:          make(map[string]providers.Provider),
		Outputs:         make(map[string]providers.Provider),
		Pipes:           make(map[string]*pipe.Pipe),
		log:             l,
	}
}

func buildContext() *hcl.EvalContext {
	var EnvFunc = function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "env",
				Type:             cty.String,
				AllowDynamicType: true,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return cty.StringVal(os.Getenv(args[0].AsString())), nil
		},
	})

	ctx := &hcl.EvalContext{
		Functions: map[string]function.Function{},
	}
	ctx.Functions["env"] = EnvFunc

	return ctx
}

func ParseFolder(folder string, l logger.Logger) (*Config, error) {
	abs, _ := filepath.Abs(folder)
	c := New(l)

	// current folder
	files, err := filepath.Glob(path.Join(abs, "*.hcl"))
	if err != nil {
		fmt.Println("err")
		return c, err
	}

	// sub folders
	filesDir, err := filepath.Glob(path.Join(abs, "**/*.hcl"))
	if err != nil {
		fmt.Println("err")
		return c, err
	}

	files = append(files, filesDir...)

	for _, f := range files {
		conf, err := ParseHCLFile(f, l)
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

func ParseHCLFile(file string, l logger.Logger) (*Config, error) {
	parser := hclparse.NewParser()
	config := New(l)

	f, diag := parser.ParseHCLFile(file)
	if diag.HasErrors() {
		return config, errors.New(diag.Error())
	}

	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return config, errors.New("Error getting body")
	}

	for _, b := range body.Blocks {
		switch b.Type {

		case "input":
			fallthrough
		case "output":
			if err := processBody(config, b); err != nil {
				return config, err
			}

		case "pipe":
			if err := processPipe(config, b); err != nil {
				return config, err
			}
		}
	}

	return config, nil
}

func processBody(c *Config, b *hclsyntax.Block) error {
	var i providers.Provider

	switch b.Labels[0] {
	case "nats_queue":
		cp := c.ConnectionPools["nats_queue"]
		if cp == nil {
			cp = nats.NewStreamingConnectionPool()
			c.ConnectionPools["nats_queue"] = cp
		}

		i = nats.NewStreamingProvider(b.Labels[1], b.Type, cp.(nats.ConnectionPool), c.log)
	case "http":
		cp := c.ConnectionPools["http"]
		if cp == nil {
			cp = web.NewHTTPConnectionPool()
			c.ConnectionPools["http"] = cp
		}

		i = web.NewHTTPProvider(b.Labels[1], b.Type, cp.(web.ConnectionPool), c.log)
	default:
		return fmt.Errorf("Provider %s, is not a known provider", b.Labels[0])
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

	// validate the expiration
	exp := 48 * time.Hour
	if p.Expiration != "" {
		var err error
		exp, err = time.ParseDuration(p.Expiration)
		if err != nil {
			return err
		}
	}

	p.ExpirationDuration = exp
	p.Name = b.Labels[0]

	c.Pipes[b.Labels[0]] = &p

	return nil
}

func decodeBody(b *hclsyntax.Block, p interface{}) error {
	diag := gohcl.DecodeBody(b.Body, ctx, p)
	if diag.HasErrors() {
		return errors.New(diag.Error())
	}

	return nil
}
