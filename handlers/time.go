package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

var acceptedReg = regexp.MustCompile(`^\w{3}-\d{2}$`)

// UpdateTime will update the time filter based on the radio selected filters by the user
func (f *Filter) UpdateTime(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	_, filterCfg := setAuthTokenIfRequired(req)

	if err := f.FilterClient.RemoveDimension(filterID, "time", filterCfg...); err != nil {
		setStatusCode(req, w, err)
		return
	}

	if err := f.FilterClient.AddDimension(filterID, "time", filterCfg...); err != nil {
		setStatusCode(req, w, err)
		return
	}

	if err := req.ParseForm(); err != nil {
		setStatusCode(req, w, err)
		return
	}

	if len(req.Form.Get("add-all")) > 0 {
		http.Redirect(w, req, fmt.Sprintf("/filters/%s/dimensions/time/add-all", filterID), 302)
		return
	}

	if len(req.Form.Get("remove-all")) > 0 {
		http.Redirect(w, req, fmt.Sprintf("/filters/%s/dimensions/time/remove-all", filterID), 302)
		return
	}

	switch req.Form.Get("time-selection") {
	case "latest":
		if err := f.FilterClient.AddDimensionValue(filterID, "time", req.Form.Get("latest-option"), filterCfg...); err != nil {
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

	_, filterCfg := setAuthTokenIfRequired(req)

	date, err := time.Parse("January 2006", fmt.Sprintf("%s %s", month, year))
	if err != nil {
		return err
	}

	return f.FilterClient.AddDimensionValue(filterID, "time", date.Format("Jan-06"), filterCfg...)
}

func (f *Filter) addTimeList(filterID string, req *http.Request) error {
	_, filterCfg := setAuthTokenIfRequired(req)

	opts, err := f.FilterClient.GetDimensionOptions(filterID, "time", filterCfg...)
	if err != nil {
		return err
	}

	// Remove any unselected times
	for _, opt := range opts {
		if _, ok := req.Form[opt.Option]; !ok {
			if err := f.FilterClient.RemoveDimensionValue(filterID, "time", opt.Option, filterCfg...); err != nil {
				log.ErrorR(req, err, nil)
			}
		}
	}

	var options []string
	for k := range req.Form {
		if _, err := time.Parse("Jan-06", k); err != nil {
			continue
		}

		options = append(options, k)
	}

	if err := f.FilterClient.AddDimensionValues(filterID, "time", options, filterCfg...); err != nil {
		log.TraceR(req, err.Error(), nil)
	}

	return nil
}

func (f *Filter) addTimeRange(filterID string, req *http.Request) error {
	startMonth := req.Form.Get("start-month")
	startYear := req.Form.Get("start-year")
	endMonth := req.Form.Get("end-month")
	endYear := req.Form.Get("end-year")

	datasetCfg, filterCfg := setAuthTokenIfRequired(req)

	values, labelIDMap, err := f.getDimensionValues(filterID, "time", datasetCfg, filterCfg)
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

	return f.FilterClient.AddDimensionValues(filterID, "time", options, filterCfg...)
}

// Time specifically handles the data for the time dimension page
func (f *Filter) Time(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	datasetCfg, filterCfg := setAuthTokenIfRequired(req)

	fj, err := f.FilterClient.GetJobState(filterID, filterCfg...)
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

	dataset, err := f.DatasetClient.Get(datasetID, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}
	ver, err := f.DatasetClient.GetVersion(datasetID, edition, version, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	allValues, err := f.DatasetClient.GetOptions(datasetID, edition, version, "time", datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	if len(allValues.Items) <= 20 || !acceptedReg.MatchString(allValues.Items[0].Option) {
		mux.Vars(req)["name"] = "time"
		f.DimensionSelector(w, req)
		return
	}

	selValues, err := f.FilterClient.GetDimensionOptions(filterID, "time", filterCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dims, err := f.DatasetClient.GetDimensions(datasetID, edition, version, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	p, err := mapper.CreateTimePage(fj, dataset, ver, allValues, selValues, dims, datasetID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	b, err := json.Marshal(p)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/time", b)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	if _, err := w.Write(templateBytes); err != nil {
		setStatusCode(req, w, err)
		return
	}

}
