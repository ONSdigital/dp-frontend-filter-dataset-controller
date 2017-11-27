package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// UpdateTime will update the time filter based on the radio selected filters by the user
func (f *Filter) UpdateTime(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	if err := f.FilterClient.RemoveDimension(filterID, "time"); err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := f.FilterClient.AddDimension(filterID, "time"); err != nil {
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
		http.Redirect(w, req, fmt.Sprintf("/filters/%s/dimensions/time/add-all", filterID), 302)
		return
	}

	switch req.Form.Get("time-selection") {
	case "latest":
		if err := f.FilterClient.AddDimensionValue(filterID, "time", req.Form.Get("latest-option")); err != nil {
			log.ErrorR(req, err, nil)
		}
	case "single":
		if err := f.addSingleTime(filterID, req); err != nil {
			log.ErrorR(req, err, nil)
		}
	case "range":
		if err := f.addTimeRange(filterID, req); err != nil {
			log.ErrorR(req, err, nil)
		}
	case "list":
		if err := f.addTimeList(filterID, req); err != nil {
			log.ErrorR(req, err, nil)
		}
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)
	http.Redirect(w, req, redirectURL, 302)
}

func (f *Filter) addSingleTime(filterID string, req *http.Request) error {
	month := req.Form.Get("month-single")
	year := req.Form.Get("year-single")

	date, err := time.Parse("January 2006", fmt.Sprintf("%s %s", month, year))
	if err != nil {
		return err
	}

	return f.FilterClient.AddDimensionValue(filterID, "time", date.Format("Jan-06"))
}

func (f *Filter) addTimeList(filterID string, req *http.Request) error {
	opts, err := f.FilterClient.GetDimensionOptions(filterID, "time")
	if err != nil {
		return err
	}

	// Remove any unselected times
	for _, opt := range opts {
		if _, ok := req.Form[opt.Option]; !ok {
			if err := f.FilterClient.RemoveDimensionValue(filterID, "time", opt.Option); err != nil {
				log.ErrorR(req, err, nil)
			}
		}
	}

	for k := range req.Form {
		if _, err := time.Parse("Jan-06", k); err != nil {
			continue
		}

		if err := f.FilterClient.AddDimensionValue(filterID, "time", k); err != nil {
			log.TraceR(req, err.Error(), nil)
			continue
		}
	}

	return nil
}

func (f *Filter) addTimeRange(filterID string, req *http.Request) error {
	startMonth := req.Form.Get("start-month")
	startYear := req.Form.Get("start-year")
	endMonth := req.Form.Get("end-month")
	endYear := req.Form.Get("end-year")

	values, labelIDMap, err := f.getDimensionValues(filterID, "time")
	if err != nil {
		return err
	}

	dats, err := dates.ConvertToReadable(values)
	if err != nil {
		return err
	}
	dats = dates.Sort(dats)

	start, err := time.Parse("01 January 2006", fmt.Sprintf("01 %s %s", startMonth, startYear))
	if err != nil {
		return err
	}

	end, err := time.Parse("01 January 2006", fmt.Sprintf("01 %s %s", endMonth, endYear))
	if err != nil {
		return err
	}

	if end.Before(start) {
		return fmt.Errorf("start date: %s before end date: %s", start.String(), end.String())
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
		if err := f.FilterClient.AddDimensionValue(filterID, "time", opt); err != nil {
			return err
		}
	}

	return nil
}

// Time specifically handles the data for the time dimension page
func (f *Filter) Time(w http.ResponseWriter, req *http.Request) {
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

	allValues, err := f.DatasetClient.GetOptions(datasetID, edition, version, "time")
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(allValues.Items) <= 20 {
		mux.Vars(req)["name"] = "time"
		f.DimensionSelector(w, req)
		return
	}

	selValues, err := f.FilterClient.GetDimensionOptions(filterID, "time")
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	p, err := mapper.CreateTimePage(fj, dataset, ver, allValues, selValues, datasetID)
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

	templateBytes, err := f.Renderer.Do("dataset-filter/time", b)
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
