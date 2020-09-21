package handlers

import (
	"fmt"
	dphandlers "github.com/ONSdigital/dp-net/handlers"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

// UseLatest will create a new filter job for the same dataset with the
// latest version
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

		datasetID, _, _, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
		if err != nil {
			log.Event(ctx, "failed to extract dataset info from path", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "path": versionPath})
			setStatusCode(req, w, err)
			return
		}

		dst, err := f.DatasetClient.Get(req.Context(), userAccessToken, "", collectionID, datasetID)
		if err != nil {
			log.Event(ctx, "failed to get dataset", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID})
			setStatusCode(req, w, err)
			return
		}

		latestVersionURL, err := url.Parse(dst.Links.LatestVersion.URL)
		if err != nil {
			log.Event(ctx, "failed to parse latest version href", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}
		latestPath := strings.TrimPrefix(latestVersionURL.Path, f.APIRouterVersion)

		_, edition, version, err := helpers.ExtractDatasetInfoFromPath(ctx, latestVersionURL.Path)
		if err != nil {
			log.Event(ctx, "failed to extract dataset info from path", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "path": latestPath})
			setStatusCode(req, w, err)
			return
		}

		newFilterID, err := f.FilterClient.CreateBlueprint(req.Context(), userAccessToken, "", "", collectionID, datasetID, edition, version, []string{})
		if err != nil {
			log.Event(ctx, "failed to create filter blueprint", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
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

			if err := f.FilterClient.AddDimensionValues(req.Context(), userAccessToken, "", collectionID, newFilterID, dim.Name, vals); err != nil {
				log.Event(ctx, "failed to add dimension values", log.ERROR, log.Error(err), log.Data{"filter_id": newFilterID, "dimension": dim.Name})
				setStatusCode(req, w, err)
				return
			}
		}

		redirectURL := fmt.Sprintf("/filters/%s/dimensions", newFilterID)
		http.Redirect(w, req, redirectURL, 302)
	})

}
