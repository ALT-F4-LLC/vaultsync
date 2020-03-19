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

	client, err := NewClient(config.TargetAddr, config.TargetToken)

	if err != nil {
		logrus.Fatal(err)
	}

	syncEnginesErr := SyncEngines(client, config)

	if syncEnginesErr != nil {
		logrus.Fatal(syncEnginesErr)
	}

	syncSecretsErr := SyncSecrets(client, config)

	if syncSecretsErr != nil {
		logrus.Fatal(syncSecretsErr)
	}
}
