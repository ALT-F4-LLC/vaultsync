package main

import (
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFlag = kingpin.Flag("config", "Sets path for Vault sync utility.").Default("./config.json").String()
)

// Args : command line arguments
type Args struct {
	ConfigPath string
}

// GetArgs : gets command line arguments
func GetArgs() (*Args, error) {
	kingpin.Parse()

	configPath, err := filepath.Abs(*configFlag)

	if err != nil {
		return nil, err
	}

	args := &Args{
		ConfigPath: configPath,
	}

	return args, nil
}
