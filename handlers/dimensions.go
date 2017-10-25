package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/validator"
	"github.com/gorilla/mux"
)

type labelID struct {
	Label string `json:"label"`
	ID    string `json:"id"`
}

// GetAllDimensionOptionsJSON will return a list of all options from the dataset api
func (f *Filter) GetAllDimensionOptionsJSON(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]

	fj, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	idNameMap, err := f.getIDNameMap(versionURL.Path, name)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var lids []labelID

	if name == "time" {

		var codedDates []string
		labelIDMap := make(map[string]string)
		for k, v := range idNameMap {
			codedDates = append(codedDates, v)
			labelIDMap[v] = k
		}

		readbleDates, err := dates.ConvertToReadable(codedDates)
		if err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		readbleDates = dates.Sort(readbleDates)

		for _, date := range readbleDates {
			lid := labelID{
				Label: fmt.Sprintf("%s %d", date.Month(), date.Year()),
				ID:    labelIDMap[date.Format("Jan-06")],
			}

			lids = append(lids, lid)
		}
	}

	b, err := json.Marshal(lids)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func (f *Filter) getIDNameMap(versionURL, dimension string) (map[string]string, error) {
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL)
	if err != nil {
		return nil, err
	}

	idNameMap := make(map[string]string)
	opts, err := f.DatasetClient.GetOptions(datasetID, edition, version, dimension)
	if err != nil {
		return nil, err
	}

	log.Debug("creating an option-label map from dataset dimensions", log.Data{"version-url": versionURL, "options": opts})

	for _, opt := range opts.Items {
		idNameMap[opt.Option] = opt.Label
	}

	return idNameMap, nil
}

// GetSelectedDimensionOptionsJSON will return a list of selected options from the filter api with corresponding label
func (f *Filter) GetSelectedDimensionOptionsJSON(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]

	opts, err := f.FilterClient.GetDimensionOptions(filterID, name)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fj, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	idNameMap, err := f.getIDNameMap(versionURL.Path, name)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var lids []labelID

	if name == "time" {

		var codedDates []string
		labelIDMap := make(map[string]string)
		for _, opt := range opts {
			codedDates = append(codedDates, idNameMap[opt.Option])
			labelIDMap[idNameMap[opt.Option]] = opt.Option
		}

		readbleDates, err := dates.ConvertToReadable(codedDates)
		if err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		readbleDates = dates.Sort(readbleDates)

		for _, date := range readbleDates {
			lid := labelID{
				Label: fmt.Sprintf("%s %d", date.Month(), date.Year()),
				ID:    labelIDMap[date.Format("Jan-06")],
			}

			lids = append(lids, lid)
		}
	}

	b, err := json.Marshal(lids)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

// DimensionSelector controls the render of the range selector template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) DimensionSelector(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]

	/*	if name == "aggregate" || name == "CPI" {
		//url := fmt.Sprintf("/filters/%s/hierarchies/%s", filterID, name)
		url := fmt.Sprintf("/filters/%s/dimensions/%s?selectorType=list", filterID, name)
		http.Redirect(w, req, url, 302)
	} */ // TODO: uncomment when hierarchy api is completed

	filter, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	selectedValues, err := f.FilterClient.GetDimensionOptions(filterID, name)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fj, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dataset, err := f.DatasetClient.Get(datasetID)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ver, err := f.DatasetClient.GetVersion(datasetID, edition, version)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	allValues, err := f.DatasetClient.GetOptions(datasetID, edition, version, name)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	selectorType := req.URL.Query().Get("selectorType")
	if selectorType == "list" {
		f.listSelector(w, req, name, selectedValues, allValues, filter, dataset, datasetID, ver.ReleaseDate)
	} else {
		f.rangeSelector(w, req, name, selectedValues, allValues, filter, dataset, datasetID, ver.ReleaseDate)
	}
}

func (f *Filter) rangeSelector(w http.ResponseWriter, req *http.Request, name string, selectedValues []filter.DimensionOption, allValues dataset.Options, filter filter.Model, dataset dataset.Model, datasetID, releaseDate string) {

	p := mapper.CreateRangeSelectorPage(name, selectedValues, allValues, filter, dataset, datasetID, releaseDate)

	b, err := json.Marshal(p)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/range-selector", b)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}

// ListSelector controls the render of the age selector list template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) listSelector(w http.ResponseWriter, req *http.Request, name string, selectedValues []filter.DimensionOption, allValues dataset.Options, filter filter.Model, dataset dataset.Model, datasetID, releaseDate string) {
	p := mapper.CreateListSelectorPage(name, selectedValues, allValues, filter, dataset, datasetID, releaseDate)

	b, err := json.Marshal(p)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/list-selector", b)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}

// Range represents range labels in the range selector page
type Range struct {
	AddAll         string `schema:"add-all"`
	AddAllRange    string `schema:"add-all-range"`
	End            string `schema:"end"`
	EndMonth       string `schema:"end-month"`
	EndYear        string `schema:"end-year"`
	RemoveAllRange string `schema:"remove-all-range"`
	Start          string `schema:"start"`
	StartMonth     string `schema:"start-month"`
	StartYear      string `schema:"start-year"`
	SaveAndReturn  string `schema:"save-and-return"`
}

// AddRange will add a range of values to a filter job
func (f *Filter) AddRange(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]

	if err := req.ParseForm(); err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var r Range

	if err := f.val.Validate(req, &r); err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		if _, ok := err.(validator.ErrFormValidationFailed); ok {
			errs := err.(validator.ErrFormValidationFailed).GetFieldErrors()
			log.Debug("field errors", log.Data{"errs": errs})
		}
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
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if name == "time" {
		dats, err := dates.ConvertToReadable(values)
		if err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
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

		start, err := time.Parse("01 January 2006", fmt.Sprintf("01 %s %s", r.StartMonth, r.StartYear))
		if err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
			http.Redirect(w, req, redirectURL, 302)
		}

		end, err := time.Parse("01 January 2006", fmt.Sprintf("01 %s %s", r.EndMonth, r.EndYear))
		if err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
			http.Redirect(w, req, redirectURL, 302)
		}

		if end.Before(start) {
			log.Info("end date before start date", log.Data{"start": start, "end": end, "setting-response-status": http.StatusInternalServerError})
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

		log.InfoR(req, "options", log.Data{"options": options, "values": values})
		for _, opt := range options {
			if err := f.FilterClient.AddDimensionValue(filterID, name, opt); err != nil {
				log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

	}

	http.Redirect(w, req, redirectURL, 302)
}

// DimensionAddAll will add all dimension values to a basket
func (f *Filter) DimensionAddAll(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	f.addAll(w, req, fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name))
}

func (f *Filter) addAll(w http.ResponseWriter, req *http.Request, redirectURL string) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]

	fj, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vals, err := f.DatasetClient.GetOptions(datasetID, edition, version, name)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		return
	}

	var options []string
	for _, val := range vals.Items {
		options = append(options, val.Option)
	}

	if err := f.FilterClient.AddDimensionValues(filterID, name, options); err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
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
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
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
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
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

	wg.Wait()

	for k := range req.Form {
		if k == ":uri" || k == "save-and-return" {
			continue
		}

		if err := f.FilterClient.AddDimensionValue(filterID, name, k); err != nil {
			log.TraceR(req, err.Error(), nil)
			continue
		}
	}

	http.Redirect(w, req, redirectURL, 302)
}

func (f *Filter) getDimensionValues(filterID, name string) (values []string, labelIDMap map[string]string, err error) {
	fj, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		return
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		return
	}

	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		return
	}

	vals, err := f.DatasetClient.GetOptions(datasetID, edition, version, name)
	if err != nil {
		return
	}

	labelIDMap = make(map[string]string)
	for _, val := range vals.Items {
		values = append(values, val.Label)
		labelIDMap[val.Label] = val.Option
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
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := f.FilterClient.AddDimension(filterID, name); err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
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
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
	if selectorType == "list" {
		redirectURL = redirectURL + "?selectorType=list"
	}
	http.Redirect(w, req, redirectURL, 302)
}
