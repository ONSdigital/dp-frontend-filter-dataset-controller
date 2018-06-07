package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/ONSdigital/go-ns/clients/search"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// Search handles a users search, calling various APIs to form a search results
// hierarchy page
func (f *Filter) Search(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]
	name := vars["name"]
	q := url.QueryEscape(req.URL.Query().Get("q"))

	req = forwardFlorenceTokenIfRequired(req)

	var searchConfig []search.Config
	if len(req.Header.Get("X-Florence-Token")) > 0 {
		cfg := config.Get()
		searchConfig = append(searchConfig, search.Config{InternalToken: cfg.SearchAPIAuthToken, FlorenceToken: req.Header.Get("X-Florence-Token")})
	}

	fil, err := f.FilterClient.GetJobState(req.Context(), filterID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	selVals, err := f.FilterClient.GetDimensionOptions(req.Context(), filterID, name)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	versionURL, err := url.Parse(fil.Links.Version.HRef)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	d, err := f.DatasetClient.Get(req.Context(), datasetID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}
	ver, err := f.DatasetClient.GetVersion(req.Context(), datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	allVals, err := f.DatasetClient.GetOptions(req.Context(), datasetID, edition, version, name)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	searchRes, err := f.SearchClient.Dimension(datasetID, edition, version, name, q, searchConfig...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dims, err := f.DatasetClient.GetDimensions(req.Context(), datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	p := mapper.CreateHierarchySearchPage(req.Context(), searchRes.Items, d, fil, selVals, dims.Items, allVals, name, req.URL.Path, datasetID, ver.ReleaseDate, req.Referer(), req.URL.Query().Get("q"))

	b, err := json.Marshal(p)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/hierarchy", b)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	w.Write(templateBytes)
}

// SearchUpdate will update a dimension based on selected search resultss
func (f *Filter) SearchUpdate(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if err := req.ParseForm(); err != nil {
		log.ErrorCtx(ctx, err, nil)
		return
	}

	vars := mux.Vars(req)
	filterID := vars["filterID"]
	name := vars["name"]
	q := url.QueryEscape(req.Form.Get("q"))

	redirectURI := fmt.Sprintf("/filters/%s/dimensions", filterID)

	req = forwardFlorenceTokenIfRequired(req)

	fil, err := f.FilterClient.GetJobState(req.Context(), filterID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	versionURL, err := url.Parse(fil.Links.Version.HRef)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	var searchConfig []search.Config
	if len(req.Header.Get("X-Florence-Token")) > 0 {
		cfg := config.Get()
		searchConfig = append(searchConfig, search.Config{InternalToken: cfg.SearchAPIAuthToken, FlorenceToken: req.Header.Get("X-Florence-Token")})
	}

	searchRes, err := f.SearchClient.Dimension(datasetID, edition, version, name, q, searchConfig...)
	if err != nil {
		log.ErrorCtx(ctx, err, nil)
		http.Redirect(w, req, fmt.Sprintf("/filters/%s/dimensions", filterID), 302)
		return
	}

	if len(req.Form["add-all"]) > 0 {
		var options []string
		for _, item := range searchRes.Items {
			options = append(options, item.Code)
		}
		if err := f.FilterClient.AddDimensionValues(req.Context(), filterID, name, options); err != nil {
			setStatusCode(req, w, err)
			return
		}
		return
	}

	if len(req.Form["remove-all"]) > 0 {
		for _, item := range searchRes.Items {
			if err := f.FilterClient.RemoveDimensionValue(req.Context(), filterID, name, item.Code); err != nil {
				setStatusCode(req, w, err)
			}
		}

		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {

		opts, err := f.FilterClient.GetDimensionOptions(req.Context(), filterID, name)
		if err != nil {
			log.ErrorCtx(ctx, err, nil)
		}

		for _, item := range searchRes.Items {
			for _, opt := range opts {
				if opt.Option == item.Code {
					if _, ok := req.Form[item.Code]; !ok {
						if err := f.FilterClient.RemoveDimensionValue(req.Context(), filterID, name, item.Code); err != nil {
							log.ErrorCtx(ctx, err, nil)
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

		if err := f.FilterClient.AddDimensionValue(req.Context(), filterID, name, k); err != nil {
			log.InfoCtx(ctx, err.Error(), nil)
			continue
		}
	}

	http.Redirect(w, req, redirectURI, 302)

}
