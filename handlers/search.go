package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/ONSdigital/dp-api-clients-go/headers"
	"github.com/ONSdigital/dp-api-clients-go/search"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

// Search handles a users search, calling various APIs to form a search results
// hierarchy page
func (f *Filter) Search(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)
	filterID := vars["filterID"]
	name := vars["name"]
	q := url.QueryEscape(req.URL.Query().Get("q"))

	collectionID := getCollectionIDFromContext(ctx)
	userAccessToken, err := headers.GetUserAuthToken(req)
	if err != nil {
		if headers.IsNotErrNotFound(err) {
			log.Event(ctx, "error getting access token header", log.WARN, log.Error(err))
		}
	}

	var searchConfig []search.Config
	if len(req.Header.Get("X-Florence-Token")) > 0 {
		searchConfig = append(searchConfig, search.Config{InternalToken: f.SearchAPIAuthToken, FlorenceToken: req.Header.Get("X-Florence-Token")})
	}

	fil, err := f.FilterClient.GetJobState(ctx, userAccessToken, "", "", collectionID, filterID)
	if err != nil {
		log.Event(ctx, "failed to get job state", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	selVals, err := f.FilterClient.GetDimensionOptions(ctx, userAccessToken, "", collectionID, filterID, name)
	if err != nil {
		log.Event(ctx, "failed to get options from filter client", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
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

	allVals, err := f.DatasetClient.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, name)
	if err != nil {
		log.Event(ctx, "failed to get options from dataset client", log.ERROR, log.Error(err),
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

	p := mapper.CreateHierarchySearchPage(req, searchRes.Items, d, fil, selVals, dims.Items, allVals, name, req.URL.Path, datasetID, ver.ReleaseDate, req.Referer(), req.URL.Query().Get("q"), f.APIRouterVersion)

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
}

// SearchUpdate will update a dimension based on selected search resultss
func (f *Filter) SearchUpdate(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if err := req.ParseForm(); err != nil {
		log.Event(ctx, "failed to parse request form", log.ERROR, log.Error(err))
		return
	}

	vars := mux.Vars(req)
	filterID := vars["filterID"]
	name := vars["name"]
	q := url.QueryEscape(req.Form.Get("q"))

	redirectURI := fmt.Sprintf("/filters/%s/dimensions", filterID)

	collectionID := getCollectionIDFromContext(ctx)
	userAccessToken, err := headers.GetUserAuthToken(req)
	if err != nil {
		if headers.IsNotErrNotFound(err) {
			log.Event(ctx, "error getting access token header", log.WARN, log.Error(err))
		}
	}

	fil, err := f.FilterClient.GetJobState(ctx, userAccessToken, "", "", collectionID, filterID)
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

	searchRes, err := f.SearchClient.Dimension(ctx, datasetID, edition, version, name, q)
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
		if err := f.FilterClient.AddDimensionValues(ctx, userAccessToken, "", collectionID, filterID, name, options); err != nil {
			log.Event(ctx, "failed to add dimension", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}
		return
	}

	if len(req.Form["remove-all"]) > 0 {
		for _, item := range searchRes.Items {
			if err := f.FilterClient.RemoveDimensionValue(ctx, userAccessToken, "", collectionID, filterID, name, item.Code); err != nil {
				log.Event(ctx, "failed to remove dimension option", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name, "option": item.Code})
				setStatusCode(req, w, err)
			}
		}

		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {

		opts, err := f.FilterClient.GetDimensionOptions(ctx, userAccessToken, "", collectionID, filterID, name)
		if err != nil {
			log.Event(ctx, "failed to retrieve dimension options", log.WARN, log.Error(err))
		}

		for _, item := range searchRes.Items {
			for _, opt := range opts {
				if opt.Option == item.Code {
					if _, ok := req.Form[item.Code]; !ok {
						if err := f.FilterClient.RemoveDimensionValue(ctx, userAccessToken, "", collectionID, filterID, name, item.Code); err != nil {
							log.Event(ctx, "failed to remove dimension value", log.WARN, log.Error(err))
						}
					}
				}
			}
		}

		wg.Done()
	}()

	for k := range req.Form {
		if k == "save-and-return" || k == ":uri" {
			continue
		}

		if strings.Contains(k, "redirect:") {
			redirectReg := regexp.MustCompile(`^redirect:(.+)$`)
			redirectSubs := redirectReg.FindStringSubmatch(k)
			redirectURI = redirectSubs[1]
			continue
		}

		if err := f.FilterClient.AddDimensionValue(ctx, userAccessToken, "", collectionID, filterID, name, k); err != nil {
			log.Event(ctx, "failed to add dimension value", log.WARN, log.Error(err))
			continue
		}
	}

	http.Redirect(w, req, redirectURI, 302)

}
