package handlers

import (
	"fmt"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// UseLatest will create a new filter job for the same dataset with the
// latest version
func (f *Filter) UseLatest(w http.ResponseWriter, req *http.Request) {
	cfg := config.Get()
	vars := mux.Vars(req)
	filterID := vars["filterID"]
	ctx := req.Context()

	req = forwardFlorenceTokenIfRequired(req)

	oldJob, err := f.FilterClient.GetJobState(req.Context(), cfg.ServiceAuthToken, "", filterID)
	if err != nil {
		log.InfoCtx(ctx, "failed to get job state", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	dims, err := f.FilterClient.GetDimensions(req.Context(), cfg.ServiceAuthToken, filterID)
	if err != nil {
		log.InfoCtx(ctx, "failed to get dimensions", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	versionURL, err := url.Parse(oldJob.Links.Version.HRef)
	if err != nil {
		log.InfoCtx(ctx, "failed to parse version href", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	datasetID, _, _, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		log.InfoCtx(ctx, "failed to extract dataset info from path", log.Data{"error": err, "filter_id": filterID, "path": versionURL})
		setStatusCode(req, w, err)
		return
	}

	dst, err := f.DatasetClient.Get(req.Context(), datasetID)
	if err != nil {
		log.InfoCtx(ctx, "failed to get dataset", log.Data{"error": err, "dataset_id": datasetID})
		setStatusCode(req, w, err)
		return
	}

	latestVersionURL, err := url.Parse(dst.Links.LatestVersion.URL)
	if err != nil {
		log.InfoCtx(ctx, "failed to parse latest version href", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	_, edition, version, err := helpers.ExtractDatasetInfoFromPath(latestVersionURL.Path)
	if err != nil {
		log.InfoCtx(ctx, "failed to extract dataset info from path", log.Data{"error": err, "filter_id": filterID, "path": versionURL})
		setStatusCode(req, w, err)
		return
	}

	newFilterID, err := f.FilterClient.CreateBlueprint(req.Context(), cfg.ServiceAuthToken, "", datasetID, edition, version, []string{})
	if err != nil {
		log.InfoCtx(ctx, "failed to create filter blueprint", log.Data{"error": err, "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	for _, dim := range dims {
		if err := f.FilterClient.AddDimension(req.Context(), cfg.ServiceAuthToken, newFilterID, dim.Name); err != nil {
			log.InfoCtx(ctx, "failed to add dimension", log.Data{"error": err, "filter_id": filterID, "dimension": dim.Name})
			setStatusCode(req, w, err)
			return
		}

		dimValues, err := f.FilterClient.GetDimensionOptions(req.Context(), cfg.ServiceAuthToken, filterID, dim.Name)
		if err != nil {
			log.InfoCtx(ctx, "failed to get options from filter client", log.Data{"error": err, "filter_id": filterID, "dimension": dim.Name})
			setStatusCode(req, w, err)
			return
		}

		var vals []string
		for _, val := range dimValues {
			vals = append(vals, val.Option)
		}

		if err := f.FilterClient.AddDimensionValues(req.Context(), cfg.ServiceAuthToken, newFilterID, dim.Name, vals); err != nil {
			log.InfoCtx(ctx, "failed to add dimension values", log.Data{"error": err, "filter_id": newFilterID, "dimension": dim.Name})
			setStatusCode(req, w, err)
			return
		}
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions", newFilterID)
	http.Redirect(w, req, redirectURL, 302)

}
