package config

// Field is a generic configuration field for structs
type Field[T any] struct {
	// FlagName is the name used for the command line flag
	FlagName string

	// FlagDescription is the description used for the command line flag
	FlagDescription string

	// EnVarName is the name used for the environment variable
	EnVarName string

	// Value is the value of the configuration item
	Value T
}

// NewField creates a new configuration field
func NewField[T any](flagName string, enVarName string, flagDescription string, value T) Field[T] {
	return Field[T]{
		FlagName:        flagName,
		FlagDescription: flagDescription + ", EnvVar: " + enVarName,
		EnVarName:       enVarName,
		Value:           value,
	}
}
