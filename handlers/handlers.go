package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/log.go/log"
)

// Filter represents the handlers for Filtering
type Filter struct {
	Renderer             Renderer
	FilterClient         FilterClient
	DatasetClient        DatasetClient
	HierarchyClient      HierarchyClient
	SearchClient         SearchClient
	SearchAPIAuthToken   string
	downloadServiceURL   string
	EnableDatasetPreview bool
	APIRouterVersion     string
	BatchSize            int
	BatchMaxWorkers      int
	maxDatasetOptions    int
}

// NewFilter creates a new instance of Filter
func NewFilter(r Renderer, fc FilterClient, dc DatasetClient, hc HierarchyClient,
	sc SearchClient, apiRouterVersion string, cfg *config.Config) *Filter {
	return &Filter{
		Renderer:             r,
		FilterClient:         fc,
		DatasetClient:        dc,
		HierarchyClient:      hc,
		SearchClient:         sc,
		APIRouterVersion:     apiRouterVersion,
		downloadServiceURL:   cfg.DownloadServiceURL,
		EnableDatasetPreview: cfg.EnableDatasetPreview,
		SearchAPIAuthToken:   cfg.SearchAPIAuthToken,
		BatchSize:            cfg.BatchSizeLimit,
		BatchMaxWorkers:      cfg.BatchMaxWorkers,
		maxDatasetOptions:    cfg.MaxDatasetOptions,
	}
}

func setStatusCode(req *http.Request, w http.ResponseWriter, err error) {
	status := http.StatusOK
	if err != nil {
		switch err.(type) {
		case filter.ErrInvalidFilterAPIResponse:
			status = http.StatusBadGateway
		case ClientError:
			status = err.(ClientError).Code()
		default:
			status = http.StatusInternalServerError
		}
	}
	log.Event(req.Context(), "setting response status", log.INFO, log.Error(err), log.Data{"status": status})
	w.WriteHeader(status)
}
