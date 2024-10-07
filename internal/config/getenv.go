package config

import (
	"os"
	"strconv"
	"time"
)

// GetEnv retrieves the value of an environment variable or returns a default value
func GetEnv[T any](key string, defaultValue T) T {
	if value, exists := os.LookupEnv(key); exists {
		switch any(defaultValue).(type) {
		case string:
			return any(value).(T)
		case int:
			if intValue, err := strconv.Atoi(value); err == nil {
				return any(intValue).(T)
			}
		case time.Duration:
			if durationValue, err := time.ParseDuration(value); err == nil {
				return any(durationValue).(T)
			}
		case bool:
			if boolValue, err := strconv.ParseBool(value); err == nil {
				return any(boolValue).(T)
			}

		default:
			return defaultValue
		}
	}

	return defaultValue
}
