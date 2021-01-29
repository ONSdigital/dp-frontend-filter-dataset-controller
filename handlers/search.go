package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/search"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	dphandlers "github.com/ONSdigital/dp-net/handlers"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

// Search handles a users search, calling various APIs to form a search results
// hierarchy page
func (f *Filter) Search() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		ctx := req.Context()
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		name := vars["name"]
		q := url.QueryEscape(req.URL.Query().Get("q"))

		var searchConfig []search.Config
		if len(req.Header.Get("X-Florence-Token")) > 0 {
			searchConfig = append(searchConfig, search.Config{InternalToken: f.SearchAPIAuthToken, FlorenceToken: req.Header.Get("X-Florence-Token")})
		}

		fil, eTag0, err := f.FilterClient.GetJobState(ctx, userAccessToken, "", "", collectionID, filterID)
		if err != nil {
			log.Event(ctx, "failed to get job state", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		selVals, eTag1, err := f.FilterClient.GetDimensionOptionsInBatches(ctx, userAccessToken, "", collectionID, filterID, name, f.BatchSize, f.BatchMaxWorkers)
		if err != nil {
			log.Event(ctx, "failed to get options from filter client", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}

		// The user might want to retry this handler if eTags don't match
		if eTag0 != eTag1 {
			err := errors.New("inconsistent filter data")
			log.Event(ctx, "data consistency cannot be guaranteed because filter was modified between calls", log.ERROR, log.Error(err),
				log.Data{"filter_id": filterID, "dimension": name, "e_tag_0": eTag0, "e_tag_1": eTag1})
			setStatusCode(req, w, err)
			return
		}

		versionURL, err := url.Parse(fil.Links.Version.HRef)
		if err != nil {
			log.Event(ctx, "failed to parse version href", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}
		versionPath := strings.TrimPrefix(versionURL.Path, f.APIRouterVersion)
		datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
		if err != nil {
			log.Event(ctx, "failed to extract dataset info from path", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "path": versionPath})
			setStatusCode(req, w, err)
			return
		}

		d, err := f.DatasetClient.Get(ctx, userAccessToken, "", collectionID, datasetID)
		if err != nil {
			log.Event(ctx, "failed to get dataset", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID})
			setStatusCode(req, w, err)
			return
		}
		ver, err := f.DatasetClient.GetVersion(ctx, userAccessToken, "", "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Event(ctx, "failed to get version", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		selValsLabelMap, err := f.getIDNameLookupFromDatasetAPI(ctx, userAccessToken, collectionID, datasetID, edition, version, name, selVals)
		if err != nil {
			log.Event(ctx, "failed to get options from dataset client for the selected values", log.ERROR, log.Error(err),
				log.Data{"dimension": name, "dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		searchRes, err := f.SearchClient.Dimension(ctx, datasetID, edition, version, name, q, searchConfig...)
		if err != nil {
			log.Event(ctx, "failed to get dimension from search client", log.ERROR, log.Error(err),
				log.Data{"dimension": name, "dataset_id": datasetID, "edition": edition, "version": version, "query": q})
			setStatusCode(req, w, err)
			return
		}

		dims, err := f.DatasetClient.GetVersionDimensions(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Event(ctx, "failed to get dimensions", log.ERROR, log.Error(err),
				log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		p := mapper.CreateHierarchySearchPage(req, searchRes.Items, d, fil, selValsLabelMap, dims.Items, name, req.URL.Path, datasetID, ver.ReleaseDate, req.Referer(), req.URL.Query().Get("q"), f.APIRouterVersion, lang)

		b, err := json.Marshal(p)
		if err != nil {
			log.Event(ctx, "failed to marshal json", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		templateBytes, err := f.Renderer.Do("dataset-filter/hierarchy", b)
		if err != nil {
			log.Event(ctx, "failed to render", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateBytes)
	})

}

// SearchUpdate will update a dimension based on selected search results
func (f *Filter) SearchUpdate() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		ctx := req.Context()
		if err := req.ParseForm(); err != nil {
			log.Event(ctx, "failed to parse request form", log.ERROR, log.Error(err))
			return
		}

		var searchConfig []search.Config
		if len(req.Header.Get("X-Florence-Token")) > 0 {
			searchConfig = append(searchConfig, search.Config{InternalToken: f.SearchAPIAuthToken, FlorenceToken: req.Header.Get("X-Florence-Token")})
		}

		vars := mux.Vars(req)
		filterID := vars["filterID"]
		name := vars["name"]
		q := url.QueryEscape(req.Form.Get("q"))

		redirectURI := fmt.Sprintf("/filters/%s/dimensions", filterID)

		fil, eTag, err := f.FilterClient.GetJobState(ctx, userAccessToken, "", "", collectionID, filterID)
		if err != nil {
			log.Event(ctx, "failed to get job state", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		versionURL, err := url.Parse(fil.Links.Version.HRef)
		if err != nil {
			log.Event(ctx, "failed to parse version href", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}
		versionPath := strings.TrimPrefix(versionURL.Path, f.APIRouterVersion)
		datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
		if err != nil {
			log.Event(ctx, "failed to extract dataset info from path", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "path": versionPath})
			setStatusCode(req, w, err)
			return
		}

		searchRes, err := f.SearchClient.Dimension(ctx, datasetID, edition, version, name, q, searchConfig...)
		if err != nil {
			log.Event(ctx, "failed to retrieve dimension search result, redirecting", log.ERROR, log.Error(err))
			http.Redirect(w, req, fmt.Sprintf("/filters/%s/dimensions", filterID), 302)
			return
		}

		if len(req.Form["add-all"]) > 0 {
			var options []string
			for _, item := range searchRes.Items {
				options = append(options, item.Code)
			}
			_, err = f.FilterClient.SetDimensionValues(ctx, userAccessToken, "", collectionID, filterID, name, options, eTag)
			if err != nil {
				log.Event(ctx, "failed to add all dimension options", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
				setStatusCode(req, w, err)
				return
			}
			return
		}

		if len(req.Form["remove-all"]) > 0 {
			options := []string{}
			for _, item := range searchRes.Items {
				options = append(options, item.Code)
			}
			_, err = f.FilterClient.PatchDimensionValues(ctx, userAccessToken, "", collectionID, filterID, name, []string{}, options, f.BatchSize, eTag)
			if err != nil {
				log.Event(ctx, "failed to remove all dimension options, via patch", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
				setStatusCode(req, w, err)
				return
			}
			return
		}

		// get all available dimension options from filter API
		opts, eTag1, err := f.FilterClient.GetDimensionOptionsInBatches(ctx, userAccessToken, "", collectionID, filterID, name, f.BatchSize, f.BatchMaxWorkers)
		if err != nil {
			log.Event(ctx, "failed to retrieve dimension options", log.WARN, log.Error(err))
			setStatusCode(req, w, err)
			return
		}

		// The user might want to retry this handler if eTags don't match
		if eTag != eTag1 {
			err := errors.New("inconsistent filter data")
			log.Event(ctx, "data consistency cannot be guaranteed because filter was modified between get calls", log.ERROR, log.Error(err),
				log.Data{"filter_id": filterID, "dimension": name, "e_tag_0": eTag, "e_tag_1": eTag1})
			setStatusCode(req, w, err)
			return
		}

		// create list of options to remove
		removeOptions := []string{}
		for _, item := range searchRes.Items {
			for _, opt := range opts.Items {
				if opt.Option == item.Code {
					if _, ok := req.Form[item.Code]; !ok {
						removeOptions = append(removeOptions, item.Code)
					}
				}
			}
		}

		// get options to add and overwrite redirectURI, if provided in the form
		var addOptions []string
		addOptions = getOptionsAndRedirect(req.Form, &redirectURI)

		// sent the PATCH with options to add and remove
		_, err = f.FilterClient.PatchDimensionValues(ctx, userAccessToken, "", collectionID, filterID, name, addOptions, removeOptions, f.BatchSize, eTag1)
		if err != nil {
			log.Event(ctx, "failed to patch dimension values", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}

		http.Redirect(w, req, redirectURI, 302)
	})

}
