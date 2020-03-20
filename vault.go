package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

type AppRole struct {
	RoleID   string `json:"role_id"`
	SecretID string `json:"secret_id"`
}

type LoginAuthResult struct {
	Accessor      string            `json:"accessor"`
	ClientToken   string            `json:"client_token"`
	LeaseDuration int               `json:"lease_duration"`
	Metadata      map[string]string `json:"metadata"`
	Policies      []string          `json:"policies"`
	Renewable     bool              `json:"renewable"`
}

type LoginResult struct {
	Auth LoginAuthResult `json:"auth"`
}

// Login : login to Vault
func Login(config *Config) (*string, error) {
	auth := AppRole{
		RoleID:   config.SourceAuth.Credentials["role_id"],
		SecretID: config.SourceAuth.Credentials["secret_id"],
	}

	data, err := json.Marshal(auth)

	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/auth/approle/login", config.SourceAddr)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))

	if err != nil {
		return nil, err
	}

	var result LoginResult

	json.NewDecoder(resp.Body).Decode(&result)

	token := result.Auth.ClientToken

	return &token, nil
}

// NewClient : get Vault client
func NewClient(address string, token string) (*api.Client, error) {
	config := &api.Config{
		Address: address,
	}

	client, err := api.NewClient(config)

	if err != nil {
		return nil, err
	}

	client.SetToken(token)

	return client, nil
}

// GetEngines : get Vault system engines
func GetEngines(client *api.Client) (map[string]*api.MountOutput, error) {
	sys := client.Sys()

	mounts, err := sys.ListMounts()

	if err != nil {
		return nil, err
	}

	return mounts, nil
}

// MountEngine : verifies engine exists and is properly configured
func MountEngine(client *api.Client, secret ConfigSecret) error {
	sys := client.Sys()

	input := &api.MountInput{
		Options: secret.Options,
		Type:    secret.Engine,
	}

	err := sys.Mount(secret.Mount, input)

	if err != nil {
		return err
	}

	fields := logrus.Fields{"mount": secret.Mount}

	logrus.WithFields(fields).Info("Created engine")

	return nil
}

// SyncEngines : syncs secret engines from config
func SyncEngines(client *api.Client, config *Config) error {
	engines, err := GetEngines(client)

	if err != nil {
		return err
	}

	for _, s := range config.SourceSecrets {
		if engines[s.Mount] == nil {
			err := MountEngine(client, s)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// SyncSecrets : sync secrets from source to target
func SyncSecrets(client *api.Client, config *Config) error {
	token, err := Login(config)

	if err != nil {
		return err
	}

	sourceClient, err := NewClient(config.SourceAddr, config.SourceToken)

	if err != nil {
		return err
	}

	sourceClient.SetToken(*token)

	source := sourceClient.Logical()

	target := client.Logical()

	for _, s := range config.SourceSecrets {
		for _, p := range s.Paths {
			path := fmt.Sprintf("%s/%s", s.Mount, p)

			secret, err := source.Read(path)

			if err != nil {
				return err
			}

			_, writeErr := target.Write(path, secret.Data)

			if writeErr != nil {
				return writeErr
			}
		}
	}

	return nil
}
