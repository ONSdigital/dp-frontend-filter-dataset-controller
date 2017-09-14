package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/geography"
	"github.com/ONSdigital/go-ns/log"
)

// Geography controls the rendering of the geography page. TODO: this handler may
// be removed if this is better placed under hierarchy
func (f *Filter) Geography(w http.ResponseWriter, req *http.Request) {
	p := geography.Page{
		FilterID: "12345",
		Data: geography.Geography{
			SaveAndReturn: geography.Link{
				URL: "/filters/12345/dimensions",
			},
			Cancel: geography.Link{
				URL: "/filters/12345/dimensions",
			},
			FiltersAmount: 2,
			FiltersAdded: []geography.Filter{
				{
					RemoveURL: "/remove-this/",
					Label:     "All ages",
				},
				{
					RemoveURL: "/remove-this-2/",
					Label:     "43",
				},
				{
					RemoveURL: "/remove-this-3/",
					Label:     "18",
				},
			},
			FilterList: []geography.List{
				{
					Location: "United Kingdom",
				},
				{
					Location: "England",
					SubNum:   10,
					SubType:  "Regions",
					SubURL:   "/regions/",
				},
				{
					Location: "Wales",
					SubNum:   5,
					SubType:  "Regions",
					SubURL:   "/regions/",
				},
			},
			RemoveAll: geography.Link{
				URL: "/remove-all/",
			},
			AddAllFilters: geography.AddAll{
				URL:    "/add-all/",
				Amount: 3,
			},
			GoBack: geography.Link{
				URL: "/back/",
			},
			Parent: "Wales: Counties",
		},
	}

	p.Breadcrumb = []model.TaxonomyNode{
		{
			Title: "Title of dataset",
			URI:   "/",
		},
		{
			Title: "Filter this dataset",
			URI:   "/",
		},
	}

	p.SearchDisabled = true

	b, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/geography", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}
