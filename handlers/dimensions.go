package handlers

import (
	"encoding/json"
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
		http.Redirect(w, req, url, 302)
	}

	filter, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	selectedValues, err := f.FilterClient.GetDimensionOptions(filterID, name)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dataset, err := f.DatasetClient.GetDataset(filterID, "2016", "v1") // TODO: this will need to be replaced with the real edition/version when it becomes available
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dim, err := f.FilterClient.GetDimension(filterID, name)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debug("dimension", log.Data{"dimension": dim})

	/*codeID := getCodeIDFromURI(dim.URI) // TODO: uncomment this code when the codelist api is updated with real urls
	if codeID == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}*/

	codeID := "64d384f1-ea3b-445c-8fb8-aa453f96e58a" // TODO: remove this with real codeids when available
	allValues, err := f.CodeListClient.GetValues(codeID)
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

func (f *Filter) rangeSelector(w http.ResponseWriter, req *http.Request, name string, selectedValues []data.DimensionOption, allValues data.DimensionValues, filter data.Filter, dataset data.Dataset) {

	p := mapper.CreateRangeSelectorPage(name, selectedValues, allValues, filter, dataset)

	b, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/range-selector", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}

// ListSelector controls the render of the age selector list template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) listSelector(w http.ResponseWriter, req *http.Request, name string, selectedValues []data.DimensionOption, allValues data.DimensionValues, filter data.Filter, dataset data.Dataset) {
	p := mapper.CreateListSelectorPage(name, selectedValues, allValues, filter, dataset)

	b, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/list-selector", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}

// Range represents range labels in the range selector page
type Range struct {
	Start          string `schema:"start"`
	End            string `schema:"end"`
	SaveAndReturn  string `schema:"save-and-return"`
	AddAll         string `schema:"add-all"`
	AddAllRange    string `schema:"add-all-range"`
	RemoveAllRange string `schema:"remove-all-range"`
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

	var r Range

	if err := f.val.Validate(req, &r); err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var redirectURL string
	if len(r.SaveAndReturn) > 0 {
		redirectURL = fmt.Sprintf("/filters/%s/dimensions", filterID)
	} else {
		redirectURL = fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
	}

	if len(r.AddAll) > 0 {
		f.addAll(w, req, redirectURL)
		return
	}

	if len(req.Form["add-all-range"]) > 0 {
		f.addAll(w, req, redirectURL)
		return
	}

	if len(req.Form["remove-all-range"]) > 0 {
		redirectURL = fmt.Sprintf("/filters/%s/dimensions/%s/remove-all", filterID, name)
		http.Redirect(w, req, redirectURL, 302)
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
		dats = dates.Sort(dats)

		if r.Start == "select" {
			r.Start = dates.ConvertToMonthYear(dats[0])
		}
		if r.End == "select" {
			r.End = dates.ConvertToMonthYear(dats[len(dats)-1])
		}

		start, err := time.Parse("01 January 2006", fmt.Sprintf("01 %s", r.Start))
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
			http.Redirect(w, req, redirectURL, 302)
		}

		end, err := time.Parse("01 January 2006", fmt.Sprintf("01 %s", r.End))
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
			http.Redirect(w, req, redirectURL, 302)
		}

		if end.Before(start) {
			log.Info("end date before start date", log.Data{"start": start, "end": end})
			w.WriteHeader(http.StatusInternalServerError)
			redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
			http.Redirect(w, req, redirectURL, 302)
		}

		values = dates.ConvertToCoded(dats)
		var options []string
		for i, dat := range dats {
			if dat.Equal(start) || dat.After(start) && dat.Before(end) || dat.Equal(end) {
				options = append(options, labelIDMap[values[i]])
			}
		}

		if err := f.FilterClient.AddDimensionValues(filterID, name, options); err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, req, redirectURL, 302)
}

func (f *Filter) addAll(w http.ResponseWriter, req *http.Request, redirectURL string) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]

	codeID := "64d384f1-ea3b-445c-8fb8-aa453f96e58a"
	vals, err := f.CodeListClient.GetValues(codeID)
	if err != nil {
		log.ErrorR(req, err, nil)
		return
	}

	var options []string
	for _, val := range vals.Items {
		options = append(options, val.ID)
	}

	if err := f.FilterClient.AddDimensionValues(filterID, name, options); err != nil {
		log.ErrorR(req, err, nil)
		return
	}

	http.Redirect(w, req, redirectURL, 302)
}

// AddList adds a list of values
func (f *Filter) AddList(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]

	if err := req.ParseForm(); err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)

	if len(req.Form["add-all"]) > 0 {
		redirectURL = fmt.Sprintf("/filters/%s/dimensions/%s?selectorType=list", filterID, name)
		f.addAll(w, req, redirectURL)
		return
	}

	if len(req.Form["remove-all"]) > 0 {
		redirectURL = fmt.Sprintf("/filters/%s/dimensions/%s/remove-all?selectorType=list", filterID, name)
		http.Redirect(w, req, redirectURL, 302)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// concurrently remove any fields that have been deselected
	go func() {
		opts, err := f.FilterClient.GetDimensionOptions(filterID, name)
		if err != nil {
			log.ErrorR(req, err, nil)
		}

		for _, opt := range opts {
			if _, ok := req.Form[opt.Option]; !ok {
				if err := f.FilterClient.RemoveDimensionValue(filterID, name, opt.Option); err != nil {
					log.ErrorR(req, err, nil)
				}
			}
		}

		wg.Done()
	}()

	for k := range req.Form {
		if k == ":uri" || k == "save-and-return" {
			continue
		}

		if err := f.FilterClient.AddDimensionValue(filterID, name, k); err != nil {
			log.TraceR(req, err.Error(), nil)
			continue
		}
	}

	wg.Wait()

	http.Redirect(w, req, redirectURL, 302)
}

func (f *Filter) getDimensionValues(filterID, name string) (values []string, labelIDMap map[string]string, err error) {
	dim, err := f.FilterClient.GetDimension(filterID, name)
	if err != nil {
		return
	}

	log.Debug("dimension", log.Data{"dimension": dim})

	/*codeID := getCodeIDFromURI(dim.URI) // TODO: uncomment when real codeid becomes available
	if codeID == "" {
		err = errors.New("missing code id from uri")
		return
	}*/

	codeID := "64d384f1-ea3b-445c-8fb8-aa453f96e58a"
	allValues, err := f.CodeListClient.GetValues(codeID)
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
	log.Debug("attempting to remove all", nil)
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	selectorType := req.URL.Query().Get("selectorType")

	if err := f.FilterClient.RemoveDimension(filterID, name); err != nil {
		log.ErrorR(req, err, nil)
	}

	if err := f.FilterClient.AddDimension(filterID, name); err != nil {
		log.ErrorR(req, err, nil)
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
	if selectorType == "list" {
		redirectURL = redirectURL + "?selectorType=list"
	}
	http.Redirect(w, req, redirectURL, 302)
}

// DimensionRemoveOne ...
func (f *Filter) DimensionRemoveOne(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	option := vars["option"]
	selectorType := req.URL.Query().Get("selectorType")

	if err := f.FilterClient.RemoveDimensionValue(filterID, name, option); err != nil {
		log.ErrorR(req, err, nil)
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
	if selectorType == "list" {
		redirectURL = redirectURL + "?selectorType=list"
	}
	http.Redirect(w, req, redirectURL, 302)
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
