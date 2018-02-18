package config

import (
	"testing"

	"github.com/matryer/is"
	"github.com/nicholasjackson/faas-nats/pipe"
	"github.com/nicholasjackson/faas-nats/providers"
)

func buildConfig() (*Config, *providers.ProviderMock) {
	providerMock := &providers.ProviderMock{
		ListenFunc: func() (<-chan *providers.Message, error) {
			panic("TODO: mock out the Listen method")
		},
		SetupFunc: func(cp providers.ConnectionPool) error {
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
	return is.New(t), c, pm
}

func TestSetupPipesSetsTheCorrectInputProviderOnThePipe(t *testing.T) {
	is, c, _ := setupPipes(t)

	p := SetupPipes(*c)

	is.Equal(1, len(p))                                               // should have created one pipe
	is.Equal(c.Inputs["test_provider"], p["test_pipe"].InputProvider) // should have set the correct input provider
}

func TestSetupPipesCallsSetupOnTheInputProvider(t *testing.T) {
	is, c, m := setupPipes(t)

	_ = SetupPipes(*c)

	is.Equal(4, len(m.SetupCalls()))                                   // should have called setup once
	is.Equal(c.ConnectionPools["mock_provider"], m.SetupCalls()[0].Cp) // should have passed the mock provider
}

func TestSetupPipesSetsTheCorrectActionOutputProviderOnThePipe(t *testing.T) {
	is, c, _ := setupPipes(t)

	p := SetupPipes(*c)

	is.Equal(1, len(p))                                                        // should have created one pipe
	is.Equal(c.Outputs["test_provider"], p["test_pipe"].Action.OutputProvider) // should have set the correct output provider
}

func TestSetupPipesCallsSetupOnTheActionOutputProvider(t *testing.T) {
	is, c, m := setupPipes(t)

	_ = SetupPipes(*c)

	is.Equal(4, len(m.SetupCalls()))                                   // should have called setup once
	is.Equal(c.ConnectionPools["mock_provider"], m.SetupCalls()[1].Cp) // should have passed the mock provider
}

func TestSetupPipesSetsTheCorrectSuccessOutputProviderOnThePipe(t *testing.T) {
	is, c, _ := setupPipes(t)

	p := SetupPipes(*c)

	is.Equal(1, len(p))                                                              // should have created one pipe
	is.Equal(c.Outputs["test_provider"], p["test_pipe"].OnSuccess[0].OutputProvider) // should have set the correct output provider
}

func TestSetupPipesCallsSetupOnTheSuccessOutputProvider(t *testing.T) {
	is, c, m := setupPipes(t)

	_ = SetupPipes(*c)

	is.Equal(4, len(m.SetupCalls()))                                   // should have called setup once
	is.Equal(c.ConnectionPools["mock_provider"], m.SetupCalls()[2].Cp) // should have passed the mock provider
}

func TestSetupPipesSetsTheCorrectFailOutputProviderOnThePipe(t *testing.T) {
	is, c, _ := setupPipes(t)

	p := SetupPipes(*c)

	is.Equal(1, len(p))                                                           // should have created one pipe
	is.Equal(c.Outputs["test_provider"], p["test_pipe"].OnFail[0].OutputProvider) // should have set the correct output provider
}

func TestSetupPipesCallsSetupOnTheFailOutputProvider(t *testing.T) {
	is, c, m := setupPipes(t)

	_ = SetupPipes(*c)

	is.Equal(4, len(m.SetupCalls()))                                   // should have called setup once
	is.Equal(c.ConnectionPools["mock_provider"], m.SetupCalls()[3].Cp) // should have passed the mock provider
}
