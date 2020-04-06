package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-frontend-filter-dataset-controller
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	RendererURL                string        `envconfig:"RENDERER_URL"`
	FilterAPIURL               string        `envconfig:"FILTER_API_URL"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	HierarchyAPIURL            string        `envconfig:"HIERARCHY_API_URL"`
	DatasetAPIAuthToken        string        `envconfig:"DATASET_API_AUTH_TOKEN" json:"-"`
	FilterAPIAuthToken         string        `envconfig:"FILTER_API_AUTH_TOKEN"  json:"-"`
	SearchAPIAuthToken         string        `envconfig:"SEARCH_API_AUTH_TOKEN"  json:"-"`
	SearchAPIURL               string        `envconfig:"SEARCH_API_URL"`
	DownloadServiceURL         string        `envconfig:"DOWNLOAD_SERVICE_URL"`
	EnableDatasetPreview       bool          `envconfig:"ENABLE_DATASET_PREVIEW"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	BatchSizeFilterAPI         int           `envconfig:"BATCH_SIZE_FILTER_API"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (cfg *Config, err error) {

	cfg = &Config{
		BindAddr:                   ":20001",
		RendererURL:                "http://localhost:20010",
		FilterAPIURL:               "http://localhost:22100",
		DatasetAPIURL:              "http://localhost:22000",
		HierarchyAPIURL:            "http://localhost:22600",
		DatasetAPIAuthToken:        "FD0108EA-825D-411C-9B1D-41EF7727F465",
		FilterAPIAuthToken:         "FD0108EA-825D-411C-9B1D-41EF7727F465",
		SearchAPIAuthToken:         "SD0108EA-825D-411C-45J3-41EF7727F123",
		SearchAPIURL:               "http://localhost:23100",
		DownloadServiceURL:         "http://localhost:23600",
		EnableDatasetPreview:       false,
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        10 * time.Second,
		HealthCheckCriticalTimeout: time.Minute,
		BatchSizeFilterAPI:         100,
	}

	return cfg, envconfig.Process("", cfg)
}
