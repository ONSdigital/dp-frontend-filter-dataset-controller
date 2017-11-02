package handlers

import (
	"fmt"
	"net/http"

	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// UseLatest will create a new filter job for the same dataset with the
// latest version
func (f *Filter) UseLatest(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	oldJob, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dims, err := f.FilterClient.GetDimensions(filterID)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newFilterID, err := f.FilterClient.CreateJob(oldJob.InstanceID, []string{})
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, dim := range dims {
		if err := f.FilterClient.AddDimension(newFilterID, dim.Name); err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		dimValues, err := f.FilterClient.GetDimensionOptions(filterID, dim.Name)
		if err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var vals []string
		for _, val := range dimValues {
			vals = append(vals, val.Option)
		}

		if err := f.FilterClient.AddDimensionValues(newFilterID, dim.Name, vals); err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions", newFilterID)
	http.Redirect(w, req, redirectURL, 302)

}
