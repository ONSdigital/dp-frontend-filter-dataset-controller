package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/data"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/renderer"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/geography"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/listSelector"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/rangeSelector"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// Filter represents the handlers for Filtering
type Filter struct {
	r  renderer.Renderer
	fc FilterClient
	dc DatasetClient
}

// NewFilter creates a new instance of Filter
func NewFilter(r renderer.Renderer, fc FilterClient, dc DatasetClient) *Filter {
	return &Filter{
		r:  r,
		fc: fc,
		dc: dc,
	}
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

	dims, err := f.fc.GetDimensions(filterID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var dimensions []data.Dimension
	for _, dim := range dims {
		var vals data.FilterDimensionValues
		vals, err = f.fc.GetDimensionOptions(filterID, dim.Name)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var values []string
		for _, val := range vals.Items {
			values = append(values, val.Name)
		}

		dimensions = append(dimensions, data.Dimension{
			Name:   dim.Name,
			Values: values,
		})
	}

	filter, err := f.fc.GetJobState(filterID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dataset, err := f.dc.GetDataset(filterID, filter.Edition, filter.Version)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
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

// RangeSelector controls the render of the range selector template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) RangeSelector(w http.ResponseWriter, req *http.Request) {
	p := rangeSelector.Page{
		FilterID: "12345",
		Data: rangeSelector.RangeSelector{
			AddFromList: rangeSelector.Link{
				Label: "Add age range",
				URL:   "/filters/12345/dimensions/age-list",
			},
			NumberOfSelectors: 1,
			AddRange: rangeSelector.Link{
				Label: "Add ages",
				URL:   "/add-to-basket/",
			},
			AddAllInRange: rangeSelector.Link{
				Label: "All ages",
			},
			AddNewRange: rangeSelector.Link{
				URL: "/add-another-range",
			},
			RemoveRange: rangeSelector.Link{
				URL:   "/remove-range",
				Label: "Remove",
			},
			SaveAndReturn: rangeSelector.Link{
				URL: "/filters/12345/dimensions",
			},
			Cancel: rangeSelector.Link{
				URL: "/filters/12345/dimensions",
			},
			FiltersAmount: 2,
			FiltersAdded: []rangeSelector.Filter{
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
			RemoveAll: rangeSelector.Link{
				URL: "/remove-all/",
			},
			RangeData: rangeSelector.Range{
				StartNum:     30,
				EndNum:       90,
				StartLabel:   "Youngest",
				EndLabel:     "Oldest",
				AppendString: "and over",
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

	templateBytes, err := f.r.Do("dataset-filter/range-selector", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}

// ListSelector controls the render of the age selector list template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) ListSelector(w http.ResponseWriter, req *http.Request) {
	p := listSelector.Page{
		FilterID: "12345",
		Data: listSelector.ListSelector{
			AddFromRange: listSelector.Link{
				Label: "add age range",
				URL:   "/filters/12345/dimensions/age-range",
			},
			SaveAndReturn: listSelector.Link{
				URL: "/filters/12345/dimensions",
			},
			AddAllInRange: listSelector.Link{
				Label: "All ages",
			},
			Cancel: listSelector.Link{
				URL: "/filters/12345/dimensions",
			},
			FiltersAdded: []listSelector.Filter{
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
			RemoveAll: listSelector.Link{
				URL: "/remove-all/",
			},
			FiltersAmount: 2,
			RangeData: listSelector.Range{
				StartNum:     30,
				EndNum:       90,
				StartLabel:   "Youngest",
				EndLabel:     "Oldest",
				AppendString: "and over",
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

	templateBytes, err := f.r.Do("dataset-filter/list-selector", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}
