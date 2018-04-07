package config

import (
	"fmt"

	"github.com/nicholasjackson/pipe/logger"
	"github.com/nicholasjackson/pipe/pipe"
)

func SetupPipes(c *Config, l logger.Logger) (map[string]*pipe.Pipe, error) {
	var errs []error
	for _, i := range c.Inputs {
		i.Setup()
	}

	for _, i := range c.Outputs {
		i.Setup()
	}

	for _, p := range c.Pipes {
		l.GetLogger().Info("Configure", "pipe", p.Name)

		ip := c.Inputs[p.Input]
		if ip == nil {
			errs = append(errs, fmt.Errorf("No input provider %s defined for pipe %s\n", p.Input, p.Name))
		}

		p.InputProvider = ip

		ap := c.Outputs[p.Action.Output]
		if ap == nil {
			errs = append(errs, fmt.Errorf("No output provider %s defined for pipe %s\n", p.Action.Output, p.Name))
		}

		p.Action.OutputProvider = ap

		for n, s := range p.OnSuccess {
			sp := c.Outputs[s.Output]
			if sp == nil {
				errs = append(errs, fmt.Errorf("No success output provider %s defined for pipe %s\n", s.Output, p.Name))
			}

			p.OnSuccess[n].OutputProvider = sp
		}

		for n, f := range p.OnFail {
			fp := c.Outputs[f.Output]
			if fp == nil {
				errs = append(errs, fmt.Errorf("No fail output provider %s defined for pipe %s\n", f.Output, p.Name))
			}

			p.OnFail[n].OutputProvider = fp
		}
	}

	if len(errs) > 0 {
		message := ""
		for _, e := range errs {
			message += e.Error()
		}

		return nil, fmt.Errorf(message)
	}

	return c.Pipes, nil
}
