package config

import (
	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/pipe/pipe"
)

func SetupPipes(c Config, logger hclog.Logger, stats *statsd.Client) map[string]*pipe.Pipe {
	for _, p := range c.Pipes {
		ip := c.Inputs[p.Input]
		ip.Setup(c.ConnectionPools[ip.Type()], logger, stats)
		p.InputProvider = ip

		ap := c.Outputs[p.Action.Output]
		ap.Setup(c.ConnectionPools[ap.Type()], logger, stats)
		p.Action.OutputProvider = ap

		for n, s := range p.OnSuccess {
			sp := c.Outputs[s.Output]
			sp.Setup(c.ConnectionPools[sp.Type()], logger, stats)
			p.OnSuccess[n].OutputProvider = sp
		}

		for n, f := range p.OnFail {
			fp := c.Outputs[f.Output]
			fp.Setup(c.ConnectionPools[fp.Type()], logger, stats)
			p.OnFail[n].OutputProvider = fp
		}
	}

	return c.Pipes
}
