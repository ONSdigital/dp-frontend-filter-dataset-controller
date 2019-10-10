package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-api-clients-go/headers"
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
	ctx := req.Context()

	dimensionName := "time"

	collectionID := getCollectionIDFromContext(ctx)
	userAccessToken, err := headers.GetUserAuthToken(req)
	if !headers.IsNotFound(err) {
		log.Error(err, nil)
	}


	if err := f.FilterClient.RemoveDimension(req.Context(), userAccessToken, "", collectionID, filterID, dimensionName); err != nil {
		log.InfoCtx(ctx, "failed to remove dimension", log.Data{"error": err, "filter_id": filterID, "dimension": dimensionName})
		setStatusCode(req, w, err)
		return
	}

	if err := f.FilterClient.AddDimension(req.Context(), userAccessToken, "", collectionID, filterID, dimensionName); err != nil {
		log.InfoCtx(ctx, "failed to add dimension", log.Data{"error": err, "filter_id": filterID, "dimension": dimensionName})
		setStatusCode(req, w, err)
		return
	}

	if err := req.ParseForm(); err != nil {
		log.InfoCtx(ctx, "failed to parse form", log.Data{"error": err, "filter_id": filterID})
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
		if err := f.FilterClient.AddDimensionValue(req.Context(), userAccessToken, "", collectionID, filterID, dimensionName, req.Form.Get("latest-option")); err != nil {
			log.ErrorCtx(ctx, err, nil)
		}
	case "single":
		if err := f.addSingleTime(filterID, req); err != nil {
			log.ErrorCtx(ctx, err, nil)
		}
	case "range":
		if err := f.addTimeRange(filterID, req); err != nil {
			log.ErrorCtx(ctx, err, nil)
		}
	case "list":
		if err := f.addTimeList(filterID, req); err != nil {
			log.ErrorCtx(ctx, err, nil)
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
	if !headers.IsNotFound(err) {
		log.Error(err, nil)
	}

	date, err := time.Parse("January 2006", fmt.Sprintf("%s %s", month, year))
	if err != nil {
		return err
	}


	return f.FilterClient.AddDimensionValue(req.Context(), userAccessToken, "", collectionID, filterID, dimensionName, date.Format("Jan-06"))
}

func (f *Filter) addTimeList(filterID string, req *http.Request) error {
	ctx := req.Context()
	collectionID := getCollectionIDFromContext(ctx)
	dimensionName := "time"

	userAccessToken, err := headers.GetUserAuthToken(req)
	if !headers.IsNotFound(err) {
		log.Error(err, nil)
	}


	opts, err := f.FilterClient.GetDimensionOptions(req.Context(), userAccessToken, "", collectionID, filterID, dimensionName)
	if err != nil {
		return err
	}

	// Remove any unselected times
	for _, opt := range opts {
		if _, ok := req.Form[opt.Option]; !ok {
			if err := f.FilterClient.RemoveDimensionValue(req.Context(), userAccessToken, "", collectionID, filterID, dimensionName, opt.Option); err != nil {
				log.ErrorCtx(ctx, err, nil)
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

	if err := f.FilterClient.AddDimensionValues(req.Context(), userAccessToken, "", collectionID, filterID, dimensionName, options); err != nil {
		log.TraceCtx(ctx, err.Error(), nil)
	}

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
	if !headers.IsNotFound(err) {
		log.Error(err, nil)
	}

	values, labelIDMap, err := f.getDimensionValues(req.Context(), userAccessToken, filterID, dimensionName)
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


	return f.FilterClient.AddDimensionValues(req.Context(), userAccessToken, "", collectionID, filterID, dimensionName, options)
}

// Time specifically handles the data for the time dimension page
func (f *Filter) Time(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]
	ctx := req.Context()
	dimensionName := "time"

	collectionID := getCollectionIDFromContext(ctx)
	userAccessToken, err := headers.GetUserAuthToken(req)
	if !headers.IsNotFound(err) {
		log.Error(err, nil)
	}

	fj, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
	if err != nil {
		log.InfoCtx(ctx, "failed to get job state", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		log.InfoCtx(ctx, "failed to parse version href", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		log.InfoCtx(ctx, "failed to extract dataset info from path", log.Data{"error": err, "filter_id": filterID, "path": versionURL})
		setStatusCode(req, w, err)
		return
	}

	dataset, err := f.DatasetClient.Get(req.Context(), userAccessToken, "", collectionID, datasetID)
	if err != nil {
		log.InfoCtx(ctx, "failed to get dataset", log.Data{"error": err, "dataset_id": datasetID})
		setStatusCode(req, w, err)
		return
	}
	ver, err := f.DatasetClient.GetVersion(req.Context(), userAccessToken, "", "", collectionID, datasetID, edition, version)
	if err != nil {
		log.InfoCtx(ctx, "failed to get version", log.Data{"error": err, "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	allValues, err := f.DatasetClient.GetOptions(req.Context(),  userAccessToken, "", collectionID, datasetID, edition, version, dimensionName)
	if err != nil {
		log.InfoCtx(ctx, "failed to get options from dataset client",
			log.Data{"error": err, "dimension": dimensionName, "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	//use normal list format unless a specially recognized time format
	if len(allValues.Items) <= 20 || !acceptedReg.MatchString(allValues.Items[0].Option) {
		mux.Vars(req)["name"] = dimensionName
		f.DimensionSelector(w, req)
		return
	}

	selValues, err := f.FilterClient.GetDimensionOptions(req.Context(), userAccessToken, "", collectionID, filterID, dimensionName)
	if err != nil {
		log.InfoCtx(ctx, "failed to get options from filter client", log.Data{"error": err, "filter_id": filterID, "dimension": dimensionName})
		setStatusCode(req, w, err)
		return
	}

	dims, err := f.DatasetClient.GetDimensions(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version)
	if err != nil {
		log.InfoCtx(ctx, "failed to get dimensions",
			log.Data{"error": err, "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	p, err := mapper.CreateTimePage(req.Context(), fj, dataset, ver, allValues, selValues, dims, datasetID)
	if err != nil {
		log.InfoCtx(ctx, "failed to map data to page", log.Data{"error": err, "filter_id": filterID, "dataset_id": datasetID, "dimension": dimensionName})
		setStatusCode(req, w, err)
		return
	}

	b, err := json.Marshal(p)
	if err != nil {
		log.InfoCtx(ctx, "failed to marshal json", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/time", b)
	if err != nil {
		log.InfoCtx(ctx, "failed to render", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	if _, err := w.Write(templateBytes); err != nil {
		log.InfoCtx(ctx, "failed to write response", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

}
