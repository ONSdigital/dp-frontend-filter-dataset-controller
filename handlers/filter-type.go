package handlers

import (
	"net/http"
	"strings"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/log.go/v2/log"
)

// Handler is middleware that a accepts a filter and dataset client that returns either the filter or filterFlex handler dependent on the type of dataset
func (f *Filter) FilterType(datasetClient DatasetClient, filter, filterFlex http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		ctx := req.Context()

		filterID, err := helpers.ReturnSecondSegmentFromPath(path)
		if err != nil {
			log.Error(ctx, "failed to extract filter id info from path", err, log.Data{"filter_id": filterID, "path": path})
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Obtain access_token from cookie
		userAccessToken := ""
		c, err := req.Cookie(`access_token`)
		if err == nil && c.Value != "" {
			userAccessToken = c.Value
			log.Info(req.Context(), "obtained access_token Cookie")
		}

		f, _, err := f.FilterClient.GetJobState(ctx, userAccessToken, "", "", "", filterID)
		if err != nil {
			log.Warn(ctx, "failed to get job state - falling through to default handling", log.Data{"filter_id": filterID})
			filter.ServeHTTP(w, req)
			return
		}

		d, err := datasetClient.Get(ctx, userAccessToken, "", "", f.Dataset.DatasetID)
		if err != nil {
			log.Warn(ctx, "failed to get dataset id - falling through to default handling", log.Data{"dataset_id": f.Dataset.DatasetID})
			filter.ServeHTTP(w, req)
			return
		}

		if strings.Contains(d.Type, "cantabular") {
			log.Info(ctx, "using flex handler")
			filterFlex.ServeHTTP(w, req)
			return
		}
		log.Info(ctx, "using filter handler")
		filter.ServeHTTP(w, req)
	})
}
