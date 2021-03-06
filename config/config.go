package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-frontend-filter-dataset-controller
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	RendererURL                string        `envconfig:"RENDERER_URL"`
	APIRouterURL               string        `envconfig:"API_ROUTER_URL"`
	DatasetAPIAuthToken        string        `envconfig:"DATASET_API_AUTH_TOKEN" json:"-"`
	FilterAPIAuthToken         string        `envconfig:"FILTER_API_AUTH_TOKEN"  json:"-"`
	SearchAPIAuthToken         string        `envconfig:"SEARCH_API_AUTH_TOKEN"  json:"-"`
	DownloadServiceURL         string        `envconfig:"DOWNLOAD_SERVICE_URL"`
	EnableDatasetPreview       bool          `envconfig:"ENABLE_DATASET_PREVIEW"`
	EnableProfiler             bool          `envconfig:"ENABLE_PROFILER"`
	PprofToken                 string        `envconfig:"PPROF_TOKEN" json:"-"`
	BatchSizeLimit             int           `envconfig:"BATCH_SIZE_LIMIT"`
	BatchMaxWorkers            int           `envconfig:"BATCH_MAX_WORKERS"`
	MaxDatasetOptions          int           `envconfig:"MAX_DATASET_OPTIONS"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (cfg *Config, err error) {

	cfg = &Config{
		BindAddr:                   ":20001",
		RendererURL:                "http://localhost:20010",
		APIRouterURL:               "http://localhost:23200/v1",
		DatasetAPIAuthToken:        "FD0108EA-825D-411C-9B1D-41EF7727F465",
		FilterAPIAuthToken:         "FD0108EA-825D-411C-9B1D-41EF7727F465",
		SearchAPIAuthToken:         "SD0108EA-825D-411C-45J3-41EF7727F123",
		DownloadServiceURL:         "http://localhost:23600",
		EnableDatasetPreview:       false,
		EnableProfiler:             false,
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		BatchSizeLimit:             1000, // maximum limit value to get items from APIs in a single call
		BatchMaxWorkers:            100,  // maximum number of concurrent go-routines requesting items concurrently from APIs with pagination
		MaxDatasetOptions:          200,  // maximum number of IDs that will be requested to dataset API in a single call as query parmeters
	}

	return cfg, envconfig.Process("", cfg)
}
