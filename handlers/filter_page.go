package handlers

import (
	"net/http"
	"strings"

	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// Handler is middleware that a accepts a filter and dataset client that returns either the filter or filterFlex handler dependent on the type of dataset
// FilterPageHandler handles requests to /filters/{filterID}
func FilterPageHandler(f FilterClient, datasetClient DatasetAPISdkClient, filter, filterFlex http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)

		userAuthToken := ""
		serviceAuthToken := ""
		collectionID := ""
		downloadServiceToken := ""

		// Obtain access_token from cookie
		c, err := r.Cookie(`access_token`)
		if err == nil && c.Value != "" {
			userAuthToken = c.Value
			log.Info(r.Context(), "obtained access_token Cookie")
		}

		headers := dpDatasetApiSdk.Headers{
			CollectionID:         collectionID,
			DownloadServiceToken: downloadServiceToken,
			ServiceToken:         serviceAuthToken,
			UserAccessToken:      userAuthToken,
		}

		filterID := vars["filterID"]
		if filterID == "" {
			http.Error(w, "missing filter ID", http.StatusBadRequest)
			return
		}

		filterModel, _, err := f.GetJobState(
			ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterID,
		)
		if err != nil {
			log.Error(ctx, "failed to get filter job state", err)
			http.Error(w, "failed to get filter", http.StatusInternalServerError)
			return
		}

		datasetDetails, err := datasetClient.GetDataset(ctx, headers, collectionID, filterModel.Dataset.DatasetID)
		if err != nil {
			log.Error(ctx, "failed to get dataset details", err)
			http.Error(w, "failed to get dataset", http.StatusInternalServerError)
			return
		}

		if strings.Contains(datasetDetails.Type, "cantabular") {
			// Redirect to dp-frontend-filter-flex-dataset and continue filter-flex journey
			filterFlex.ServeHTTP(w, r)
			return
		}

		// If CMD type, the CMD filter journey works as it currently does i.e. to frontend-filter-dataset-controller
		filter.ServeHTTP(w, r)
	}
}
