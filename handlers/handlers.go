package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/log.go/v2/log"
)

// Filter represents the handlers for Filtering
type Filter struct {
	Render               RenderClient
	FilterClient         FilterClient
	DatasetClient        DatasetClient
	ZebedeeClient        ZebedeeClient
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
func NewFilter(rc RenderClient, fc FilterClient, dc DatasetClient, hc HierarchyClient,
	sc SearchClient, zc ZebedeeClient, apiRouterVersion string, cfg *config.Config) *Filter {
	return &Filter{
		Render:               rc,
		FilterClient:         fc,
		DatasetClient:        dc,
		HierarchyClient:      hc,
		SearchClient:         sc,
		ZebedeeClient:        zc,
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
	log.Info(req.Context(), "setting response status", log.FormatErrors([]error{err}), log.Data{"status": status})
	w.WriteHeader(status)
}
