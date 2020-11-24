package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

		oldJob, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
		if err != nil {
			log.Event(ctx, "failed to get job state", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		dims, err := f.FilterClient.GetDimensions(req.Context(), userAccessToken, "", collectionID, filterID)
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

		newFilterID, err := f.FilterClient.CreateBlueprint(req.Context(), userAccessToken, "", "", collectionID, datasetID, edition, string(editionDetails.Links.LatestVersion.ID), []string{})
		if err != nil {
			log.Event(ctx, "failed to create filter blueprint", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": editionDetails.Links.LatestVersion.ID})
			setStatusCode(req, w, err)
			return
		}

		for _, dim := range dims {
			if err := f.FilterClient.AddDimension(req.Context(), userAccessToken, "", collectionID, newFilterID, dim.Name); err != nil {
				log.Event(ctx, "failed to add dimension", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dim.Name})
				setStatusCode(req, w, err)
				return
			}

			dimValues, err := f.FilterClient.GetDimensionOptions(req.Context(), userAccessToken, "", collectionID, filterID, dim.Name)
			if err != nil {
				log.Event(ctx, "failed to get options from filter client", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dim.Name})
				setStatusCode(req, w, err)
				return
			}

			var vals []string
			for _, val := range dimValues {
				vals = append(vals, val.Option)
			}

			if err := f.FilterClient.SetDimensionValues(req.Context(), userAccessToken, "", collectionID, newFilterID, dim.Name, vals); err != nil {
				log.Event(ctx, "failed to add dimension values", log.ERROR, log.Error(err), log.Data{"filter_id": newFilterID, "dimension": dim.Name})
				setStatusCode(req, w, err)
				return
			}
		}

		redirectURL := fmt.Sprintf("/filters/%s/dimensions", newFilterID)
		http.Redirect(w, req, redirectURL, 302)
	})

}
