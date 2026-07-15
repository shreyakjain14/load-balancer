package config

import (
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type Config struct {
    Server   ServerConfig    `yaml:"server"`
    Health   HealthConfig    `yaml:"health"`
    Backends []BackendConfig `yaml:"backends"`
}

type ServerConfig struct {
    Port         string        `yaml:"port"`
    ReadTimeout  time.Duration `yaml:"readTimeout"`
    WriteTimeout time.Duration `yaml:"writeTimeout"`
}

type BackendConfig struct {
    URL string `yaml:"url"`
}

type HealthConfig struct {
    Interval time.Duration `yaml:"interval"`
    Timeout  time.Duration `yaml:"timeout"`
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