package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// PreviewPage controls the rendering of the preview and download page
func (f *Filter) PreviewPage(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	dimensions := []filter.ModelDimension{
		{
			Name:   "Time",
			Values: []string{"January 2017", "January 2016", "January 2015", "January 2014", "January 2013", "January 2012"},
		},
		{
			Name:   "Goods and Services",
			Values: []string{"Clothing", "Education", "Aviation", "12", "11", "10"},
		},
		{
			Name:   "CPI",
			Values: []string{"0.23", "0.48", "0.593", "0.38", "0.349", "0.389"},
		},
	}

	fil, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if fil.State == "created" {
		fil.State = "submitted"

		f.FilterClient.RemoveDimension(filterID, "time")
		f.FilterClient.RemoveDimension(filterID, "goods-and-services")

		f.FilterClient.AddDimension(filterID, "Time")
		f.FilterClient.AddDimension(filterID, "Geography")
		f.FilterClient.AddDimension(filterID, "Aggregate")

		f.FilterClient.AddDimensionValues(filterID, "Time", []string{"Jan-96", "Feb-96", "Feb-97", "Feb-98", "Mar-02", "Jun-08", "Dec-10", "Nov-11"})
		f.FilterClient.AddDimensionValue(filterID, "Geography", "K02000001")
		f.FilterClient.AddDimensionValues(filterID, "Aggregate", []string{"cpi1dim1G10100", "cpi1dim1G20100"})

		f.FilterClient.UpdateJob(fil)
	}

	// versionURL := filter.
	versionURL := "/datasets/95c4669b-3ae9-4ba7-b690-87e890a1c67c/editions/2016/versions/1"
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
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

	p := mapper.CreatePreviewPage(dimensions, fil, dataset, filterID, datasetID, ver.ReleaseDate)

	if fil.State != "completed" {
		p.IsContentLoaded = false
	}

	body, err := json.Marshal(p)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := f.Renderer.Do("dataset-filter/preview-page", body)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(b); err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
