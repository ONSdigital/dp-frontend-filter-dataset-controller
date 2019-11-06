package config

import (
	"os"
	"strconv"

	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
)

// Config represents service configuration for dp-frontend-filter-dataset-controller
type Config struct {
	BindAddr             string
	RendererURL          string
	FilterAPIURL         string
	DatasetAPIURL        string
	HierarchyAPIURL      string
	DatasetAPIAuthToken  string
	FilterAPIAuthToken   string
	SearchAPIAuthToken   string
	SearchAPIURL         string
	DownloadServiceURL   string
	EnableDatasetPreview bool
}

// Get returns the default config with any modifications through environment
// variables
func Get() *Config {
	cfg := &Config{
		BindAddr:             ":20001",
		RendererURL:          "http://localhost:20010",
		FilterAPIURL:         "http://localhost:22100",
		DatasetAPIURL:        "http://localhost:22000",
		HierarchyAPIURL:      "http://localhost:22600",
		DatasetAPIAuthToken:  "FD0108EA-825D-411C-9B1D-41EF7727F465",
		FilterAPIAuthToken:   "FD0108EA-825D-411C-9B1D-41EF7727F465",
		SearchAPIAuthToken:   "SD0108EA-825D-411C-45J3-41EF7727F123",
		SearchAPIURL:         "http://localhost:23100",
		DownloadServiceURL:   "http://localhost:23600",
		EnableDatasetPreview: false,
	}

	if v := os.Getenv("BIND_ADDR"); len(v) > 0 {
		cfg.BindAddr = v
	}
	if v := os.Getenv("RENDERER_URL"); len(v) > 0 {
		cfg.RendererURL = v
	}
	if v := os.Getenv("FILTER_API_URL"); len(v) > 0 {
		cfg.FilterAPIURL = v
	}
	if v := os.Getenv("DATASET_API_URL"); len(v) > 0 {
		cfg.DatasetAPIURL = v
	}
	if v := os.Getenv("HIERARCHY_API_URL"); len(v) > 0 {
		cfg.HierarchyAPIURL = v
	}
	if v := os.Getenv("SEARCH_API_URL"); len(v) > 0 {
		cfg.SearchAPIURL = v
	}
	if v := os.Getenv("DATASET_API_AUTH_TOKEN"); len(v) > 0 {
		cfg.DatasetAPIAuthToken = v
	}
	if v := os.Getenv("SEARCH_API_AUTH_TOKEN"); len(v) > 0 {
		cfg.SearchAPIAuthToken = v
	}
	if v := os.Getenv("FILTER_API_AUTH_TOKEN"); len(v) > 0 {
		cfg.FilterAPIAuthToken = v
	}
	if v := os.Getenv("DOWNLOAD_SERVICE_URL"); len(v) > 0 {
		cfg.DownloadServiceURL = v
	}
	if v := os.Getenv("ENABLE_DATASET_PREVIEW"); len(v) > 0 {
		var err error
		cfg.EnableDatasetPreview, err = strconv.ParseBool(v)
		if err != nil {
			log.Error(errors.WithMessage(err, "error parsing 'ENABLE_DATASET_PREVIEW' flag"), nil)

		}
	}
	return cfg
}
