package config

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
	Name           string    `yaml:"name"`
	FunctionName   string    `yaml:"function_name"`
	Query          string    `yaml:"query_string"`
	Message        string    `yaml:"message"`
	SuccessMessage string    `yaml:"success_message"`
	Templates      Templates `yaml:"templates"`
}

// Templates is a structure containing the input and output transformation templates
type Templates struct {
	InputTemplate  string `yaml:"input_template"`
	OutputTemplate string `yaml:"output_template"`
}
