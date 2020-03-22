package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

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

// SyncAppRoles : syncs approles from configuration to Vault service
func SyncAppRoles(client *api.Client, config *Config) error {
	logical := client.Logical()

	for _, r := range config.TargetAuthAppRoles {
		rolePath := fmt.Sprintf("auth/%s/role/%s", r.Path, r.Name)

		exists, err := logical.Read(rolePath)

		if err != nil {
			return err
		}

		if exists == nil {
			_, err := logical.Write(rolePath, r.Options)

			if err != nil {
				return err
			}

			var secretIDData map[string]interface{}

			secretIDPath := fmt.Sprintf("%s/secret-id", rolePath)

			secretIDSecret, err := logical.Write(secretIDPath, secretIDData)

			if err != nil {
				return err
			}

			roleIDPath := fmt.Sprintf("%s/role-id", rolePath)

			roleIDSecret, err := logical.Read(roleIDPath)

			if err != nil {
				return err
			}

			roleID := roleIDSecret.Data["role_id"]

			secretID := secretIDSecret.Data["secret_id"]

			fields := logrus.Fields{"role_id": roleID, "role_name": r.Name}

			logrus.WithFields(fields).Info("Created AppRole in Vault")

			if r.Output != nil {
				outputPath, err := filepath.Abs(*r.Output)

				if err != nil {
					return err
				}

				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					err := os.MkdirAll(outputPath, 0755)

					if err != nil {
						return err
					}
				}

				approle := map[string]interface{}{
					"role_id":   roleID,
					"secret_id": secretID,
				}

				json, err := json.MarshalIndent(approle, "", "\t")

				if err != nil {
					return err
				}

				filename := fmt.Sprintf("%s.json", r.Name)

				writePath := fmt.Sprintf("%s/%s", outputPath, filename)

				writeErr := ioutil.WriteFile(writePath, json, 0644)

				if writeErr != nil {
					return writeErr
				}

				fields := logrus.Fields{"filename": filename}

				logrus.WithFields(fields).Info("Created AppRole in folder")
			}
		}
	}

	return nil
}

// SyncAuthMethods : syncs auth methods from config to client Vault service
func SyncAuthMethods(client *api.Client, config *Config) error {
	sys := client.Sys()

	mounts, err := sys.ListAuth()

	if err != nil {
		return err
	}

	for _, m := range config.TargetAuthMethods {
		path := fmt.Sprintf("%s/", m.Path)

		if mounts[path] == nil {
			input := &api.MountInput{
				Type: m.Options["type"],
			}

			enableErr := sys.EnableAuthWithOptions(m.Path, input)

			if enableErr != nil {
				return enableErr
			}
		}
	}

	return nil
}

// SyncEngines : syncs secret engines from config
func SyncEngines(client *api.Client, config *Config) error {
	engines, err := GetEngines(client)

	if err != nil {
		return err
	}

	for _, s := range config.SourceSecrets {
		path := fmt.Sprintf("%s/", s.Mount)

		if engines[path] == nil {
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
			return putErr
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
