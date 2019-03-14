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
	ctx := req.Context()

	req = forwardFlorenceTokenIfRequired(req)

	dims, err := f.FilterClient.GetDimensions(req.Context(), filterID)
	if err != nil {
		log.InfoCtx(ctx, "failed to get dimensions", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	fj, err := f.FilterClient.GetJobState(req.Context(), filterID)
	if err != nil {
		log.InfoCtx(ctx, "failed to get job state", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		log.InfoCtx(ctx, "failed to parse version href", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		log.InfoCtx(ctx, "failed to extract dataset info from path", log.Data{"error": err, "filter_id": filterID, "path": versionURL})
		setStatusCode(req, w, err)
		return
	}

	datasetDimensions, err := f.DatasetClient.GetDimensions(req.Context(), datasetID, edition, version)
	if err != nil {
		log.InfoCtx(ctx, "failed to get dimensions",
			log.Data{"error": err, "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	dimensionIDNameLookup := make(map[string]map[string]string)
	for _, dim := range datasetDimensions.Items {
		idNameLookup := make(map[string]string)
		options, err := f.DatasetClient.GetOptions(req.Context(), datasetID, edition, version, dim.Name)
		if err != nil {
			log.InfoCtx(ctx, "failed to get options from dataset client",
				log.Data{"error": err, "dimension": dim.Name, "dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		for _, opt := range options.Items {
			idNameLookup[opt.Option] = opt.Label
		}
		dimensionIDNameLookup[dim.Name] = idNameLookup
	}

	var dimensions FilterModelDimensions
	for _, dim := range dims {
		var vals []filter.DimensionOption
		vals, err = f.FilterClient.GetDimensionOptions(req.Context(), filterID, dim.Name)
		if err != nil {
			log.InfoCtx(ctx, "failed to get options from filter client", log.Data{"error": err, "filter_id": filterID, "dimension": dim.Name})
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

	dataset, err := f.DatasetClient.Get(req.Context(), datasetID)
	if err != nil {
		log.InfoCtx(ctx, "failed to get dataset", log.Data{"error": err, "dataset_id": datasetID})
		setStatusCode(req, w, err)
		return
	}

	ver, err := f.DatasetClient.GetVersion(req.Context(), datasetID, edition, version)
	if err != nil {
		log.InfoCtx(ctx, "failed to get version", log.Data{"error": err, "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	latestURL, err := url.Parse(dataset.Links.LatestVersion.URL)
	if err != nil {
		log.InfoCtx(ctx, "failed to parse latest version href", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	p := mapper.CreateFilterOverview(req.Context(), dimensions, datasetDimensions.Items, fj, dataset, filterID, datasetID, ver.ReleaseDate)

	if latestURL.Path == versionURL.Path {
		p.Data.IsLatestVersion = true
	}

	p.Data.LatestVersion.DatasetLandingPageURL = latestURL.Path
	p.Data.LatestVersion.FilterJourneyWithLatestJourney = fmt.Sprintf("/filters/%s/use-latest-version", filterID)

	b, err := json.Marshal(p)
	if err != nil {
		log.InfoCtx(ctx, "failed to marshal json", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/filter-overview", b)
	if err != nil {
		log.InfoCtx(ctx, "failed to render", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	w.Write(templateBytes)
}

// FilterOverviewClearAll removes all selected options for all dimensions
func (f *Filter) FilterOverviewClearAll(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]
	ctx := req.Context()

	req = forwardFlorenceTokenIfRequired(req)

	dims, err := f.FilterClient.GetDimensions(req.Context(), filterID)
	if err != nil {
		log.ErrorCtx(ctx, err, nil)
		return
	}

	for _, dim := range dims {
		if err := f.FilterClient.RemoveDimension(req.Context(), filterID, dim.Name); err != nil {
			log.InfoCtx(ctx, "failed to remove dimension", log.Data{"error": err, "filter_id": filterID, "dimension": dim.Name})
			setStatusCode(req, w, err)
			return
		}

		if err := f.FilterClient.AddDimension(req.Context(), filterID, dim.Name); err != nil {
			log.InfoCtx(ctx, "failed to add dimension", log.Data{"error": err, "filter_id": filterID, "dimension": dim.Name})
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
