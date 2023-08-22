package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"unicode"

	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	dphandlers "github.com/ONSdigital/dp-net/v2/handlers"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// FilterOverview controls the render of the filter overview template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) FilterOverview() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		ctx := req.Context()

		hasUnsetDimensions := req.URL.Query().Get("hasUnsetDimensions")

		dims, eTag0, err := f.FilterClient.GetDimensions(req.Context(), userAccessToken, "", collectionID, filterID, nil)
		if err != nil {
			log.Error(ctx, "failed to get dimensions", err, log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		fj, eTag1, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
		if err != nil {
			log.Error(ctx, "failed to get job state", err, log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		// The user might want to retry this handler if eTags don't match
		if eTag0 != eTag1 {
			conflictErr := errors.New("inconsistent filter data")
			log.Error(ctx, "data consistency cannot be guaranteed because filter was modified between calls", conflictErr,
				log.Data{"filter_id": filterID, "e_tag_0": eTag0, "e_tag_1": eTag1})
			setStatusCode(req, w, conflictErr)
			return
		}

		versionURL, err := url.Parse(fj.Links.Version.HRef)
		if err != nil {
			log.Error(ctx, "failed to parse version href", err, log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}
		versionPath := strings.TrimPrefix(versionURL.Path, f.APIRouterVersion)

		datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
		if err != nil {
			log.Error(ctx, "failed to extract dataset info from path", err, log.Data{"filter_id": filterID, "path": versionPath})
			setStatusCode(req, w, err)
			return
		}

		datasetDimensions, err := f.DatasetClient.GetVersionDimensions(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Error(ctx, "failed to get dimensions", err, log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		// get selected options from filter API for each dimension and then get the labels from dataset API for each option
		var dimensions FilterModelDimensions
		for i := range dims.Items {
			selVals, eTag2, oErr := f.FilterClient.GetDimensionOptionsInBatches(req.Context(), userAccessToken, "", collectionID, filterID, dims.Items[i].Name, f.BatchSize, f.BatchMaxWorkers)
			if oErr != nil {
				log.Error(ctx, "failed to get options from filter client", oErr, log.Data{"filter_id": filterID, "dimension": dims.Items[i].Name})
				setStatusCode(req, w, oErr)
				return
			}

			// The user might want to retry this handler if eTags don't match
			if eTag2 != eTag1 {
				conflictErr := errors.New("inconsistent filter data")
				log.Error(ctx, "data consistency cannot be guaranteed because filter was modified between calls", conflictErr,
					log.Data{"filter_id": filterID, "e_tag_1": eTag1, "e_tag_2": eTag2})
				setStatusCode(req, w, conflictErr)
				return
			}

			selValsLabelMap, oErr := f.getIDNameLookupFromDatasetAPI(ctx, userAccessToken, collectionID, datasetID, edition, version, dims.Items[i].Name, selVals)
			if oErr != nil {
				log.Error(ctx, "failed to get options from dataset client", oErr, log.Data{"dimension": dims.Items[i].Name, "dataset_id": datasetID, "edition": edition, "version": version})
				setStatusCode(req, w, oErr)
				return
			}

			labels := []string{}
			for _, label := range selValsLabelMap {
				labels = append(labels, label)
			}

			dimensions = append(dimensions, filter.ModelDimension{
				Name:   dims.Items[i].Name,
				Values: labels,
			})
		}
		sort.Sort(dimensions)

		dataset, err := f.DatasetClient.Get(req.Context(), userAccessToken, "", collectionID, datasetID)
		if err != nil {
			log.Error(ctx, "failed to get dataset", err, log.Data{"dataset_id": datasetID})
			setStatusCode(req, w, err)
			return
		}

		homepageContent, err := f.ZebedeeClient.GetHomepageContent(ctx, userAccessToken, collectionID, lang, "/")
		if err != nil {
			log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
		}

		bp := f.RenderClient.NewBasePageModel()
		p := mapper.CreateFilterOverview(req, bp, dimensions, datasetDimensions.Items, fj, dataset, filterID, datasetID, f.APIRouterVersion, lang, homepageContent.ServiceMessage, homepageContent.EmergencyBanner)

		editionDetails, err := f.DatasetClient.GetEdition(req.Context(), userAccessToken, "", collectionID, datasetID, edition)
		if err != nil {
			log.Error(ctx, "failed to get edition details", err, log.Data{"dataset": datasetID, "edition": edition})
			setStatusCode(req, w, err)
			return
		}

		latestVersionInEditionPath := fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", datasetID, edition, editionDetails.Links.LatestVersion.ID)
		if latestVersionInEditionPath == versionPath {
			p.Data.IsLatestVersion = true
		}

		p.Data.LatestVersion.DatasetLandingPageURL = latestVersionInEditionPath
		p.Data.LatestVersion.FilterJourneyWithLatestJourney = fmt.Sprintf("/filters/%s/use-latest-version", filterID)

		if hasUnsetDimensions == "true" {
			p.Data.HasUnsetDimensions = true
		}

		f.RenderClient.BuildPage(w, p, "filter-overview")
	})
}

// FilterOverviewClearAll removes all selected options for all dimensions
func (f *Filter) FilterOverviewClearAll() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		ctx := req.Context()

		dims, eTag, err := f.FilterClient.GetDimensions(req.Context(), userAccessToken, "", collectionID, filterID, nil)
		if err != nil {
			log.Error(ctx, "failed to get dimensions", err)
			return
		}

		for i := range dims.Items {
			eTag, err = f.FilterClient.RemoveDimension(req.Context(), userAccessToken, "", collectionID, filterID, dims.Items[i].Name, eTag)
			if err != nil {
				log.Error(ctx, "failed to remove dimension", err, log.Data{"filter_id": filterID, "dimension": dims.Items[i].Name})
				setStatusCode(req, w, err)
				return
			}

			eTag, err = f.FilterClient.AddDimension(req.Context(), userAccessToken, "", collectionID, filterID, dims.Items[i].Name, eTag)
			if err != nil {
				log.Error(ctx, "failed to add dimension", err, log.Data{"filter_id": filterID, "dimension": dims.Items[i].Name})
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
