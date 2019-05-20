package handlers

import (
	"net/http"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
)

// Filter represents the handlers for Filtering
type Filter struct {
	Renderer           Renderer
	FilterClient       FilterClient
	DatasetClient      DatasetClient
	CodeListClient     CodelistClient
	HierarchyClient    HierarchyClient
	SearchClient       SearchClient
	val                Validator
	downloadServiceURL string
}

// NewFilter creates a new instance of Filter
func NewFilter(r Renderer, fc FilterClient, dc DatasetClient, clc CodelistClient, hc HierarchyClient, sc SearchClient, val Validator, downloadServiceURL string) *Filter {
	return &Filter{
		Renderer:           r,
		FilterClient:       fc,
		DatasetClient:      dc,
		CodeListClient:     clc,
		HierarchyClient:    hc,
		SearchClient:       sc,
		val:                val,
		downloadServiceURL: downloadServiceURL,
	}
}

func setStatusCode(req *http.Request, w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if err, ok := err.(ClientError); ok {
		if err.Code() == http.StatusNotFound {
			status = err.Code()
		}
	}
	log.ErrorCtx(req.Context(), err, log.Data{"setting-response-status": status})
	w.WriteHeader(status)
}

func forwardFlorenceTokenIfRequired(req *http.Request) *http.Request {
	if len(req.Header.Get(common.FlorenceHeaderKey)) > 0 {
		ctx := common.SetFlorenceIdentity(req.Context(), req.Header.Get(common.FlorenceHeaderKey))
		return req.WithContext(ctx)
	}
	return req
}
