package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

var waited = make(map[string]bool)

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

	dataset := dataset.Model{
		ID:          "849209",
		ReleaseDate: "17 January 2017",
		NextRelease: "17 February 2017",
		Contact: dataset.Contact{
			Name:      "Matt Rout",
			Telephone: "07984593234",
			Email:     "matt@gmail.com",
		},
		Title: "Small Area Population Estimates",
	}

	filter := filter.Model{
		FilterID: vars["filterID"],
		Edition:  "12345",
		Dataset:  "849209",
		Version:  "2017",
		Downloads: map[string]filter.Download{
			"csv": {
				Size: "362783",
				URL:  "/",
			},
			"xls": {
				Size: "373929",
				URL:  "/",
			},
		},
	}

	p := mapper.CreatePreviewPage(dimensions, filter, dataset, vars["filterID"])

	if _, ok := waited[filterID]; !ok {
		p.IsContentLoaded = false
	}
	waited[filterID] = true

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
