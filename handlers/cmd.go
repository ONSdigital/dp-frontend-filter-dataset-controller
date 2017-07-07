package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/renderer"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/ageSelectorList"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/ageSelectorRange"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/filterOverview"
	"github.com/ONSdigital/dp-frontend-models/model/datasetpages/finishPage"
	"github.com/ONSdigital/dp-frontend-models/model/datasetpages/middlePage"
	"github.com/ONSdigital/dp-frontend-models/model/datasetpages/startPage"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// CMD represents the handlers for CMD
type CMD struct {
	r renderer.Renderer
}

// NewCMD creates a new instance of CMD
func NewCMD(r renderer.Renderer) *CMD {
	return &CMD{r: r}
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

// Landing handles the controller functionality for the landing page
func (c *CMD) Landing(w http.ResponseWriter, req *http.Request) {
	var p startPage.Page

	// Needs to be populated from API - this is stubbed data
	p.Metadata.Footer = getStubbedMetadataFooter()
	p.SearchDisabled = true

	pBytes, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := c.r.Do("dataset/startpage", pBytes)
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

// CreateJobID controls the creating of a job idea when a new user journey is
// requested
func (c *CMD) CreateJobID(w http.ResponseWriter, req *http.Request) {
	// TODO: This is a stubbed job id - replace with real job id from api once
	// code has been written
	jobID := rand.Intn(100000000)
	jid := strconv.Itoa(jobID)

	log.Trace("created job id", log.Data{"job_id": jid})
	http.Redirect(w, req, "/jobs/"+jid, 301)
}

// Middle controls the rendering of a "middle" cmd page - this will be replaced
// by other handlers when further pages are defined by UX
func (c *CMD) Middle(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	var p middlePage.Page

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

	b, err := c.r.Do("dataset/middlepage", body)
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

// FilterOverview controls the render of the filter overview template
// Contains stubbed data for now - page to be populated by the API
func (c *CMD) FilterOverview(w http.ResponseWriter, req *http.Request) {

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
						URL:   "/jobs/12345/dimensions/age",
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

	// p.Breadcrumb = []model.TaxonomyNode{
	// 	{
	// 		Title: "Title of dataset",
	// 		URI:   "/",
	// 	},
	// 	{
	// 		Title: "Filter this dataset",
	// 		URI:   "/",
	// 	},
	// }
	//
	// p.Metadata.Footer = getStubbedMetadataFooter()

	b, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := c.r.Do("cmd/filter-overview", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}

// ageSelectorRange controls the render of the age selector template
// Contains stubbed data for now - page to be populated by the API
func (c *CMD) AgeSelectorRange(w http.ResponseWriter, req *http.Request) {
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
				URL: "/save/",
			},
			Cancel: ageSelectorRange.Link{
				URL: "/cancel/",
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

	p.SearchDisabled = true

	b, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := c.r.Do("cmd/age-selector-range", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}

// ageSelectorList controls the render of the age selector list template
// Contains stubbed data for now - page to be populated by the API
func (c *CMD) AgeSelectorList(w http.ResponseWriter, req *http.Request) {
	var p ageSelectorList.Page

	p.SearchDisabled = true

	b, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := c.r.Do("cmd/age-selector-list", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}

// PreviewAndDownload will control the rendering of the preview and download page
func (c *CMD) PreviewAndDownload(w http.ResponseWriter, req *http.Request) {
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

	b, err := c.r.Do("dataset/finishpage", pBytes)
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
