package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/data"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// DimensionSelector controls the render of the range selector template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) DimensionSelector(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]

	if name == "goods-and-services" || name == "CPI" {
		url := fmt.Sprintf("/filters/%s/hierarchies/%s", filterID, name)
		http.Redirect(w, req, url, 301)
	}

	filter, err := f.fc.GetJobState(filterID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	selectedValues, err := f.fc.GetDimensionOptions(filterID, name)
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

	dim, err := f.fc.GetDimension(filterID, name)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	codeID := getCodeIDFromURI(dim.URI)
	if codeID == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	allValues, err := f.clc.GetValues(codeID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	selectorType := req.URL.Query().Get("selectorType")
	if selectorType == "list" {
		f.listSelector(w, req, name, selectedValues, allValues, filter, dataset)
	} else {
		f.rangeSelector(w, req, name, selectedValues, allValues, filter, dataset)
	}
}

func (f *Filter) rangeSelector(w http.ResponseWriter, req *http.Request, name string, selectedValues, allValues data.DimensionValues, filter data.Filter, dataset data.Dataset) {

	p := mapper.CreateRangeSelectorPage(name, selectedValues, allValues, filter, dataset)

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
func (f *Filter) listSelector(w http.ResponseWriter, req *http.Request, name string, selectedValues, allValues data.DimensionValues, filter data.Filter, dataset data.Dataset) {
	p := mapper.CreateListSelectorPage(name, selectedValues, allValues, filter, dataset)

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

// Range represents range labels in the range selector page
type Range struct {
	Start string `schema:"start"`
	End   string `schema:"end"`
}

// AddRange will add a range of values to a filter job
func (f *Filter) AddRange(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]

	if err := req.ParseForm(); err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, ok := req.Form["All times"]; ok {
		f.AddAll(w, req)
		return
	}

	var r Range

	if err := f.val.Validate(req, &r); err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	values, labelIDMap, err := f.getDimensionValues(filterID, name)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if name == "time" {
		dats, err := dates.ConvertToReadable(values)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		start, err := time.Parse("01 January 2006", fmt.Sprintf("01 %s", r.Start))
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
			http.Redirect(w, req, redirectURL, 301)
		}

		end, err := time.Parse("01 January 2006", fmt.Sprintf("01 %s", r.End))
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
			http.Redirect(w, req, redirectURL, 301)
		}

		if end.Before(start) {
			log.Info("end date before start date", log.Data{"start": start, "end": end})
			w.WriteHeader(http.StatusInternalServerError)
			redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
			http.Redirect(w, req, redirectURL, 301)
		}

		dats = dates.Sort(dats)
		values = dates.ConvertToCoded(dats)
		for i, dat := range dats {
			if dat.Equal(start) || dat.After(start) && dat.Before(end) || dat.Equal(end) {
				if err := f.fc.AddDimensionValue(filterID, name, labelIDMap[values[i]]); err != nil {
					log.TraceR(req, err.Error(), nil)
					continue
				}
			}
		}
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
	http.Redirect(w, req, redirectURL, 301)
}

// AddAll ...
func (f *Filter) AddAll(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]

	vals, err := f.fc.GetDimensionOptions(filterID, name)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
		http.Redirect(w, req, redirectURL, 301)
	}

	var wg sync.WaitGroup
	for _, val := range vals.Items {
		wg.Add(1)
		go func(val data.DimensionValueItem) {
			if err := f.fc.AddDimensionValue(filterID, name, val.ID); err != nil {
				log.ErrorR(req, err, nil)
			}
			wg.Done()
		}(val)
	}
	wg.Wait()

	redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
	http.Redirect(w, req, redirectURL, 301)
}

// AddList ...
func (f *Filter) AddList(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]

	if err := req.ParseForm(); err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, ok := req.Form["All times"]; ok {
		f.AddAll(w, req)
		return
	}

	values, labelIDMap, err := f.getDimensionValues(filterID, name)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if name == "time" {
		dats, err := dates.ConvertToReadable(values)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		values = dates.ConvertToCoded(dats)

		for k := range req.Form {
			if k == ":uri" {
				continue
			}

			val, err := time.Parse("01 January 2006", fmt.Sprintf("01 %s", k))
			if err != nil {
				log.ErrorR(req, err, nil)
				w.WriteHeader(http.StatusInternalServerError)
				redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
				http.Redirect(w, req, redirectURL, 301)
			}

			for i, dat := range dats {
				if dat.Equal(val) {
					if err := f.fc.AddDimensionValue(filterID, name, labelIDMap[values[i]]); err != nil {
						log.TraceR(req, err.Error(), nil)
						continue
					}
				}
			}
		}
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s?selectorType=list", filterID, name)
	http.Redirect(w, req, redirectURL, 301)
}

func (f *Filter) getDimensionValues(filterID, name string) (values []string, labelIDMap map[string]string, err error) {
	dim, err := f.fc.GetDimension(filterID, name)
	if err != nil {
		return
	}

	codeID := getCodeIDFromURI(dim.URI)
	if codeID == "" {
		err = errors.New("missing code id from uri")
		return
	}

	allValues, err := f.clc.GetValues(codeID)
	if err != nil {
		return
	}

	labelIDMap = make(map[string]string)
	for _, val := range allValues.Items {
		values = append(values, val.Label)
		labelIDMap[val.Label] = val.ID
	}

	return
}

// DimensionRemoveAll ...
func (f *Filter) DimensionRemoveAll(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]

	vals, err := f.fc.GetDimensionOptions(filterID, name)
	if err != nil {
		log.ErrorR(req, err, nil)
		redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)
		http.Redirect(w, req, redirectURL, 301)
	}

	var wg sync.WaitGroup
	for _, val := range vals.Items {
		wg.Add(1)
		go func(val data.DimensionValueItem) {
			if err := f.fc.RemoveDimensionValue(filterID, name, val.ID); err != nil {
				log.ErrorR(req, err, nil)
			}
			wg.Done()
		}(val)
	}
	wg.Wait()

	redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
	http.Redirect(w, req, redirectURL, 301)
}

// DimensionRemoveOne ...
func (f *Filter) DimensionRemoveOne(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	option := vars["option"]

	if err := f.fc.RemoveDimensionValue(filterID, name, option); err != nil {
		log.ErrorR(req, err, nil)
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
	http.Redirect(w, req, redirectURL, 301)
}

func getCodeIDFromURI(uri string) string {
	codeReg := regexp.MustCompile(`^\/code-lists\/(.+)\/codes$`)
	subs := codeReg.FindStringSubmatch(uri)

	if len(subs) == 2 {
		return subs[1]
	}

	log.Info("could not extract codeID from uri", nil)
	return ""
}
