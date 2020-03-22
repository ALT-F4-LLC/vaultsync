package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ConfigAuth : auth configuration
type ConfigAuth struct {
	Address     string            `json:"address"`
	Credentials map[string]string `json:"credentials"`
	Method      string            `json:"method"`
}

// ConfigAuthMethod : auth method configuration
type ConfigAuthMethod struct {
	Options map[string]string `json:"options"`
	Path    string            `json:"path"`
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
	SourceAuth         ConfigAuth         `json:"source_auth"`
	SourceSecrets      []ConfigSecret     `json:"source_secrets"`
	SourcePoliciesPath string             `json:"source_policies_path"`
	TargetAuth         ConfigAuth         `json:"target_auth"`
	TargetAuthMethods  []ConfigAuthMethod `json:"target_auth_methods"`
}

// FilenameWithoutExt : removes extension from filename
func FilenameWithoutExt(name string) string {
	return strings.TrimSuffix(name, path.Ext(name))
}

// GetConfig : gets configuration from path
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

// GetPolicies : gets all policies from path
func GetPolicies(path string) ([]os.FileInfo, error) {
	policies := make([]os.FileInfo, 0)

	dir, err := ioutil.ReadDir(path)

	if err != nil {
		return nil, err
	}

	for _, f := range dir {
		name := f.Name()

		file := fmt.Sprintf("%s/%s", path, name)

		extension := filepath.Ext(file)

		if extension == ".hcl" {
			policies = append(policies, f)
		}
	}

	return policies, nil
}
