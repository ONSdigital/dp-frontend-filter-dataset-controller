package config

import (
	"os"
)

// Config represents service configuration for dp-frontend-filter-dataset-controller
type Config struct {
	BindAddr            string
	RendererURL         string
	FilterAPIURL        string
	DatasetAPIURL       string
	CodeListAPIURL      string
	HierarchyAPIURL     string
	DatasetAPIAuthToken string
}

// Get returns the default config with any modifications through environment
// variables
func Get() *Config {
	cfg := &Config{
		BindAddr:        ":20001",
		RendererURL:     "http://localhost:20010",
		CodeListAPIURL:  "http://localhost:22400",
		FilterAPIURL:    "http://localhost:22100",
		DatasetAPIURL:   "http://localhost:22000",
		HierarchyAPIURL: "http://localhost:22600",
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
	if v := os.Getenv("CODELIST_API_URL"); len(v) > 0 {
		cfg.CodeListAPIURL = v
	}
	if v := os.Getenv("DATASET_API_URL"); len(v) > 0 {
		cfg.DatasetAPIURL = v
	}
	if v := os.Getenv("HIERARCHY_API_URL"); len(v) > 0 {
		cfg.HierarchyAPIURL = v
	}
	if v := os.Getenv("DATASET_API_AUTH_TOKEN"); len(v) > 0 {
		cfg.DatasetAPIAuthToken = v
	}

	return cfg
}
