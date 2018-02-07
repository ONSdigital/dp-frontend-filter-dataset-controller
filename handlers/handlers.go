package handlers

import (
	"net/http"
	"os"

	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/log"
)

var datasetAuthToken = os.Getenv("DATASET_API_AUTH_TOKEN")

// Filter represents the handlers for Filtering
type Filter struct {
	Renderer        Renderer
	FilterClient    FilterClient
	DatasetClient   DatasetClient
	CodeListClient  CodelistClient
	HierarchyClient HierarchyClient
	val             Validator
}

// NewFilter creates a new instance of Filter
func NewFilter(r Renderer, fc FilterClient, dc DatasetClient, clc CodelistClient, hc HierarchyClient, val Validator) *Filter {
	return &Filter{
		Renderer:        r,
		FilterClient:    fc,
		DatasetClient:   dc,
		CodeListClient:  clc,
		HierarchyClient: hc,
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

func setAuthTokenIfRequired(req *http.Request) []dataset.Config {
	var datasetConfig []dataset.Config
	if len(req.Header.Get("X-Florence-Token")) > 0 {
		datasetConfig = append(datasetConfig, dataset.Config{InternalToken: datasetAuthToken})
	}
	return datasetConfig
}
