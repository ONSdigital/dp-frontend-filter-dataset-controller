package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	dphandlers "github.com/ONSdigital/dp-net/handlers"

	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

// FilterOverview controls the render of the filter overview template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) FilterOverview() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {

		var tGetDimensions, tGetJobState, tGetVersionDimensions, tGetAllOptionsAndLookups, tGetDataset, tGetVersion, tGetEdition, tRender time.Duration
		t0 := time.Now()

		logTime := func() {
			log.Event(nil, "+++ PERFORMANCE TEST", log.Data{
				"method":                          "filter-overview.FilterOverview",
				"whole":                           fmtDuration(time.Since(t0)),
				"filter_get_dimensions":           fmtDuration(tGetDimensions),
				"filter_get_job_state":            fmtDuration(tGetJobState),
				"dataset_get_version_dimensions":  fmtDuration(tGetVersionDimensions),
				"get_get_all_options_and_lookups": fmtDuration(tGetAllOptionsAndLookups),
				"dataset_get_dataset":             fmtDuration(tGetDataset),
				"dataset_get_version":             fmtDuration(tGetVersion),
				"dataset_get_edition":             fmtDuration(tGetEdition),
				"render":                          fmtDuration(tRender),
			})
		}

		defer logTime()

		vars := mux.Vars(req)
		filterID := vars["filterID"]
		ctx := req.Context()

		t := time.Now()
		dims, err := f.FilterClient.GetDimensions(req.Context(), userAccessToken, "", collectionID, filterID, filter.QueryParams{})
		if err != nil {
			log.Event(ctx, "failed to get dimensions", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}
		tGetDimensions = time.Since(t)

		t = time.Now()
		fj, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
		if err != nil {
			log.Event(ctx, "failed to get job state", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}
		tGetJobState = time.Since(t)

		versionURL, err := url.Parse(fj.Links.Version.HRef)
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

		t = time.Now()
		datasetDimensions, err := f.DatasetClient.GetVersionDimensions(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Event(ctx, "failed to get dimensions", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}
		tGetVersionDimensions = time.Since(t)

		t1 := time.Now()
		// get selected options from filter API for each dimension and then get the labels from dataset API for each option
		var dimensions FilterModelDimensions
		for _, dim := range dims.Items {
			selVals, err := f.FilterClient.GetDimensionOptionsInBatches(req.Context(), userAccessToken, "", collectionID, filterID, dim.Name, f.BatchSize, f.BatchMaxWorkers)
			if err != nil {
				log.Event(ctx, "failed to get options from filter client", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dim.Name})
				setStatusCode(req, w, err)
				return
			}

			selValsLabelMap, err := f.getIDNameLookupFromDatasetAPI(ctx, userAccessToken, collectionID, datasetID, edition, version, dim.Name, selVals)
			if err != nil {
				log.Event(ctx, "failed to get options from dataset client", log.ERROR, log.Error(err), log.Data{"dimension": dim.Name, "dataset_id": datasetID, "edition": edition, "version": version})
				setStatusCode(req, w, err)
				return
			}

			labels := []string{}
			for _, label := range selValsLabelMap {
				labels = append(labels, label)
			}

			dimensions = append(dimensions, filter.ModelDimension{
				Name:   dim.Name,
				Values: labels,
			})
		}
		tGetAllOptionsAndLookups = time.Since(t1)
		sort.Sort(dimensions)

		t = time.Now()
		dataset, err := f.DatasetClient.Get(req.Context(), userAccessToken, "", collectionID, datasetID)
		if err != nil {
			log.Event(ctx, "failed to get dataset", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID})
			setStatusCode(req, w, err)
			return
		}
		tGetDataset = time.Since(t)

		t = time.Now()
		ver, err := f.DatasetClient.GetVersion(req.Context(), userAccessToken, "", "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Event(ctx, "failed to get version", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}
		tGetVersion = time.Since(t)

		p := mapper.CreateFilterOverview(req, dimensions, datasetDimensions.Items, fj, dataset, filterID, datasetID, ver.ReleaseDate, f.APIRouterVersion, lang)

		t = time.Now()
		editionDetails, err := f.DatasetClient.GetEdition(req.Context(), userAccessToken, "", collectionID, datasetID, edition)
		if err != nil {
			log.Event(ctx, "failed to get edition details", log.ERROR, log.Error(err), log.Data{"dataset": datasetID, "edition": edition})
			setStatusCode(req, w, err)
			return
		}
		tGetEdition = time.Since(t)

		latestVersionInEditionPath := fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", datasetID, edition, editionDetails.Links.LatestVersion.ID)
		if latestVersionInEditionPath == versionPath {
			p.Data.IsLatestVersion = true
		}

		p.Data.LatestVersion.DatasetLandingPageURL = latestVersionInEditionPath
		p.Data.LatestVersion.FilterJourneyWithLatestJourney = fmt.Sprintf("/filters/%s/use-latest-version", filterID)

		b, err := json.Marshal(p)
		if err != nil {
			log.Event(ctx, "failed to marshal json", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		t = time.Now()
		templateBytes, err := f.Renderer.Do("dataset-filter/filter-overview", b)
		if err != nil {
			log.Event(ctx, "failed to render", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}
		tRender = time.Since(t)

		w.Write(templateBytes)
	})

}

// FilterOverviewClearAll removes all selected options for all dimensions
func (f *Filter) FilterOverviewClearAll() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		ctx := req.Context()

		dims, err := f.FilterClient.GetDimensions(req.Context(), userAccessToken, "", collectionID, filterID, filter.QueryParams{})
		if err != nil {
			log.Event(ctx, "failed to get dimensions", log.ERROR, log.Error(err))
			return
		}

		for _, dim := range dims.Items {
			if err := f.FilterClient.RemoveDimension(req.Context(), userAccessToken, "", collectionID, filterID, dim.Name); err != nil {
				log.Event(ctx, "failed to remove dimension", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dim.Name})
				setStatusCode(req, w, err)
				return
			}

			if err := f.FilterClient.AddDimension(req.Context(), userAccessToken, "", collectionID, filterID, dim.Name); err != nil {
				log.Event(ctx, "failed to add dimension", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dim.Name})
				setStatusCode(req, w, err)
				return
			}
		}

		redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)

		http.Redirect(w, req, redirectURL, 302)
	})
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
