package config

import "github.com/nicholasjackson/faas-nats/pipe"

func SetupPipes(c Config) map[string]*pipe.Pipe {
	for _, p := range c.Pipes {
		ip := c.Inputs[p.Input]
		ip.Setup(c.ConnectionPools[ip.Type()])
		p.InputProvider = ip

		ap := c.Outputs[p.Action.Output]
		ap.Setup(c.ConnectionPools[ap.Type()])
		p.Action.OutputProvider = ap

		for n, s := range p.OnSuccess {
			sp := c.Outputs[s.Output]
			sp.Setup(c.ConnectionPools[sp.Type()])
			p.OnSuccess[n].OutputProvider = sp
		}

		for n, f := range p.OnFail {
			fp := c.Outputs[f.Output]
			fp.Setup(c.ConnectionPools[fp.Type()])
			p.OnFail[n].OutputProvider = fp
		}
	}

	return c.Pipes
}
