package config

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

// Config defines the stucture which the config file wil be parsed into
type Config struct {
	Nats          string     `yaml:"nats"`
	NatsClusterID string     `yaml:"nats_cluster_id"`
	Gateway       string     `yaml:"gateway"`
	StatsD        string     `yaml:"statsd"`
	LogLevel      string     `yaml:"log_level"`
	LogFormat     string     `yaml:"log_format"`
	Functions     []Function `yaml:"functions"`
}

// Function contains the config for a particular function
type Function struct {
	Name            string           `yaml:"name"`
	FunctionName    string           `yaml:"function_name"`
	Query           string           `yaml:"query_string"`
	Message         string           `yaml:"message"`
	Expiration      string           `yaml:"expiration"`
	SuccessMessages []SuccessMessage `yaml:"success_messages"`
	InputTemplate   string           `yaml:"input_template"`
}

// SuccessMessage is a structure containing the details of a message to broadcast on success
type SuccessMessage struct {
	Name           string `yaml:"name"`
	OutputTemplate string `yaml:"output_template"`
}

// Unmarshal parses a slice of bytes into the config template
func (c *Config) Unmarshal(data []byte) error {
	err := yaml.Unmarshal(data, c)
	if err != nil {
		return fmt.Errorf("Unable to read config: %s", err)
	}

	return nil
}
