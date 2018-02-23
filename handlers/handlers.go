package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/log"
)

// Filter represents the handlers for Filtering
type Filter struct {
	Renderer        Renderer
	FilterClient    FilterClient
	DatasetClient   DatasetClient
	CodeListClient  CodelistClient
	HierarchyClient HierarchyClient
	SearchClient    SearchClient
	val             Validator
}

// NewFilter creates a new instance of Filter
func NewFilter(r Renderer, fc FilterClient, dc DatasetClient, clc CodelistClient, hc HierarchyClient, sc SearchClient, val Validator) *Filter {
	return &Filter{
		Renderer:        r,
		FilterClient:    fc,
		DatasetClient:   dc,
		CodeListClient:  clc,
		HierarchyClient: hc,
		SearchClient:    sc,
		val:             val,
	}
}

func setStatusCode(req *http.Request, w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if err, ok := err.(ClientError); ok {
		if err.Code() == http.StatusNotFound {
			status = err.Code()
		}
	}
	log.ErrorR(req, err, log.Data{"setting-response-status": status})
	w.WriteHeader(status)
}

func setAuthTokenIfRequired(req *http.Request) ([]dataset.Config, []filter.Config) {
	var datasetConfig []dataset.Config
	var filterConfig []filter.Config
	if len(req.Header.Get("X-Florence-Token")) > 0 {
		cfg := config.Get()
		datasetConfig = append(datasetConfig, dataset.Config{InternalToken: cfg.DatasetAPIAuthToken})
		filterConfig = append(filterConfig, filter.Config{InternalToken: cfg.FilterAPIAuthToken})
	}
	return datasetConfig, filterConfig
}
