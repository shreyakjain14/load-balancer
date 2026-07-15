package config

import (
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Server ServerConfig
	Health HealthConfig
	Backends []BackendConfig
}

type ServerConfig struct {
	Port string
	ReadTimeout time.Duration
	WriteTimeout time.Duration
}

type HealthConfig struct {
	Duration time.Duration
	Interval time.Duration
}

type BackendConfig struct {
	URL string	
}

func GetConfig(path string) (Config, error) {
	config := Config{}

	data, err := os.ReadFile(path)

	if err != nil {
		return Config{}, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
    	return Config{}, err	
	}

	return config, nil
}