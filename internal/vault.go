package internal

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

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

	for _, s := range config.Secrets {
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
	sourceClient, err := NewClient(config.SourceAddr, config.SourceToken)

	if err != nil {
		return err
	}

	source := sourceClient.Logical()

	target := client.Logical()

	for _, s := range config.Secrets {
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
