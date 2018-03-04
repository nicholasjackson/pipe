package config

import (
	"testing"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/matryer/is"
	"github.com/nicholasjackson/pipe/pipe"
	"github.com/nicholasjackson/pipe/providers"
)

var testLogger hclog.Logger
var testStatsD *statsd.Client

func buildConfig() (*Config, *providers.ProviderMock) {
	providerMock := &providers.ProviderMock{
		ListenFunc: func() (<-chan *providers.Message, error) {
			panic("TODO: mock out the Listen method")
		},
		SetupFunc: func(cp providers.ConnectionPool, logger hclog.Logger, stats *statsd.Client) error {
			return nil
		},
		StopFunc: func() error {
			panic("TODO: mock out the Stop method")
		},
		TypeFunc: func() string {
			return "mock_provider"
		},
	}

	return &Config{
			Inputs: map[string]providers.Provider{
				"test_provider": providerMock,
			},
			Outputs: map[string]providers.Provider{
				"test_provider": providerMock,
			},
			Pipes: map[string]*pipe.Pipe{"test_pipe": &pipe.Pipe{
				Input: "test_provider",
				Action: pipe.Action{
					Output: "test_provider",
				},
				OnSuccess: []pipe.Action{
					pipe.Action{
						Output: "test_provider",
					},
				},
				OnFail: []pipe.Action{
					pipe.Action{
						Output: "test_provider",
					},
				},
			}},
			ConnectionPools: map[string]providers.ConnectionPool{
				"mock_provider": &providers.ConnectionPoolMock{},
			},
		},
		providerMock
}

func setupPipes(t *testing.T) (*is.I, *Config, *providers.ProviderMock) {
	c, pm := buildConfig()
	testLogger = hclog.Default()
	testStatsD, _ = statsd.New("localhost:8125")

	return is.New(t), c, pm
}

func TestSetupPipesCallsSetupOnTheOutputProviders(t *testing.T) {
	is, c, m := setupPipes(t)

	SetupPipes(c, testLogger, testStatsD)

	is.Equal(2, len(m.SetupCalls()))                                   // should have called setup once
	is.Equal(c.ConnectionPools["mock_provider"], m.SetupCalls()[0].Cp) // should have passed the mock provider
}

func TestSetupPipesCallsSetupOnTheInputProviders(t *testing.T) {
	is, c, m := setupPipes(t)

	SetupPipes(c, testLogger, testStatsD)

	is.Equal(2, len(m.SetupCalls()))                                   // should have called setup once
	is.Equal(c.ConnectionPools["mock_provider"], m.SetupCalls()[0].Cp) // should have passed the mock provider
}

func TestSetupPipesSetsTheCorrectInputProviderOnThePipe(t *testing.T) {
	is, c, _ := setupPipes(t)

	p, err := SetupPipes(c, testLogger, testStatsD)

	is.NoErr(err)
	is.Equal(1, len(p))                                               // should have created one pipe
	is.Equal(c.Inputs["test_provider"], p["test_pipe"].InputProvider) // should have set the correct input provider
}

func TestSetupPipesSetsTheCorrectActionOutputProviderOnThePipe(t *testing.T) {
	is, c, _ := setupPipes(t)

	p, err := SetupPipes(c, testLogger, testStatsD)

	is.NoErr(err)
	is.Equal(1, len(p))                                                        // should have created one pipe
	is.Equal(c.Outputs["test_provider"], p["test_pipe"].Action.OutputProvider) // should have set the correct output provider
}

func TestSetupPipesSetsTheCorrectSuccessOutputProviderOnThePipe(t *testing.T) {
	is, c, _ := setupPipes(t)

	p, err := SetupPipes(c, testLogger, testStatsD)

	is.NoErr(err)
	is.Equal(1, len(p))                                                              // should have created one pipe
	is.Equal(c.Outputs["test_provider"], p["test_pipe"].OnSuccess[0].OutputProvider) // should have set the correct output provider
}

func TestSetupPipesSetsTheCorrectFailOutputProviderOnThePipe(t *testing.T) {
	is, c, _ := setupPipes(t)

	p, err := SetupPipes(c, testLogger, testStatsD)

	is.NoErr(err)
	is.Equal(1, len(p))                                                           // should have created one pipe
	is.Equal(c.Outputs["test_provider"], p["test_pipe"].OnFail[0].OutputProvider) // should have set the correct output provider
}

func TestSetupPipesReturnsErrorWhenNoOutputFound(t *testing.T) {
	is, c, _ := setupPipes(t)
	c.Outputs = make(map[string]providers.Provider)

	_, err := SetupPipes(c, testLogger, testStatsD)

	is.True(err != nil) // should have returned and error
}

func TestSetupPipesReturnsErrorWhenNoInputFound(t *testing.T) {
	is, c, _ := setupPipes(t)
	c.Inputs = make(map[string]providers.Provider)

	_, err := SetupPipes(c, testLogger, testStatsD)

	is.True(err != nil) // should have returned and error
}

func TestSetupPipesReturnsErrorWhenNoSuccessOutputFound(t *testing.T) {
	is, c, _ := setupPipes(t)
	c.Pipes["test_pipe"].OnSuccess[0].Output = "does not exist"

	_, err := SetupPipes(c, testLogger, testStatsD)

	is.True(err != nil) // should have returned and error
}

func TestSetupPipesReturnsErrorWhenNoFailOutputFound(t *testing.T) {
	is, c, _ := setupPipes(t)
	c.Pipes["test_pipe"].OnFail[0].Output = "does not exist"

	_, err := SetupPipes(c, testLogger, testStatsD)

	is.True(err != nil) // should have returned and error
}
