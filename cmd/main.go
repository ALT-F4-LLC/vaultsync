package main

import (
	"github.com/erkrnt/vaultsync/internal"
	"github.com/sirupsen/logrus"
)

func main() {
	args, err := internal.GetArgs()

	if err != nil {
		logrus.Fatal(err)
	}

	config, err := internal.GetConfig(args.ConfigPath)

	if err != nil {
		logrus.Fatal("invalid_config")
	}

	client, err := internal.NewClient(config.TargetAddr, config.TargetToken)

	if err != nil {
		logrus.Fatal(err)
	}

	syncEnginesErr := internal.SyncEngines(client, config)

	if syncEnginesErr != nil {
		logrus.Fatal(syncEnginesErr)
	}

	syncSecretsErr := internal.SyncSecrets(client, config)

	if syncSecretsErr != nil {
		logrus.Fatal(syncSecretsErr)
	}
}
