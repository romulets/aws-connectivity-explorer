package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log/slog"
	"os"
)

type Config struct {
	Aws   AwsConfig   `yaml:"aws"`
	Neo4j Neo4jConfig `neo4j`
}

type AwsConfig struct {
	Region string `yaml:"region"`
}

type Neo4jConfig struct {
	Uri      string
	Username string
	Password string
}

func LoadConfig(logger *slog.Logger) (*Config, error) {
	logger.Info("Loading config.yml")

	var cfg Config
	f, err := os.Open("config.yml")
	if err != nil {
		return nil, fmt.Errorf("cannot open config.yml: %w", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logger.Info("Could not close config.yml")
		}
	}(f)

	decoder := yaml.NewDecoder(f)
	if err = decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("cannot decode config.yml: %w", err)
	}

	return &cfg, nil
}
