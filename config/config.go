package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-frontend-filter-dataset-controller
type Config struct {
	APIRouterURL               string        `envconfig:"API_ROUTER_URL"`
	BatchMaxWorkers            int           `envconfig:"BATCH_MAX_WORKERS"`
	BatchSizeLimit             int           `envconfig:"BATCH_SIZE_LIMIT"`
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	Debug                      bool          `envconfig:"DEBUG"`
	DownloadServiceURL         string        `envconfig:"DOWNLOAD_SERVICE_URL"`
	EnableDatasetPreview       bool          `envconfig:"ENABLE_DATASET_PREVIEW"`
	EnableProfiler             bool          `envconfig:"ENABLE_PROFILER"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	MaxDatasetOptions          int           `envconfig:"MAX_DATASET_OPTIONS"`
	PatternLibraryAssetsPath   string        `envconfig:"PATTERN_LIBRARY_ASSETS_PATH"`
	PprofToken                 string        `envconfig:"PPROF_TOKEN" json:"-"`
	SearchAPIAuthToken         string        `envconfig:"SEARCH_API_AUTH_TOKEN"  json:"-"`
	SiteDomain                 string        `envconfig:"SITE_DOMAIN"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	cfg, err := get()
	if err != nil {
		return nil, err
	}

	if cfg.Debug {
		cfg.PatternLibraryAssetsPath = "http://localhost:9002/dist/assets"
	} else {
		cfg.PatternLibraryAssetsPath = "//cdn.ons.gov.uk/dp-design-system/afa6add"
	}

	return cfg, nil
}

func get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		APIRouterURL:               "http://localhost:23200/v1",
		BatchMaxWorkers:            100,
		BatchSizeLimit:             1000,
		BindAddr:                   "localhost:20001",
		Debug:                      false,
		DownloadServiceURL:         "http://localhost:23600",
		EnableDatasetPreview:       false,
		EnableProfiler:             false,
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		MaxDatasetOptions:          200,
		SiteDomain:                 "localhost",
	}

	return cfg, envconfig.Process("", cfg)
}
