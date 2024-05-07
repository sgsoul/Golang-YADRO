package config

import (
	"os"
	"gopkg.in/yaml.v2"
	log "github.com/rs/zerolog/log"
)

type Config struct {
	SourceURL string `yaml:"source_url"`
	DBFile    string `yaml:"db_file"`
	Parallel  int    `yaml:"parallel"`
	IndexFile string `yaml:"index_file"`
	Port      int    `yaml:"port"`
}

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Error().Err(err).Msg("error reading configuration fail")
		return nil, err
	}

	config := &Config{}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		log.Error().Err(err).Msg("error unmarshal configuration fail")
		return nil, err
	}

	return config, err
}

