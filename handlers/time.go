package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/headers"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

var acceptedReg = regexp.MustCompile(`^\w{3}-\d{2}$`)

// UpdateTime will update the time filter based on the radio selected filters by the user
func (f *Filter) UpdateTime(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]
	ctx := req.Context()

	dimensionName := "time"

	collectionID := getCollectionIDFromContext(ctx)
	userAccessToken, err := headers.GetUserAuthToken(req)
	if err != nil {
		if headers.IsNotErrNotFound(err) {
			log.Event(ctx, "error getting access token header", log.WARN, log.Error(err))
		}
	}

	if err := f.FilterClient.RemoveDimension(ctx, userAccessToken, "", collectionID, filterID, dimensionName); err != nil {
		log.Event(ctx, "failed to remove dimension", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dimensionName})
		setStatusCode(req, w, err)
		return
	}

	if err := f.FilterClient.AddDimension(ctx, userAccessToken, "", collectionID, filterID, dimensionName); err != nil {
		log.Event(ctx, "failed to add dimension", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dimensionName})
		setStatusCode(req, w, err)
		return
	}

	if err := req.ParseForm(); err != nil {
		log.Event(ctx, "failed to parse form", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
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
		if err := f.FilterClient.AddDimensionValue(ctx, userAccessToken, "", collectionID, filterID, dimensionName, req.Form.Get("latest-option")); err != nil {
			log.Event(ctx, "failed to add dimension value", log.ERROR, log.Error(err))
		}
	case "single":
		if err := f.addSingleTime(filterID, req); err != nil {
			log.Event(ctx, "failed to add single time", log.ERROR, log.Error(err))
		}
	case "range":
		if err := f.addTimeRange(filterID, req); err != nil {
			log.Event(ctx, "failed to add range of times", log.ERROR, log.Error(err))
		}
	case "list":
		if err := f.addTimeList(filterID, req); err != nil {
			log.Event(ctx, "failed to add list of times", log.ERROR, log.Error(err))
		}
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)
	http.Redirect(w, req, redirectURL, 302)
}

func (f *Filter) addSingleTime(filterID string, req *http.Request) error {
	ctx := req.Context()

	month := req.Form.Get("month-single")
	year := req.Form.Get("year-single")
	dimensionName := "time"

	collectionID := getCollectionIDFromContext(ctx)
	userAccessToken, err := headers.GetUserAuthToken(req)
	if err != nil {
		if headers.IsNotErrNotFound(err) {
			log.Event(ctx, "error getting access token header", log.WARN, log.Error(err))
		}
	}

	date, err := time.Parse("January 2006", fmt.Sprintf("%s %s", month, year))
	if err != nil {
		return err
	}

	return f.FilterClient.AddDimensionValue(ctx, userAccessToken, "", collectionID, filterID, dimensionName, date.Format("Jan-06"))
}

func (f *Filter) addTimeList(filterID string, req *http.Request) error {
	ctx := req.Context()
	collectionID := getCollectionIDFromContext(ctx)
	dimensionName := "time"

	userAccessToken, err := headers.GetUserAuthToken(req)
	if err != nil {
		if headers.IsNotErrNotFound(err) {
			log.Event(ctx, "error getting access token header", log.WARN, log.Error(err))
		}
	}

	opts, err := f.FilterClient.GetDimensionOptions(ctx, userAccessToken, "", collectionID, filterID, dimensionName)
	if err != nil {
		return err
	}

	// Remove any unselected times
	for _, opt := range opts {
		if _, ok := req.Form[opt.Option]; !ok {
			if err := f.FilterClient.RemoveDimensionValue(ctx, userAccessToken, "", collectionID, filterID, dimensionName, opt.Option); err != nil {
				log.Event(ctx, "failed to remove dimension value", log.WARN, log.Error(err))
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

	if err := f.FilterClient.AddDimensionValues(ctx, userAccessToken, "", collectionID, filterID, dimensionName, options); err != nil {
		log.Event(ctx, "failed to add dimension values", log.ERROR, log.Error(err))
	}

	// Should we not be returning error?
	return nil
}

func (f *Filter) addTimeRange(filterID string, req *http.Request) error {
	startMonth := req.Form.Get("start-month")
	startYear := req.Form.Get("start-year")
	endMonth := req.Form.Get("end-month")
	endYear := req.Form.Get("end-year")
	ctx := req.Context()
	dimensionName := "time"

	collectionID := getCollectionIDFromContext(ctx)
	userAccessToken, err := headers.GetUserAuthToken(req)
	if err != nil {
		if headers.IsNotErrNotFound(err) {
			log.Event(ctx, "error getting access token header", log.WARN, log.Error(err))
		}
	}

	values, labelIDMap, err := f.getDimensionValues(ctx, userAccessToken, filterID, dimensionName)
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

	return f.FilterClient.AddDimensionValues(ctx, userAccessToken, "", collectionID, filterID, dimensionName, options)
}

// Time specifically handles the data for the time dimension page
func (f *Filter) Time(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]
	ctx := req.Context()
	dimensionName := "time"

	collectionID := getCollectionIDFromContext(ctx)
	userAccessToken, err := headers.GetUserAuthToken(req)
	if err != nil {
		if headers.IsNotErrNotFound(err) {
			log.Event(ctx, "error getting access token header", log.WARN, log.Error(err))
		}
	}

	fj, err := f.FilterClient.GetJobState(ctx, userAccessToken, "", "", collectionID, filterID)
	if err != nil {
		log.Event(ctx, "failed to get job state", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		log.Event(ctx, "failed to parse version href", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}
	versionPath := strings.TrimPrefix(versionURL.Path, f.APIRouterVersion)
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
	if err != nil {
		log.Event(ctx, "failed to extract dataset info from path", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "path": versionPath})
		setStatusCode(req, w, err)
		return
	}

	dataset, err := f.DatasetClient.Get(ctx, userAccessToken, "", collectionID, datasetID)
	if err != nil {
		log.Event(ctx, "failed to get dataset", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID})
		setStatusCode(req, w, err)
		return
	}
	ver, err := f.DatasetClient.GetVersion(ctx, userAccessToken, "", "", collectionID, datasetID, edition, version)
	if err != nil {
		log.Event(ctx, "failed to get version", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	allValues, err := f.DatasetClient.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dimensionName)
	if err != nil {
		log.Event(ctx, "failed to get options from dataset client", log.ERROR, log.Error(err),
			log.Data{"dimension": dimensionName, "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	//use normal list format unless a specially recognized time format
	if len(allValues.Items) <= 20 || !acceptedReg.MatchString(allValues.Items[0].Option) {
		mux.Vars(req)["name"] = dimensionName
		f.DimensionSelector(w, req)
		return
	}

	selValues, err := f.FilterClient.GetDimensionOptions(ctx, userAccessToken, "", collectionID, filterID, dimensionName)
	if err != nil {
		log.Event(ctx, "failed to get options from filter client", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dimensionName})
		setStatusCode(req, w, err)
		return
	}

	dims, err := f.DatasetClient.GetVersionDimensions(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
	if err != nil {
		log.Event(ctx, "failed to get dimensions", log.ERROR, log.Error(err),
			log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	p, err := mapper.CreateTimePage(req, fj, dataset, ver, allValues, selValues, dims, datasetID, f.APIRouterVersion)
	if err != nil {
		log.Event(ctx, "failed to map data to page", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dataset_id": datasetID, "dimension": dimensionName})
		setStatusCode(req, w, err)
		return
	}

	b, err := json.Marshal(p)
	if err != nil {
		log.Event(ctx, "failed to marshal json", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/time", b)
	if err != nil {
		log.Event(ctx, "failed to render", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	if _, err := w.Write(templateBytes); err != nil {
		log.Event(ctx, "failed to write response", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

}
