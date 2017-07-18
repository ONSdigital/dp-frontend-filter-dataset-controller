package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/renderer"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/ageSelectorList"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/ageSelectorRange"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/filterOverview"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/geography"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/previewPage"
	"github.com/ONSdigital/dp-frontend-models/model/datasetpages/finishPage"
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
	p.Data.JobID = vars["jobID"]

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
		JobID: "12345",
		Data: geography.Geography{
			SaveAndReturn: geography.Link{
				URL: "/jobs/12345/dimensions",
			},
			Cancel: geography.Link{
				URL: "/jobs/12345/dimensions",
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

	p := filterOverview.Page{
		JobID: "12345",
		Data: filterOverview.FilterOverview{
			Dimensions: []filterOverview.Dimension{
				{
					Filter:          "Year",
					AddedCategories: "2014",
				},
				{
					Filter:          "Geographic Areas",
					AddedCategories: "(1) All persons",
					Link: filterOverview.Link{
						URL:   "/jobs/12345/dimensions/geography",
						Label: "Please add",
					},
				},
				{
					Filter:          "Sex",
					AddedCategories: "(1) All Persons",
					Link: filterOverview.Link{
						URL:   "/jobs/12345/dimensions/sex",
						Label: "Filter",
					},
				},
				{
					Filter:          "Age",
					AddedCategories: "(1) 0 - 92",
					Link: filterOverview.Link{
						URL:   "/jobs/12345/dimensions/age-range",
						Label: "Filter",
					},
				},
			},
			PreviewAndDownload: filterOverview.Link{
				URL: "/jobs/12345",
			},
			Cancel: filterOverview.Link{
				URL: "https://ons.gov.uk",
			},
		},
	}

	p.SearchDisabled = true

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

	p.Metadata.Footer = getStubbedMetadataFooter()

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
		JobID: "12345",
		Data: ageSelectorRange.AgeSelectorRange{
			AddFromList: ageSelectorRange.Link{
				URL: "/jobs/12345/dimensions/age-list",
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
				URL: "/jobs/12345/dimensions",
			},
			Cancel: ageSelectorRange.Link{
				URL: "/jobs/12345/dimensions",
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
		JobID: "12345",
		Data: ageSelectorList.AgeSelectorList{
			AddFromRange: ageSelectorList.Link{
				URL: "/jobs/12345/dimensions/age-range",
			},
			SaveAndReturn: ageSelectorList.Link{
				URL: "/jobs/12345/dimensions",
			},
			Cancel: ageSelectorList.Link{
				URL: "/jobs/12345/dimensions",
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

// PreviewAndDownload will control the rendering of the preview and download page
func (f *Filter) PreviewAndDownload(w http.ResponseWriter, req *http.Request) {
	var p finishPage.Page

	// Needs to be populated from API - this is stubbed data
	p.Metadata.Footer = getStubbedMetadataFooter()
	p.SearchDisabled = true

	pBytes, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := f.r.Do("dataset/finishpage", pBytes)
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
