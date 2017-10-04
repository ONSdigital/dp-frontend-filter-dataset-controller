package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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

	dims, err := f.FilterClient.GetDimensions(filterID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fj, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	datasetDimensions, err := f.DatasetClient.GetDimensions(datasetID, edition, version)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	idNameLookup := make(map[string]string)
	for _, dim := range datasetDimensions.Items {
		options, err := f.DatasetClient.GetOptions(datasetID, edition, version, dim.ID)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, opt := range options.Items {
			idNameLookup[opt.Option] = opt.Label
		}
	}

	var dimensions []filter.ModelDimension
	for _, dim := range dims {
		var vals []filter.DimensionOption
		vals, err = f.FilterClient.GetDimensionOptions(filterID, dim.Name)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var values []string
		for _, val := range vals {
			values = append(values, idNameLookup[val.Option])
		}

		dimensions = append(dimensions, filter.ModelDimension{
			Name:   dim.Name,
			Values: values,
		})
	}

	dataset, err := f.DatasetClient.Get(datasetID)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ver, err := f.DatasetClient.GetVersion(datasetID, edition, version)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	latestURL, err := url.Parse(dataset.Links.LatestVersion.URL)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	p := mapper.CreateFilterOverview(dimensions, fj, dataset, filterID, datasetID, ver.ReleaseDate)

	if latestURL.Path == versionURL.Path {
		p.Data.IsLatestVersion = true
	}

	p.Data.LatestVersion.DatasetLandingPageURL = versionURL.Path
	p.Data.LatestVersion.FilterJourneyWithLatestJourney = fmt.Sprintf("/filters/%s/use-latest-version", filterID)

	b, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/filter-overview", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
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
			log.ErrorR(req, err, nil)
			return
		}

		if err := f.FilterClient.AddDimension(filterID, dim.Name); err != nil {
			log.ErrorR(req, err, nil)
			return
		}
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)

	http.Redirect(w, req, redirectURL, 302)
}
