package main

import (
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

	syncAppRolesErr := SyncAppRoles(targetClient, config)

	if syncAppRolesErr != nil {
		logrus.Fatal(syncAppRolesErr)
	}
}
