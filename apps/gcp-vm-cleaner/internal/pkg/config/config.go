package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
)

// Config holds the service configuration.
type Config struct {
	Project            string
	GCloudPollInterval time.Duration
	VMLifetimeDuration time.Duration
	VMNamePrefix       string
}

// FromEnv returns a new configuration loaded from the environment or the first
// encountered error.
// It will mutate the process environment after writing out the Google Service
// Account Key to a file for use by the GCP Client.
func FromEnv(logger hclog.Logger) (*Config, error) {
	project, ok := os.LookupEnv("PROJECT")
	if !ok {
		return &Config{}, errors.New("PROJECT environment variable not found")
	}
	logger.Info("Using GCP Project", "id", project)

	gcpServiceAccountKey, ok := os.LookupEnv("GCP_SERVICE_ACCOUNT_KEY")
	if !ok {
		return &Config{}, errors.New("GCP_SERVICE_ACCOUNT_KEY environment variable not found")
	}
	gcpKeyFile, err := ioutil.TempFile("", "gcp-service-account-key")
	if err != nil {
		return &Config{}, fmt.Errorf("failed to create temporary file for gcp service account key: %w", err)
	}
	defer gcpKeyFile.Close()

	_, err = gcpKeyFile.Write([]byte(gcpServiceAccountKey))
	if err != nil {
		return &Config{}, fmt.Errorf("failed to write to temporary file for gcp service account key: %w", err)
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", gcpKeyFile.Name())

	gcloudPollIntervalStr, ok := os.LookupEnv("GCLOUD_POLL_INTERVAL")
	if !ok {
		gcloudPollIntervalStr = "10m"
	}
	logger.Info("Using Poll Interval", "interval", gcloudPollIntervalStr)

	gcloudPollInterval, err := time.ParseDuration(gcloudPollIntervalStr)
	if err != nil {
		return &Config{}, fmt.Errorf("failed to parse GCLOUD_POLL_INTERVAL environment variable: %w", err)
	}

	vmLifetimeDurationStr, ok := os.LookupEnv("VM_LIFETIME_DURATION")
	if !ok {
		vmLifetimeDurationStr = "24h"
	}
	logger.Info("Using Maximum Lifetime", "duration", vmLifetimeDurationStr)

	vmLifetimeDuration, err := time.ParseDuration(vmLifetimeDurationStr)
	if err != nil {
		return &Config{}, fmt.Errorf("failed to parse VM_LIFETIME_DURATION environment variable: %w", err)
	}

	vmNamePrefix, ok := os.LookupEnv("VM_NAME_PREFIX")
	if !ok {
		vmNamePrefix = "test-cos-beta-"
	}
	logger.Info("Using vm prefix", "prefix", vmNamePrefix)

	return &Config{
		Project:            project,
		GCloudPollInterval: gcloudPollInterval,
		VMLifetimeDuration: vmLifetimeDuration,
		VMNamePrefix:       vmNamePrefix,
	}, nil
}
