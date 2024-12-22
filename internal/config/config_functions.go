package config

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func init() {
	// set environment variables from .env file
	if err := setEnvVarFromFile(); err != nil {
		slog.Warn("failed to set environment variables from .env file", "error", err)
	}
}

// GetEnv retrieves the value of an environment variable
// or returns a default value if the environment variable is not set
func GetEnv[T any](key string, defaultValue T) T {
	if value, exists := os.LookupEnv(key); exists {
		switch any(defaultValue).(type) {
		case string:
			return any(value).(T)
		case int, uint, uint8, uint16, uint32, uint64:
			if intValue, err := strconv.Atoi(value); err == nil {
				return any(intValue).(T)
			}
		case float32, float64:
			if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
				return any(floatValue).(T)
			}
		case time.Duration:
			if durationValue, err := time.ParseDuration(value); err == nil {
				return any(durationValue).(T)
			}
		case bool:
			if boolValue, err := strconv.ParseBool(value); err == nil {
				return any(boolValue).(T)
			}
		case FileVar:
			// create a file using the value as the path
			file, err := os.OpenFile(value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				return defaultValue
			}

			return any(FileVar{File: file, flag: os.O_APPEND | os.O_CREATE | os.O_WRONLY}).(T)
		default:
			return defaultValue
		}
	}

	return defaultValue
}

// setEnvVarFromFile loads all .env files in the current current working directory
// and sets the key-value pairs in the environment
// if there are multiple .env files, it returns an error
func setEnvVarFromFile() error {
	// Get the current working directory
	execDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Find all .env files in the execution directory
	envFiles, err := filepath.Glob(filepath.Join(execDir, "*.env"))
	if err != nil {
		return err
	}

	if len(envFiles) > 1 {
		return fmt.Errorf("multiple '.env' files found in the execution directory: %v", envFiles)
	}

	// Load each .env file
	for _, envFile := range envFiles {
		file, err := os.Open(envFile)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			// Skip comments and empty lines
			if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
				continue
			}

			// Split the line into key and value
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				slog.Warn("invalid line in .env file", "line", line, "avoiding", true)
				continue
			}

			key := strings.TrimSpace(parts[0])

			// remove quotes from value
			parts[1] = strings.Trim(parts[1], "\"")
			parts[1] = strings.Trim(parts[1], "'")

			value := strings.TrimSpace(parts[1])

			// Set the environment variable
			os.Setenv(key, value)
		}

		if err := scanner.Err(); err != nil {
			return err
		}
	}

	return nil
}

// Validator is an interface for validating configuration values
// Implement this interface for configuration structs
// and add the validation logic in the Validate method
type Validator interface {
	Validate() error
}

// EnvVarsParser is an interface for parsing configuration values from environment variables
// Implement this interface for configuration structs
// and add the parsing logic in the ParseEnvVars method
type EnvVarsParser interface {
	ParseEnvVars()
}

// Validate validates the configuration values
// by calling the Validate method of each configuration struct
// and returns the first error encountered
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
// by calling the ParseEnvVars method of each configuration struct
func ParseEnvVars(configs ...EnvVarsParser) {
	for _, config := range configs {
		config.ParseEnvVars()
	}
}

// FileVar is a custom flag type for files
// This should implement the Value interface of the flag package
// Reference: https://pkg.go.dev/gg-scm.io/tool/internal/flag#FlagSet.Var
type FileVar struct {
	*os.File

	// flag is the flag to open the file with
	// os.O_APPEND|os.O_CREATE|os.O_WRONLY
	flag int
}

// String presents the current value as a string.
func (f *FileVar) String() string {
	if f.File == nil {
		return ""
	}

	return f.Name()
}

// Set is called once, in command line order, for each flag present.
func (f *FileVar) Set(value string) error {
	file, err := os.OpenFile(value, f.flag, 0o644)
	if err != nil {
		return err
	}

	f.File = file
	return nil
}

// Get returns the contents of the Value.
func (f *FileVar) Get() interface{} {
	return f.File
}

// IsBoolFlag returns true if the flag is a boolean flag
func (f *FileVar) IsBoolFlag() bool {
	return false
}

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
	ret := Field[T]{
		FlagName:        flagName,
		FlagDescription: flagDescription,
		EnVarName:       enVarName,
		Value:           value,
	}
	if enVarName != "" {
		ret.FlagDescription += ", EnvVar: " + enVarName
	}

	return ret
}
