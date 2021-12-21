package aconfig

import (
	"os"
	"strconv"
)

func GetEnvOrInt(name string, defaultValue int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil {
		return defaultValue
	}

	return value
}
