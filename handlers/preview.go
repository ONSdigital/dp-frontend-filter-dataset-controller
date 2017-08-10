package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/data"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// PreviewPage controls the rendering of the preview and download page
func (f *Filter) PreviewPage(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	dimensions := []data.Dimension{
		{
			Name:   "Geography",
			Values: []string{"Wales", "Cardiff", "Caerphilly", "Newport", "Pontypridd", "Merthyr Tydfil"},
		},
		{
			Name:   "Population",
			Values: []string{"30000000", "284384", "37238", "428219", "32161", "281994"},
		},
	}

	dataset := data.Dataset{
		ID:          "849209",
		ReleaseDate: "17 January 2017",
		NextRelease: "17 February 2017",
		Contact: data.Contact{
			Name:      "Matt Rout",
			Telephone: "07984593234",
			Email:     "matt@gmail.com",
		},
		Title: "Small Area Population Estimates",
	}

	filter := data.Filter{
		FilterID: vars["filterID"],
		Edition:  "12345",
		Dataset:  "849209",
		Version:  "2017",
		Downloads: map[string]data.Download{
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

	body, err := json.Marshal(p)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := f.r.Do("dataset-filter/preview-page", body)
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
