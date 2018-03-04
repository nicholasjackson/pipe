package helpers

import (
	"bytes"
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/kr/pretty"
	"github.com/nicholasjackson/pipe/config"
	"github.com/nicholasjackson/pipe/server"
)

func StartServer(configFolder string) (*server.PipeServer, *bytes.Buffer, error) {
	c, err := config.ParseFolder(configFolder)
	if err != nil {
		return nil, nil, err
	}

	if len(c.Pipes) < 1 || len(c.Inputs) < 1 || len(c.Outputs) < 1 {
		return nil, nil, fmt.Errorf("Ensure config has at least 1 pipe, 1 input, and 1 output: %s", pretty.Sprint(c))
	}

	buff := bytes.NewBuffer([]byte{})

	lo := hclog.DefaultOptions
	lo.Level = hclog.Trace
	lo.Output = buff

	l := hclog.New(lo)
	stats, _ := statsd.New("localhost:8125")

	c.Pipes, _ = config.SetupPipes(c, l, stats)

	s := server.New(c, l, stats)
	s.Listen()

	return s, buff, nil
}
