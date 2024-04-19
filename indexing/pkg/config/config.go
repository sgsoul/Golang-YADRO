package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	SourceURL string `yaml:"source_url"`
	DBFile    string `yaml:"db_file"`
	Parallel  int    `yaml:"parallel"`
	IndexFile string `yaml:"index_file"`
}

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration fail: %v", err)
	}

	config := &Config{}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshal configuration fail: %v", err)
	}

	return config, nil
}
