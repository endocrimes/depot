package config

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestFromEnv(t *testing.T) {
	tt := []struct {
		Name           string
		Env            map[string]string
		ExpectedConfig *Config
		ExpectedErr    error
	}{
		{
			Name: "Missing PROJECT",
			Env: map[string]string{
				"PROJECT":                 "",
				"GCP_SERVICE_ACCOUNT_KEY": "",
				"GCLOUD_POLL_INTERVAL":    "",
				"VM_LIFETIME_DURATION":    "",
				"VM_NAME_PREFIX":          "",
			},
			ExpectedConfig: nil,
			ExpectedErr:    errors.New("PROJECT environment variable not found"),
		},
		{
			Name: "Missing GCP_SERVICE_ACCOUNT_KEY",
			Env: map[string]string{
				"PROJECT":                 "foo",
				"GCP_SERVICE_ACCOUNT_KEY": "",
				"GCLOUD_POLL_INTERVAL":    "",
				"VM_LIFETIME_DURATION":    "",
				"VM_NAME_PREFIX":          "",
			},
			ExpectedConfig: nil,
			ExpectedErr:    errors.New("GCP_SERVICE_ACCOUNT_KEY environment variable not found"),
		},
		{
			Name: "Defaults",
			Env: map[string]string{
				"PROJECT":                 "foo",
				"GCP_SERVICE_ACCOUNT_KEY": "bar",
				"GCLOUD_POLL_INTERVAL":    "",
				"VM_LIFETIME_DURATION":    "",
				"VM_NAME_PREFIX":          "",
			},
			ExpectedConfig: &Config{
				Project:            "foo",
				GCloudPollInterval: 10 * time.Minute,
				VMLifetimeDuration: 24 * time.Hour,
				VMNamePrefix:       "test-cos-beta-",
			},
			ExpectedErr: errors.New("GCP_SERVICE_ACCOUNT_KEY environment variable not found"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			for key, value := range tc.Env {
				setenv(t, key, value)
			}

			cfg, err := FromEnv(hclog.NewNullLogger())
			if tc.ExpectedConfig != nil {
				require.NoError(t, err)
				require.Equal(t, tc.ExpectedConfig, cfg)
			} else {
				require.Equal(t, tc.ExpectedErr, err)
			}
		})
	}
}

func setenv(t *testing.T, key, value string) {
	t.Helper()
	prev, ok := os.LookupEnv(key)

	if value == "" {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("cannot unset environment variable: %v", err)
		}
	} else {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("cannot set environment variable: %v", err)
		}
	}

	if ok {
		t.Cleanup(func() {
			os.Setenv(key, prev)
		})
	} else {
		t.Cleanup(func() {
			os.Unsetenv(key)
		})
	}
}
