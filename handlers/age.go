package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

	log.Debug("updating age", nil)

	if err := f.FilterClient.RemoveDimension(filterID, "age"); err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := f.FilterClient.AddDimension(filterID, "age"); err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := req.ParseForm(); err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
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

	for k := range req.Form {
		if _, err := strconv.Atoi(k); err != nil {
			if !strings.Contains(k, "+") {
				continue
			}
		}

		if err := f.FilterClient.AddDimensionValue(filterID, "age", k); err != nil {
			log.TraceR(req, err.Error(), nil)
			continue
		}
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

	allValues, err := f.DatasetClient.GetOptions(datasetID, edition, version, "age")
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(allValues.Items) <= 20 {
		mux.Vars(req)["name"] = "age"
		f.DimensionSelector(w, req)
		return
	}

	selValues, err := f.FilterClient.GetDimensionOptions(filterID, "age")
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	p, err := mapper.CreateAgePage(fj, dataset, ver, allValues, selValues, datasetID)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(p)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/age", b)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(templateBytes); err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
