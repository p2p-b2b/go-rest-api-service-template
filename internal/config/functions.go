package config

type Validator interface {
	Validate() error
}

type EnvVarsParser interface {
	ParseEnvVars()
}

// Validate validates the configuration values
func Validate(configs ...Validator) error {
	for _, config := range configs {
		if err := config.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ParseEnvVars reads the configuration from environment variables
// and sets the values in the configuration
func ParseEnvVars(configs ...EnvVarsParser) {
	for _, config := range configs {
		config.ParseEnvVars()
	}
}
