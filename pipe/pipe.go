package pipe

import (
	"time"

	"github.com/nicholasjackson/pipe/providers"
)

type Pipe struct {
	Name string

	Input         string `hcl:"input"`
	InputProvider providers.Provider

	Expiration         string `hcl:"expiration,optional"`
	ExpirationDuration time.Duration

	Action    Action   `hcl:"action,block"`
	OnSuccess []Action `hcl:"on_success,block"`
	OnFail    []Action `hcl:"on_fail,block"`
}

type Action struct {
	Output         string `hcl:"output"`
	OutputProvider providers.Provider

	Template string `hcl:"template,optional"`
}

// SetName implements NamedBlock interface
func (p *Pipe) SetName(v string) {
	p.Name = v
}
