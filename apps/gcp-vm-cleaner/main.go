package main

import (
	"os"

	"github.com/endocrimes/depot/apps/gcp-vm-cleaner/internal/pkg/config"
	"github.com/hashicorp/go-hclog"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "gcp-vm-cleaner",
		Level:      hclog.LevelFromString("DEBUG"),
		JSONFormat: true,
	})

	logger.Info("Starting up")

	_, err := config.FromEnv(logger.Named("config"))
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(2)
	}
}
