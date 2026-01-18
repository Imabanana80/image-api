package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port string
	MaxFileSize int64
	APIKeys	map[string]bool
}

var Configuration *Config

func Load() error {
	maxFileSize := int64(getEnvAsInt("MAX_FILE_SIZE_MB", 10)) << 20

	allApiKeys := os.Getenv("API_KEYS")
    if allApiKeys == "" {
        return fmt.Errorf("API_KEYS not set!")
    }
	apiKeys := make(map[string]bool)
    for _, key := range strings.Split(allApiKeys, ",") {
        trimmedKey := strings.TrimSpace(key)
        if trimmedKey != "" {
            apiKeys[trimmedKey] = true
        }
    }
    if len(apiKeys) == 0 {
        return fmt.Errorf("no valid API keys found")
    }

    Configuration = &Config{
        Port:            getEnv("PORT", "8888"),
        MaxFileSize:     maxFileSize,
        APIKeys:         apiKeys,
    }

    return nil
}

func (c *Config) IsValidAPIKey(key string) bool {
    return c.APIKeys[key]
}

func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}

func getEnvAsInt(key string, defaultValue int) int {
    valueStr := os.Getenv(key)
    if valueStr == "" {
        return defaultValue
    }
    value, err := strconv.Atoi(valueStr)
    if err != nil {
        log.Printf("Invalid integer for %s, using default: %d", 
            key, defaultValue)
        return defaultValue
    }
    return value
}
