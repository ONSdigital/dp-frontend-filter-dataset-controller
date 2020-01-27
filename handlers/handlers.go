package handlers

import (
	"context"
	"net/http"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/log"
	"github.com/pkg/errors"
)

// Filter represents the handlers for Filtering
type Filter struct {
	Renderer             Renderer
	FilterClient         FilterClient
	DatasetClient        DatasetClient
	HierarchyClient      HierarchyClient
	SearchClient         SearchClient
	val                  Validator
	downloadServiceURL   string
	EnableDatasetPreview bool
	EnableLoop11         bool
}

// NewFilter creates a new instance of Filter
func NewFilter(r Renderer, fc FilterClient, dc DatasetClient, hc HierarchyClient, sc SearchClient, val Validator, downloadServiceURL string, enableDatasetPreview, enableLoop11 bool) *Filter {
	return &Filter{
		Renderer:             r,
		FilterClient:         fc,
		DatasetClient:        dc,
		HierarchyClient:      hc,
		SearchClient:         sc,
		val:                  val,
		downloadServiceURL:   downloadServiceURL,
		EnableDatasetPreview: enableDatasetPreview,
		EnableLoop11:         enableLoop11,
	}
}

func setStatusCode(req *http.Request, w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if err, ok := err.(ClientError); ok {
		if err.Code() == http.StatusNotFound {
			status = err.Code()
		}
	}
	log.Event(req.Context(), "setting response status", log.Error(err), log.Data{"status": status})
	w.WriteHeader(status)
}

func getCollectionIDFromContext(ctx context.Context) string {
	if ctx.Value(common.CollectionIDHeaderKey) != nil {
		collectionID, ok := ctx.Value(common.CollectionIDHeaderKey).(string)
		if !ok {
			log.Event(ctx, "failed to get collection ID", log.Error(errors.New("error casting collection ID context value to string")))
		}
		return collectionID
	}
	return ""
}
