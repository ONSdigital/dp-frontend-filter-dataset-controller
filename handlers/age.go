package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

func (f *Filter) UpdateAge(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	if err := f.FilterClient.RemoveDimension(filterID, "age"); err != nil {
		setStatusCode(req, w, err)
		return
	}

	if err := f.FilterClient.AddDimension(filterID, "age"); err != nil {
		setStatusCode(req, w, err)
		return
	}

	if err := req.ParseForm(); err != nil {
		setStatusCode(req, w, err)
		return
	}

	if len(req.Form.Get("add-all")) > 0 {
		http.Redirect(w, req, fmt.Sprintf("/filters/%s/dimensions/age/add-all", filterID), 302)
		return
	}

	if len(req.Form.Get("remove-all")) > 0 {
		http.Redirect(w, req, fmt.Sprintf("/filters/%s/dimensions/age/remove-all", filterID), 302)
		return
	}

	log.Debug("age-selection", log.Data{"age": req.Form.Get("age-selection")})

	switch req.Form.Get("age-selection") {
	case "all":
		if err := f.FilterClient.AddDimensionValue(filterID, "age", req.Form.Get("all-ages-option")); err != nil {
			log.ErrorR(req, err, nil)
		}
	case "range":
		if err := f.addAgeRange(filterID, req); err != nil {
			log.ErrorR(req, err, nil)
		}
	case "list":
		if err := f.addAgeList(filterID, req); err != nil {
			log.ErrorR(req, err, nil)
		}
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)
	http.Redirect(w, req, redirectURL, 302)
}

func (f *Filter) addAgeList(filterID string, req *http.Request) error {
	opts, err := f.FilterClient.GetDimensionOptions(filterID, "age")
	if err != nil {
		return err
	}

	// Remove any unselected ages
	for _, opt := range opts {
		if _, ok := req.Form[opt.Option]; !ok {
			if err := f.FilterClient.RemoveDimensionValue(filterID, "age", opt.Option); err != nil {
				log.ErrorR(req, err, nil)
			}
		}
	}

	var options []string
	for k := range req.Form {
		if _, err := strconv.Atoi(k); err != nil {
			if !strings.Contains(k, "+") {
				continue
			}
		}

		options = append(options, k)

	}

	if err := f.FilterClient.AddDimensionValues(filterID, "age", options); err != nil {
		log.TraceR(req, err.Error(), nil)
	}

	return nil
}

func (f *Filter) addAgeRange(filterID string, req *http.Request) error {
	youngest := req.Form.Get("youngest")
	oldest := req.Form.Get("oldest")

	values, labelIDMap, err := f.getDimensionValues(filterID, "age")
	if err != nil {
		return err
	}

	var intValues []int
	for _, val := range values {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			break
		}

		intValues = append(intValues, intVal)
	}

	sort.Ints(intValues)

	for i, val := range intValues {
		values[i] = strconv.Itoa(val)
	}

	var isInRange bool
	var options []string
	for _, age := range values {
		if youngest == age {
			isInRange = true
		}

		if isInRange {
			options = append(options, labelIDMap[age])
		}

		if oldest == age {
			isInRange = false
		}
	}

	return f.FilterClient.AddDimensionValues(filterID, "age", options)
}

func (f *Filter) Age(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	fj, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dataset, err := f.DatasetClient.Get(datasetID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}
	ver, err := f.DatasetClient.GetVersion(datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	allValues, err := f.DatasetClient.GetOptions(datasetID, edition, version, "age")
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	if len(allValues.Items) <= 20 {
		mux.Vars(req)["name"] = "age"
		f.DimensionSelector(w, req)
		return
	}

	selValues, err := f.FilterClient.GetDimensionOptions(filterID, "age")
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dims, err := f.DatasetClient.GetDimensions(datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	p, err := mapper.CreateAgePage(fj, dataset, ver, allValues, selValues, dims, datasetID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	b, err := json.Marshal(p)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/age", b)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	if _, err := w.Write(templateBytes); err != nil {
		setStatusCode(req, w, err)
		return
	}

}
