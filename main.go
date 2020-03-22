package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func main() {
	args, err := GetArgs()

	if err != nil {
		logrus.Fatal(err)
	}

	config, err := GetConfig(args.ConfigPath)

	if err != nil {
		logrus.Fatal("invalid_config")
	}

	targetToken, err := Login(config.TargetAuth)

	if err != nil {
		logrus.Fatal(err)
	}

	targetClient, err := NewClient(config.TargetAuth.Address, *targetToken)

	if err != nil {
		logrus.Fatal(err)
	}

	syncEnginesErr := SyncEngines(targetClient, config)

	if syncEnginesErr != nil {
		logrus.Fatal(syncEnginesErr)
	}

	syncSecretsErr := SyncSecrets(targetClient, config)

	if syncSecretsErr != nil {
		logrus.Fatal(syncSecretsErr)
	}

	syncPoliciesErr := SyncPolicies(targetClient, config)

	if syncEnginesErr != nil {
		logrus.Fatal(syncPoliciesErr)
	}

	syncAuthMethodsErr := SyncAuthMethods(targetClient, config)

	if syncAuthMethodsErr != nil {
		logrus.Fatal(syncAuthMethodsErr)
	}

	logical := targetClient.Logical()

	for _, r := range config.TargetAuthAppRoles {
		rolePath := fmt.Sprintf("auth/%s/role/%s", r.Path, r.Name)

		exists, err := logical.Read(rolePath)

		if err != nil {
			logrus.Fatal(err)
		}

		if exists == nil {
			_, err := logical.Write(rolePath, r.Options)

			if err != nil {
				logrus.Fatal(err)
			}

			var secretIDData map[string]interface{}

			secretIDPath := fmt.Sprintf("%s/secret-id", rolePath)

			secretIDSecret, secretIDErr := logical.Write(secretIDPath, secretIDData)

			if secretIDErr != nil {
				logrus.Fatal(secretIDErr)
			}

			roleIDPath := fmt.Sprintf("%s/role-id", rolePath)

			roleIDSecret, err := logical.Read(roleIDPath)

			if err != nil {
				logrus.Fatal(err)
			}

			if config.TargetAuthAppRolesOutput != nil {
				outputPath, err := filepath.Abs(*config.TargetAuthAppRolesOutput)

				if err != nil {
					logrus.Fatal(err)
				}

				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					err := os.MkdirAll(outputPath, 0755)

					if err != nil {
						logrus.Fatal(err)
					}
				}

				roleID := roleIDSecret.Data["role_id"]

				secretID := secretIDSecret.Data["secret_id"]

				approle := map[string]interface{}{
					"role_id":   roleID,
					"secret_id": secretID,
				}

				json, err := json.Marshal(approle)

				if err != nil {
					logrus.Fatal(err)
				}

				writePath := fmt.Sprintf("%s/%s.json", outputPath, r.Name)

				writeErr := ioutil.WriteFile(writePath, json, 0644)

				if writeErr != nil {
					logrus.Fatal(writeErr)
				}

				fields := logrus.Fields{"role_id": roleID, "role_name": r.Name}

				logrus.WithFields(fields).Info("Created AppRole in Vault")
			}
		}
	}
}
