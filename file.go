package main

import (
	"encoding/json"
	"io/ioutil"
)

// ConfigSecret :
type ConfigSecret struct {
	Engine  string            `json:"engine"`
	Mount   string            `json:"mount"`
	Options map[string]string `json:"options"`
	Paths   []string          `json:"paths"`
}

// Config : configuration for vaultsync
type Config struct {
	Secrets     []ConfigSecret `json:"vault_secrets"`
	SourceAddr  string         `json:"vault_source_addr"`
	SourceToken string         `json:"vault_source_token"`
	TargetAddr  string         `json:"vault_target_addr"`
	TargetToken string         `json:"vault_target_token"`
}

// GetConfig : gets configuration
func GetConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var config Config

	if len(data) > 0 {
		jsonErr := json.Unmarshal(data, &config)

		if jsonErr != nil {
			return nil, jsonErr
		}
	}

	return &config, nil
}
