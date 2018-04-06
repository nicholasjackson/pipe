package config

import (
	"testing"

	"github.com/matryer/is"
	"github.com/nicholasjackson/pipe/logger"
	"github.com/nicholasjackson/pipe/pipe"
	"github.com/nicholasjackson/pipe/providers"
)

func buildConfig() (*Config, *providers.ProviderMock) {
	providerMock := &providers.ProviderMock{
		ListenFunc: func() (<-chan *providers.Message, error) {
			panic("TODO: mock out the Listen method")
		},
		SetupFunc: func() error {
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

func setupPipes(t *testing.T) (*is.I, *Config, logger.Logger, *providers.ProviderMock) {
	c, pm := buildConfig()
	l := testGetLogger()

	return is.New(t), c, l, pm
}

func TestSetupPipesCallsSetupOnTheOutputProviders(t *testing.T) {
	is, c, l, m := setupPipes(t)

	SetupPipes(c, l)

	is.Equal(2, len(m.SetupCalls())) // should have called setup once
}

func TestSetupPipesCallsSetupOnTheInputProviders(t *testing.T) {
	is, c, l, m := setupPipes(t)

	SetupPipes(c, l)

	is.Equal(2, len(m.SetupCalls())) // should have called setup once
}

func TestSetupPipesSetsTheCorrectInputProviderOnThePipe(t *testing.T) {
	is, c, l, _ := setupPipes(t)

	p, err := SetupPipes(c, l)

	is.NoErr(err)
	is.Equal(1, len(p))                                               // should have created one pipe
	is.Equal(c.Inputs["test_provider"], p["test_pipe"].InputProvider) // should have set the correct input provider
}

func TestSetupPipesSetsTheCorrectActionOutputProviderOnThePipe(t *testing.T) {
	is, c, l, _ := setupPipes(t)

	p, err := SetupPipes(c, l)

	is.NoErr(err)
	is.Equal(1, len(p))                                                        // should have created one pipe
	is.Equal(c.Outputs["test_provider"], p["test_pipe"].Action.OutputProvider) // should have set the correct output provider
}

func TestSetupPipesSetsTheCorrectSuccessOutputProviderOnThePipe(t *testing.T) {
	is, c, l, _ := setupPipes(t)

	p, err := SetupPipes(c, l)

	is.NoErr(err)
	is.Equal(1, len(p))                                                              // should have created one pipe
	is.Equal(c.Outputs["test_provider"], p["test_pipe"].OnSuccess[0].OutputProvider) // should have set the correct output provider
}

func TestSetupPipesSetsTheCorrectFailOutputProviderOnThePipe(t *testing.T) {
	is, c, l, _ := setupPipes(t)

	p, err := SetupPipes(c, l)

	is.NoErr(err)
	is.Equal(1, len(p))                                                           // should have created one pipe
	is.Equal(c.Outputs["test_provider"], p["test_pipe"].OnFail[0].OutputProvider) // should have set the correct output provider
}

func TestSetupPipesReturnsErrorWhenNoOutputFound(t *testing.T) {
	is, c, l, _ := setupPipes(t)
	c.Outputs = make(map[string]providers.Provider)

	_, err := SetupPipes(c, l)

	is.True(err != nil) // should have returned and error
}

func TestSetupPipesReturnsErrorWhenNoInputFound(t *testing.T) {
	is, c, l, _ := setupPipes(t)
	c.Inputs = make(map[string]providers.Provider)

	_, err := SetupPipes(c, l)

	is.True(err != nil) // should have returned and error
}

func TestSetupPipesReturnsErrorWhenNoSuccessOutputFound(t *testing.T) {
	is, c, l, _ := setupPipes(t)
	c.Pipes["test_pipe"].OnSuccess[0].Output = "does not exist"

	_, err := SetupPipes(c, l)

	is.True(err != nil) // should have returned and error
}

func TestSetupPipesReturnsErrorWhenNoFailOutputFound(t *testing.T) {
	is, c, l, _ := setupPipes(t)
	c.Pipes["test_pipe"].OnFail[0].Output = "does not exist"

	_, err := SetupPipes(c, l)

	is.True(err != nil) // should have returned and error
}
