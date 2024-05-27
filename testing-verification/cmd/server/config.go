package main

import (
	"os"

	"github.com/sgsoul/internal/core"
	log "github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

func LoadConfig(configPath string) (*core.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Error().Err(err).Msg("error reading configuration fail")
		return nil, err
	}

	config := &core.Config{}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		log.Error().Err(err).Msg("error unmarshal configuration fail")
		return nil, err
	}

	return config, err
}

func New(configPath string) *core.Config {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		log.Error().Err(err).Msg("error reading configuration fail")
		return nil
	}
	return cfg
}
