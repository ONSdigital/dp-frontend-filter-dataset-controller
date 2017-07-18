package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/data"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/renderer"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/ageSelectorList"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/ageSelectorRange"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/geography"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/previewPage"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// Filter represents the handlers for Filtering
type Filter struct {
	r renderer.Renderer
}

// NewFilter creates a new instance of Filter
func NewFilter(r renderer.Renderer) *Filter {
	return &Filter{r: r}
}

func getStubbedMetadataFooter() model.Footer {
	return model.Footer{
		Enabled:     true,
		Contact:     "Matt Rout",
		ReleaseDate: "11 November 2016",
		NextRelease: "11 November 2017",
		DatasetID:   "MR",
	}
}

// PreviewPage controls the rendering of the preview and download page
func (f *Filter) PreviewPage(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	var p previewPage.Page

	// Needs to be populated from API - this is stubbed data
	p.Metadata.Footer = getStubbedMetadataFooter()
	p.SearchDisabled = true
	p.Data.FilterID = vars["filterID"]

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

	body, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := f.r.Do("dataset-filter/preview-page", body)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(b); err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Geography ...
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

	p.Metadata.Footer = getStubbedMetadataFooter()

	b, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.r.Do("dataset-filter/geography", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}

// FilterOverview controls the render of the filter overview template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) FilterOverview(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	dimensions := []data.Dimension{
		{
			Name:   "year",
			Values: []string{"2014"},
		},
		{
			Name:   "geography",
			Values: []string{"England and Wales, Bristol"},
		},
		{
			Name:   "sex",
			Values: []string{"All persons"},
		},
		{
			Name:   "age-range",
			Values: []string{"0 - 92", "2 - 18", "18 - 65"},
		},
	}

	filter := data.Filter{
		FilterID: filterID,
		Edition:  "12345",
		Dataset:  "849209",
		Version:  "2017",
	}

	dataset := data.Dataset{
		ID:          "849209",
		ReleaseDate: "17 January 2017",
		Contact: data.Contact{
			Name:      "Matt Rout",
			Telephone: "07984593234",
			Email:     "matt@gmail.com",
		},
		Title: "Small Area Population Estimates",
	}

	p := mapper.CreateFilterOverview(dimensions, filter, dataset, filterID)

	b, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.r.Do("dataset-filter/filter-overview", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}

// AgeSelectorRange controls the render of the age selector template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) AgeSelectorRange(w http.ResponseWriter, req *http.Request) {
	p := ageSelectorRange.Page{
		FilterID: "12345",
		Data: ageSelectorRange.AgeSelectorRange{
			AddFromList: ageSelectorRange.Link{
				URL: "/filters/12345/dimensions/age-list",
			},
			NumberOfSelectors: 1,
			AddAges: ageSelectorRange.Link{
				Label: "Add ages",
				URL:   "/add-to-basket/",
			},
			AddNewRange: ageSelectorRange.Link{
				URL: "/add-another-range",
			},
			RemoveRange: ageSelectorRange.Link{
				URL:   "/remove-range",
				Label: "Remove",
			},
			SaveAndReturn: ageSelectorRange.Link{
				URL: "/filters/12345/dimensions",
			},
			Cancel: ageSelectorRange.Link{
				URL: "/filters/12345/dimensions",
			},
			FiltersAmount: 2,
			FiltersAdded: []ageSelectorRange.Filter{
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
			RemoveAll: ageSelectorRange.Link{
				URL: "/remove-all/",
			},
			AgeRange: ageSelectorRange.Range{
				StartNum: 30,
				EndNum:   90,
			},
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

	p.Metadata.Footer = getStubbedMetadataFooter()

	b, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.r.Do("dataset-filter/age-selector-range", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}

// AgeSelectorList controls the render of the age selector list template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) AgeSelectorList(w http.ResponseWriter, req *http.Request) {
	p := ageSelectorList.Page{
		FilterID: "12345",
		Data: ageSelectorList.AgeSelectorList{
			AddFromRange: ageSelectorList.Link{
				URL: "/filters/12345/dimensions/age-range",
			},
			SaveAndReturn: ageSelectorList.Link{
				URL: "/filters/12345/dimensions",
			},
			Cancel: ageSelectorList.Link{
				URL: "/filters/12345/dimensions",
			},
			FiltersAdded: []ageSelectorList.Filter{
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
			RemoveAll: ageSelectorList.Link{
				URL: "/remove-all/",
			},
			FiltersAmount: 2,
			AgeRange: ageSelectorList.Range{
				StartNum: 30,
				EndNum:   90,
			},
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

	p.Metadata.Footer = getStubbedMetadataFooter()

	b, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.r.Do("dataset-filter/age-selector-list", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}
