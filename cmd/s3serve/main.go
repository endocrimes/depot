package main

import (
	"flag"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/IGLOU-EU/go-wildcard/v2"
	"github.com/endocrimes/depot/pkg/s3fileserver"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/s3utils"
	"github.com/rs/cors"
)

var (
	endpoint          string
	accessKey         string
	accessKeyFile     string
	secretKey         string
	secretKeyFile     string
	address           string
	bucket            string
	bucketPath        string
	allowedCorsOrigin string
)

func init() {
	flag.StringVar(&endpoint, "endpoint", "", "AWS S3 compatible endpoint")
	flag.StringVar(&bucket, "bucket", "", "bucket name to serve from")
	flag.StringVar(&bucketPath, "bucketPath", "/", "path within the bucket to treat as the serving root")
	flag.StringVar(&address, "address", "0.0.0.0:8080", "bind to a specific ADDRESS:PORT, ADDRESS can be an IP or hostname")
	flag.StringVar(&allowedCorsOrigin, "allowed-cors-origins", "", "a list of origins a cross-domain request can be executed from")
	flag.StringVar(&accessKey, "accessKey", "", "access key for server")
	flag.StringVar(&secretKey, "secretKey", "", "secret key for server")
	flag.StringVar(&accessKeyFile, "accessKeyFile", "", "file that contains the access key")
	flag.StringVar(&secretKeyFile, "secretKeyFile", "", "file that contains the secret key")
}

func newHTTPTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          256,
		MaxIdleConnsPerHost:   256,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	}
}

func main() {
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if strings.TrimSpace(endpoint) == "" {
		logger.Error("endpoint cannot be empty")
		os.Exit(2)
	}
	if strings.TrimSpace(bucket) == "" {
		logger.Error("bucket name cannot be empty")
		os.Exit(2)
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		logger.Error("failed to parse endpoint", "error", err)
		os.Exit(2)
	}

	credentialProviders := []credentials.Provider{
		&credentials.EnvAWS{},
		&credentials.FileAWSCredentials{},
		&credentials.EnvMinio{},
	}
	if accessKeyFile != "" {
		if keyBytes, err := os.ReadFile(accessKeyFile); err == nil {
			accessKey = strings.TrimSpace(string(keyBytes))
		} else {
			logger.Error("Failed to read access key file", "file", accessKeyFile)
			os.Exit(2)
		}
	}
	if secretKeyFile != "" {
		if keyBytes, err := os.ReadFile(secretKeyFile); err == nil {
			secretKey = strings.TrimSpace(string(keyBytes))
		} else {
			logger.Error("Failed to read secret key file", "file", secretKeyFile)
			os.Exit(2)
		}
	}
	if accessKey != "" && secretKey != "" {
		credentialProviders = []credentials.Provider{
			&credentials.Static{
				Value: credentials.Value{
					AccessKeyID:     accessKey,
					SecretAccessKey: secretKey,
				},
			},
		}
	}

	creds := credentials.NewChainCredentials(credentialProviders)

	client, err := minio.New(u.Host, &minio.Options{
		Creds:        creds,
		Secure:       u.Scheme == "https",
		Region:       s3utils.GetRegionFromURL(*u),
		BucketLookup: minio.BucketLookupAuto,
		Transport:    newHTTPTransport(),
	})
	if err != nil {
		logger.Error("failed to setup s3 client", "error", err)
		os.Exit(2)
	}

	mux := http.FileServer(&s3fileserver.Impl{
		Client:     client,
		Bucket:     bucket,
		BucketPath: bucketPath,
		Logger:     logger,
	})

	// Wrap the existing mux with the CORS middleware.
	opts := cors.Options{
		AllowOriginFunc: func(origin string) bool {
			if allowedCorsOrigin == "" {
				return true
			}
			for _, allowedOrigin := range strings.Split(allowedCorsOrigin, ",") {
				if wildcard.Match(allowedOrigin, origin) {
					return true
				}
			}
			return false
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPut,
			http.MethodHead,
			http.MethodPost,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodPatch,
		},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
	}

	handler := cors.New(opts).Handler(mux)

	logger.Info("Started listening", "addr", address)
	server := &http.Server{
		Addr:         address,
		Handler:      handler,
		ReadTimeout:  55 * time.Second,
		WriteTimeout: 55 * time.Second,
	}
	err = server.ListenAndServe()
	if err != nil {
		logger.Error("error serving content", "error", err)
		os.Exit(1)
	}
}
