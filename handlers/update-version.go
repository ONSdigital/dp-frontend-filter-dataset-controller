package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/filter"
	dphandlers "github.com/ONSdigital/dp-net/handlers"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

// UseLatest will create a new filter job for the same dataset with the
// latest version in that edition
func (f *Filter) UseLatest() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		ctx := req.Context()

		oldJob, _, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
		if err != nil {
			log.Event(ctx, "failed to get job state", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		dims, _, err := f.FilterClient.GetDimensions(req.Context(), userAccessToken, "", collectionID, filterID, nil)
		if err != nil {
			log.Event(ctx, "failed to get dimensions", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		versionURL, err := url.Parse(oldJob.Links.Version.HRef)
		if err != nil || versionURL.Path == "" {
			log.Event(ctx, "failed to parse version href", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}
		versionPath := strings.TrimPrefix(versionURL.Path, f.APIRouterVersion)

		datasetID, edition, _, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
		if err != nil {
			log.Event(ctx, "failed to extract dataset info from path", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "path": versionPath})
			setStatusCode(req, w, err)
			return
		}

		editionDetails, err := f.DatasetClient.GetEdition(req.Context(), userAccessToken, "", collectionID, datasetID, edition)
		if err != nil {
			log.Event(ctx, "failed to get edition details", log.ERROR, log.Error(err), log.Data{"dataset": datasetID, "edition": edition})
			setStatusCode(req, w, err)
			return
		}

		newFilterID, newFilterETag, err := f.FilterClient.CreateBlueprint(req.Context(), userAccessToken, "", "", collectionID, datasetID, edition, string(editionDetails.Links.LatestVersion.ID), []string{})
		if err != nil {
			log.Event(ctx, "failed to create filter blueprint", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": editionDetails.Links.LatestVersion.ID})
			setStatusCode(req, w, err)
			return
		}

		for _, dim := range dims.Items {

			// Copy dimension to new filter
			newFilterETag, err = f.FilterClient.AddDimension(req.Context(), userAccessToken, "", collectionID, newFilterID, dim.Name, newFilterETag)
			if err != nil {
				log.Event(ctx, "failed to add dimension", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dim.Name})
				setStatusCode(req, w, err)
				return
			}

			// Copy each batch of options to the new filter dimension via PATCH operations.
			processBatch := f.batchAddOptions(req.Context(), userAccessToken, collectionID, newFilterID, dim.Name, newFilterETag)

			// Call filter API GetOptions in batches and aggregate the responses
			newFilterETag, err = f.FilterClient.GetDimensionOptionsBatchProcess(req.Context(), userAccessToken, "", collectionID, filterID, dim.Name, processBatch, f.BatchSize, f.BatchMaxWorkers, true)
			if err != nil {
				log.Event(ctx, "failed to get and process options from filter client in batches", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dim.Name})
				setStatusCode(req, w, err)
				return
			}
		}

		redirectURL := fmt.Sprintf("/filters/%s/dimensions", newFilterID)
		http.Redirect(w, req, redirectURL, 302)
	})

}

// batchAddOptions generates a batch processor to add the dimension options for each provided batch to filter API, by calling the patch endpoint.
func (f *Filter) batchAddOptions(ctx context.Context, userAccessToken, collectionID, filterID, dimensionName, initialETag string) filter.DimensionOptionsBatchProcessor {
	currentETag := initialETag
	return func(batch filter.DimensionOptions, oldFilterETag string) (forceAbort bool, err error) {
		var vals []string
		for _, val := range batch.Items {
			vals = append(vals, val.Option)
		}
		currentETag, err = f.FilterClient.PatchDimensionValues(ctx, userAccessToken, "", collectionID, filterID, dimensionName, vals, []string{}, f.BatchSize, currentETag)
		return false, err
	}
}
