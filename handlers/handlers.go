package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/filter"
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
	val                  Validator
	downloadServiceURL   string
	EnableDatasetPreview bool
	APIRouterVersion     string
}

// NewFilter creates a new instance of Filter
func NewFilter(r Renderer, fc FilterClient, dc DatasetClient, hc HierarchyClient,
	sc SearchClient, val Validator, searchAPIAuthToken, downloadServiceURL, apiRouterVersion string, enableDatasetPreview bool) *Filter {

	return &Filter{
		Renderer:             r,
		FilterClient:         fc,
		DatasetClient:        dc,
		HierarchyClient:      hc,
		SearchClient:         sc,
		val:                  val,
		downloadServiceURL:   downloadServiceURL,
		EnableDatasetPreview: enableDatasetPreview,
		SearchAPIAuthToken:   searchAPIAuthToken,
		APIRouterVersion:     apiRouterVersion,
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
