package config

// Config defines the stucture which the config file wil be parsed into
type Config struct {
	Nats      string     `yaml:"nats"`
	Gateway   string     `yaml:"gateway"`
	Functions []Function `yaml:"functions"`
}

// Function contains the config for a particular function
type Function struct {
	Name           string    `yaml:"name"`
	FunctionName   string    `yaml:"function_name"`
	Message        string    `yaml:"message"`
	SuccessMessage string    `yaml:"success_message"`
	Templates      Templates `yaml:"templates"`
}

// Templates is a structure containing the input and output transformation templates
type Templates struct {
	InputTemplate  string `yaml:"input_template"`
	OutputTemplate string `yaml:"output_template"`
}
