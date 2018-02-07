package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"unicode"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// FilterOverview controls the render of the filter overview template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) FilterOverview(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	datasetCfg := setAuthTokenIfRequired(req)

	dims, err := f.FilterClient.GetDimensions(filterID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	fj, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	datasetDimensions, err := f.DatasetClient.GetDimensions(datasetID, edition, version, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dimensionIDNameLookup := make(map[string]map[string]string)
	for _, dim := range datasetDimensions.Items {
		idNameLookup := make(map[string]string)
		options, err := f.DatasetClient.GetOptions(datasetID, edition, version, dim.ID, datasetCfg...)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		for _, opt := range options.Items {
			idNameLookup[opt.Option] = opt.Label
		}
		dimensionIDNameLookup[dim.ID] = idNameLookup
	}

	var dimensions FilterModelDimensions
	for _, dim := range dims {
		var vals []filter.DimensionOption
		vals, err = f.FilterClient.GetDimensionOptions(filterID, dim.Name)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}
		var values []string
		for _, val := range vals {
			values = append(values, dimensionIDNameLookup[dim.Name][val.Option])
		}

		dimensions = append(dimensions, filter.ModelDimension{
			Name:   dim.Name,
			Values: values,
		})
	}

	sort.Sort(dimensions)

	dataset, err := f.DatasetClient.Get(datasetID, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}
	ver, err := f.DatasetClient.GetVersion(datasetID, edition, version, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	latestURL, err := url.Parse(dataset.Links.LatestVersion.URL)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	p := mapper.CreateFilterOverview(dimensions, fj, dataset, filterID, datasetID, ver.ReleaseDate)

	if latestURL.Path == versionURL.Path {
		p.Data.IsLatestVersion = true
	}

	p.Data.LatestVersion.DatasetLandingPageURL = latestURL.Path
	p.Data.LatestVersion.FilterJourneyWithLatestJourney = fmt.Sprintf("/filters/%s/use-latest-version", filterID)

	b, err := json.Marshal(p)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/filter-overview", b)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	w.Write(templateBytes)
}

// FilterOverviewClearAll removes all selected options for all dimensions
func (f *Filter) FilterOverviewClearAll(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	dims, err := f.FilterClient.GetDimensions(filterID)
	if err != nil {
		log.ErrorR(req, err, nil)
		return
	}

	for _, dim := range dims {
		if err := f.FilterClient.RemoveDimension(filterID, dim.Name); err != nil {
			setStatusCode(req, w, err)
			return
		}

		if err := f.FilterClient.AddDimension(filterID, dim.Name); err != nil {
			setStatusCode(req, w, err)
			return
		}
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)

	http.Redirect(w, req, redirectURL, 302)
}

// FilterModelDimensions represents a list of dimensions
type FilterModelDimensions []filter.ModelDimension

func (d FilterModelDimensions) Len() int      { return len(d) }
func (d FilterModelDimensions) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d FilterModelDimensions) Less(i, j int) bool {
	iRunes := []rune(d[i].Name)
	jRunes := []rune(d[j].Name)

	max := len(iRunes)
	if max > len(jRunes) {
		max = len(jRunes)
	}

	for idx := 0; idx < max; idx++ {
		ir := iRunes[idx]
		jr := jRunes[idx]

		lir := unicode.ToLower(ir)
		ljr := unicode.ToLower(jr)

		if lir != ljr {
			return lir < ljr
		}

		// the lowercase runes are the same, so compare the original
		if ir != jr {
			return ir < jr
		}
	}

	return false
}
