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
	DownloadServiceURL         string        `envconfig:"DOWNLOAD_SERVICE_URL"`
	EnableDatasetPreview       bool          `envconfig:"ENABLE_DATASET_PREVIEW"`
	EnableProfiler             bool          `envconfig:"ENABLE_PROFILER"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	MaxDatasetOptions          int           `envconfig:"MAX_DATASET_OPTIONS"`
	PprofToken                 string        `envconfig:"PPROF_TOKEN" json:"-"`
	RendererURL                string        `envconfig:"RENDERER_URL"`
	SearchAPIAuthToken         string        `envconfig:"SEARCH_API_AUTH_TOKEN"  json:"-"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (cfg *Config, err error) {

	cfg = &Config{
		APIRouterURL:               "http://localhost:23200/v1",
		BatchMaxWorkers:            100,
		BatchSizeLimit:             1000,
		BindAddr:                   ":20001",
		DownloadServiceURL:         "http://localhost:23600",
		EnableDatasetPreview:       false,
		EnableProfiler:             false,
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		MaxDatasetOptions:          200,
		RendererURL:                "http://localhost:20010",
	}

	return cfg, envconfig.Process("", cfg)
}
