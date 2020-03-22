package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

// AppRole : defines app role for authentication
type AppRole struct {
	RoleID   string `json:"role_id"`
	SecretID string `json:"secret_id"`
}

// LoginAuthResult : defines vault successful login
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
func Login(auth ConfigAuth) (*string, error) {
	var token string

	if auth.Method == "approle" {
		approle := AppRole{
			RoleID:   auth.Credentials["role_id"],
			SecretID: auth.Credentials["secret_id"],
		}

		data, err := json.Marshal(approle)

		if err != nil {
			return nil, err
		}

		url := fmt.Sprintf("%s/v1/auth/approle/login", auth.Address)

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))

		if err != nil {
			return nil, err
		}

		var result LoginResult

		json.NewDecoder(resp.Body).Decode(&result)

		token = result.Auth.ClientToken
	}

	if auth.Method == "token" {
		token = auth.Credentials["token"]
	}

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

// SyncPolicies : syncs policies to client Vault service
func SyncPolicies(client *api.Client, config *Config) error {
	policies, err := GetPolicies(config.SourcePoliciesPath)

	if err != nil {
		return err
	}

	sys := client.Sys()

	for _, p := range policies {
		filename := p.Name()

		path := fmt.Sprintf("%s/%s", config.SourcePoliciesPath, filename)

		policy, err := ioutil.ReadFile(path)

		if err != nil {
			return err
		}

		name := FilenameWithoutExt(filename)

		putErr := sys.PutPolicy(name, string(policy))

		if putErr != nil {
			return err
		}
	}

	return nil
}

// SyncSecrets : sync secrets from source to target
func SyncSecrets(client *api.Client, config *Config) error {
	sourceToken, err := Login(config.SourceAuth)

	if err != nil {
		return err
	}

	sourceClient, err := NewClient(config.SourceAuth.Address, *sourceToken)

	if err != nil {
		return err
	}

	sourceClient.SetToken(*sourceToken)

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
