package env

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// GetEnvString retrieves the value of the environment variable named by the key.
// If the variable is present in the environment the value (which may be empty)
// is returned as string, otherwise the defaultValue is returned.
func GetEnvString(key, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return value
}

// GetEnvStringList retrieves the value of the environment variable named by the key.
// If the variable is present in the environment the value (which may be empty)
// is split by ',' or ':' and returned as string List, otherwise the defaultValue
// is returned.
func GetEnvStringList(key string, defaultValue []string) []string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return regexp.MustCompile(`\s*[,:]\s*`).Split(value, -1)
}

// GetEnvInt64 retrieves the value of the environment variable named by the key.
// If the variable is present in the environment the value (which may be empty)
// is returned as int64, otherwise the defaultValue is returned.
func GetEnvInt64(key string, defaultValue int64) int64 {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	if intValue, err := strconv.ParseInt(value, 0, 64); err == nil {
		return intValue
	}
	return defaultValue
}

// GetEnvInt64 retrieves the value of the environment variable named by the key.
// If the variable is present in the environment the value (which may be empty)
// is returned as bool, otherwise the defaultValue is returned.
// Accepted true value are in ["1", "t", "y", "true", "yes"] in any case.
// Accepted false value are in ["0", "f", "n", "false", "no"] in any case.
func GetEnvBool(key string, defaultValue bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	value = strings.ToLower(value)
	switch value {
	case "1", "t", "y", "true", "yes":
		return true
	case "0", "f", "n", "false", "no":
		return false
	default:
		return defaultValue
	}
}
