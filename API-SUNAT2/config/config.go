package config

import (
	"os"
)

type Config struct {
	Port         string `json:"port"`
	XMLStorePath string `json:"xmlStorePath"`
	LogLevel     string `json:"logLevel"`
}

func LoadConfig() *Config {
	return &Config{
		Port:         getEnvOrDefault("PORT", "8080"),
		XMLStorePath: getEnvOrDefault("XML_STORE_PATH", "./xml_output"),
		LogLevel:     getEnvOrDefault("LOG_LEVEL", "info"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 