package main

import (
	"encoding/json"
	"io/ioutil"
)

// ConfigAuth : auth configuration
type ConfigAuth struct {
	Credentials map[string]string `json:"credentials"`
	Method      string            `json:"method"`
}

// ConfigSecret : secrets configuration
type ConfigSecret struct {
	Engine    string            `json:"engine"`
	Mount     string            `json:"mount"`
	Options   map[string]string `json:"options"`
	Overwrite bool              `json:"overwrite"`
	Paths     []string          `json:"paths"`
}

// Config : configuration for vaultsync
type Config struct {
	SourceAddr    string         `json:"vault_source_addr"`
	SourceAuth    ConfigAuth     `json:"vault_source_auth"`
	SourceSecrets []ConfigSecret `json:"vault_source_secrets"`
	SourceToken   string         `json:"vault_source_token"`
	TargetAddr    string         `json:"vault_target_addr"`
	TargetToken   string         `json:"vault_target_token"`
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
