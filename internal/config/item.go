package config

// Item is a generic configuration item
type Item[T any] struct {
	// FlagName is the name used for the command line flag
	FlagName string

	// EnVarName is the name used for the environment variable
	EnVarName string

	// Value is the value of the configuration item
	Value T
}

// NewItem creates a new configuration item
func NewItem[T any](flagName string, enVarName string, value T) Item[T] {
	return Item[T]{
		FlagName:  flagName,
		EnVarName: enVarName,
		Value:     value,
	}
}
