package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
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
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		ctx := req.Context()

		dims, eTag0, err := f.FilterClient.GetDimensions(req.Context(), userAccessToken, "", collectionID, filterID, filter.QueryParams{})
		if err != nil {
			log.Event(ctx, "failed to get dimensions", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		fj, eTag1, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
		if err != nil {
			log.Event(ctx, "failed to get job state", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		// TODO we might want to retry this handler if eTags don't match
		if eTag0 != eTag1 {
			err := errors.New("inconsistent filter data")
			log.Event(ctx, "data consistency cannot be guaranteed because filter was modified between calls", log.ERROR, log.Error(err),
				log.Data{"filter_id": filterID, "e_tag_0": eTag0, "e_tag_1": eTag1})
			setStatusCode(req, w, err)
			return
		}

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

		datasetDimensions, err := f.DatasetClient.GetVersionDimensions(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Event(ctx, "failed to get dimensions", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		// get selected options from filter API for each dimension and then get the labels from dataset API for each option
		var dimensions FilterModelDimensions
		for _, dim := range dims.Items {
			selVals, eTag2, err := f.FilterClient.GetDimensionOptionsInBatches(req.Context(), userAccessToken, "", collectionID, filterID, dim.Name, f.BatchSize, f.BatchMaxWorkers)
			if err != nil {
				log.Event(ctx, "failed to get options from filter client", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dim.Name})
				setStatusCode(req, w, err)
				return
			}

			// TODO we might want to retry this handler if eTags don't match
			if eTag2 != eTag1 {
				err := errors.New("inconsistent filter data")
				log.Event(ctx, "data consistency cannot be guaranteed because filter was modified between calls", log.ERROR, log.Error(err),
					log.Data{"filter_id": filterID, "e_tag_1": eTag1, "e_tag_2": eTag2})
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
		sort.Sort(dimensions)

		dataset, err := f.DatasetClient.Get(req.Context(), userAccessToken, "", collectionID, datasetID)
		if err != nil {
			log.Event(ctx, "failed to get dataset", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID})
			setStatusCode(req, w, err)
			return
		}

		ver, err := f.DatasetClient.GetVersion(req.Context(), userAccessToken, "", "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Event(ctx, "failed to get version", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		p := mapper.CreateFilterOverview(req, dimensions, datasetDimensions.Items, fj, dataset, filterID, datasetID, ver.ReleaseDate, f.APIRouterVersion, lang)

		editionDetails, err := f.DatasetClient.GetEdition(req.Context(), userAccessToken, "", collectionID, datasetID, edition)
		if err != nil {
			log.Event(ctx, "failed to get edition details", log.ERROR, log.Error(err), log.Data{"dataset": datasetID, "edition": edition})
			setStatusCode(req, w, err)
			return
		}

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

		templateBytes, err := f.Renderer.Do("dataset-filter/filter-overview", b)
		if err != nil {
			log.Event(ctx, "failed to render", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateBytes)
	})

}

// FilterOverviewClearAll removes all selected options for all dimensions
func (f *Filter) FilterOverviewClearAll() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		ctx := req.Context()

		dims, eTag, err := f.FilterClient.GetDimensions(req.Context(), userAccessToken, "", collectionID, filterID, filter.QueryParams{})
		if err != nil {
			log.Event(ctx, "failed to get dimensions", log.ERROR, log.Error(err))
			return
		}

		for _, dim := range dims.Items {
			eTag, err = f.FilterClient.RemoveDimension(req.Context(), userAccessToken, "", collectionID, filterID, dim.Name, eTag)
			if err != nil {
				log.Event(ctx, "failed to remove dimension", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dim.Name})
				setStatusCode(req, w, err)
				return
			}

			eTag, err = f.FilterClient.AddDimension(req.Context(), userAccessToken, "", collectionID, filterID, dim.Name, eTag)
			if err != nil {
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
