package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	dphandlers "github.com/ONSdigital/dp-net/v2/handlers"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

var acceptedReg = regexp.MustCompile(`^\w{3}-\d{2}$`)

// UpdateTime will update the time filter based on the radio selected filters by the user
func (f *Filter) UpdateTime() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		ctx := req.Context()
		dimensionName := "time"

		eTag, err := f.FilterClient.RemoveDimension(ctx, userAccessToken, "", collectionID, filterID, dimensionName, headers.IfMatchAnyETag)
		if err != nil {
			log.Error(ctx, "failed to remove dimension", err, log.Data{"filter_id": filterID, "dimension": dimensionName})
			setStatusCode(req, w, err)
			return
		}

		eTag, err = f.FilterClient.AddDimension(ctx, userAccessToken, "", collectionID, filterID, dimensionName, eTag)
		if err != nil {
			log.Error(ctx, "failed to add dimension", err, log.Data{"filter_id": filterID, "dimension": dimensionName})
			setStatusCode(req, w, err)
			return
		}

		if err := req.ParseForm(); err != nil {
			log.Error(ctx, "failed to parse form", err, log.Data{"filter_id": filterID})
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
			eTag, err = f.FilterClient.AddDimensionValue(ctx, userAccessToken, "", collectionID, filterID, dimensionName, req.Form.Get("latest-option"), eTag)
			if err != nil {
				log.Error(ctx, "failed to add dimension value", err)
			}
		case "single":
			eTag, err = f.addSingleTime(filterID, userAccessToken, collectionID, req, eTag)
			if err != nil {
				log.Error(ctx, "failed to add single time", err)
			}
		case "range":
			eTag, err = f.addTimeRange(filterID, userAccessToken, collectionID, req, eTag)
			if err != nil {
				log.Error(ctx, "failed to add range of times", err)
			}
		case "list":
			eTag, err = f.addTimeList(filterID, userAccessToken, collectionID, req, eTag)
			if err != nil {
				log.Error(ctx, "failed to add list of times", err)
			}
		}

		redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)
		http.Redirect(w, req, redirectURL, 302)
	})

}

func (f *Filter) addSingleTime(filterID, userAccessToken, collectionID string, req *http.Request, eTag string) (newETag string, err error) {
	ctx := req.Context()

	month := req.Form.Get("month-single")
	year := req.Form.Get("year-single")
	dimensionName := "time"

	date, err := time.Parse("January 2006", fmt.Sprintf("%s %s", month, year))
	if err != nil {
		return eTag, err
	}

	return f.FilterClient.AddDimensionValue(ctx, userAccessToken, "", collectionID, filterID, dimensionName, date.Format("Jan-06"), eTag)
}

// addTimeList will save form time grouped list form data to the filter
func (f *Filter) addTimeList(filterID, userAccessToken, collectionID string, req *http.Request, eTag string) (newETag string, err error) {
	ctx := req.Context()
	dimensionName := "time"

	var options []string
	startYearStr := req.Form.Get("start-year-grouped")
	endYearStr := req.Form.Get("end-year-grouped")
	startYearInt, err := strconv.Atoi(startYearStr)
	if err != nil {
		log.Error(ctx, "failed to convert filter start year string to integer", err)
		return eTag, err
	}
	endYearInt, err := strconv.Atoi(endYearStr)
	if err != nil {
		log.Error(ctx, "failed to convert filter end year string to integer", err)
		return eTag, err
	}

	selectedMonths := req.Form["months"]

	for year := startYearInt; year <= endYearInt; year++ {
		yearStr := strconv.Itoa(year)
		for _, month := range selectedMonths {
			monthYearComboStr := fmt.Sprintf("%s %s", month, yearStr)
			monthYearComboTime, err := time.Parse("January 2006", monthYearComboStr)
			if err != nil {
				log.Error(ctx, "failed to convert filtered month and year combo to time format", err)
				return eTag, err
			}
			monthYearCombo := monthYearComboTime.Format("Jan-06")
			options = append(options, monthYearCombo)
		}
	}

	newETag, err = f.FilterClient.SetDimensionValues(ctx, userAccessToken, "", collectionID, filterID, dimensionName, options, eTag)
	if err != nil {
		log.Error(ctx, "failed to add dimension values", err)
		return eTag, err
	}
	return newETag, nil
}

func (f *Filter) addTimeRange(filterID, userAccessToken, collectionID string, req *http.Request, eTag string) (newETag string, err error) {
	startMonth := req.Form.Get("start-month")
	startYear := req.Form.Get("start-year")
	endMonth := req.Form.Get("end-month")
	endYear := req.Form.Get("end-year")
	ctx := req.Context()
	dimensionName := "time"

	values, labelIDMap, err := f.getDimensionValues(ctx, userAccessToken, collectionID, filterID, dimensionName)
	if err != nil {
		return "", err
	}

	dats, err := dates.ConvertToReadable(values)
	if err != nil {
		return "", err
	}
	dats = dates.Sort(dats)

	start, err := time.Parse("01 January 2006", fmt.Sprintf("01 %s %s", startMonth, startYear))
	if err != nil {
		return "", err
	}

	end, err := time.Parse("01 January 2006", fmt.Sprintf("01 %s %s", endMonth, endYear))
	if err != nil {
		return "", err
	}

	if end.Before(start) {
		return "", fmt.Errorf("start date: %s before end date: %s", start.String(), end.String())
	}

	values = dates.ConvertToCoded(dats)
	var options []string
	for i, dat := range dats {
		if dat.Equal(start) || dat.After(start) && dat.Before(end) || dat.Equal(end) {
			options = append(options, labelIDMap[values[i]])
		}
	}

	return f.FilterClient.SetDimensionValues(ctx, userAccessToken, "", collectionID, filterID, dimensionName, options, eTag)
}

// Time specifically handles the data for the time dimension page
func (f *Filter) Time() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		ctx := req.Context()
		dimensionName := "time"

		fj, eTag0, err := f.FilterClient.GetJobState(ctx, userAccessToken, "", "", collectionID, filterID)
		if err != nil {
			log.Error(ctx, "failed to get job state", err, log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		versionURL, err := url.Parse(fj.Links.Version.HRef)
		if err != nil {
			log.Error(ctx, "failed to parse version href", err, log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}
		versionPath := strings.TrimPrefix(versionURL.Path, f.APIRouterVersion)
		datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
		if err != nil {
			log.Error(ctx, "failed to extract dataset info from path", err, log.Data{"filter_id": filterID, "path": versionPath})
			setStatusCode(req, w, err)
			return
		}

		datasetDetails, err := f.DatasetClient.Get(ctx, userAccessToken, "", collectionID, datasetID)
		if err != nil {
			log.Error(ctx, "failed to get dataset", err, log.Data{"dataset_id": datasetID})
			setStatusCode(req, w, err)
			return
		}
		ver, err := f.DatasetClient.GetVersion(ctx, userAccessToken, "", "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Error(ctx, "failed to get version", err, log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		// count number of options for the dimension in dataset API
		opts, err := f.DatasetClient.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dimensionName, &dataset.QueryParams{Offset: 0, Limit: 1})
		if err != nil {
			log.Error(ctx, "failed to get options from dataset client", err,
				log.Data{"dimension": dimensionName, "dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		//use normal list format unless a specially recognized time format
		if opts.TotalCount <= MaxNumOptionsOnPage || !acceptedReg.MatchString(opts.Items[0].Option) {
			mux.Vars(req)["name"] = dimensionName
			f.DimensionSelector().ServeHTTP(w, req)
			return
		}

		dims, err := f.DatasetClient.GetVersionDimensions(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Error(ctx, "failed to get dimensions", err,
				log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		selValues, eTag1, err := f.FilterClient.GetDimensionOptionsInBatches(ctx, userAccessToken, "", collectionID, filterID, dimensionName, f.BatchSize, f.BatchMaxWorkers)
		if err != nil {
			log.Error(ctx, "failed to get options from filter client", err, log.Data{"filter_id": filterID, "dimension": dimensionName})
			setStatusCode(req, w, err)
			return
		}

		if eTag0 != eTag1 {
			err := errors.New("inconsistent filter data")
			log.Error(ctx, "data consistency cannot be guaranteed because filter was modified between calls", err,
				log.Data{"filter_id": filterID, "e_tag_0": eTag0, "e_tag_1": eTag1})
			// The user might want to retry this handler in this case
			setStatusCode(req, w, err)
			return
		}

		allValues, err := f.DatasetClient.GetOptionsInBatches(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dimensionName, f.BatchSize, f.BatchMaxWorkers)
		if err != nil {
			log.Error(ctx, "failed to get options from dataset client", err,
				log.Data{"dimension": dimensionName, "dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		p, err := mapper.CreateTimePage(req, fj, datasetDetails, ver, allValues, selValues.Items, dims, datasetID, f.APIRouterVersion, lang)
		if err != nil {
			log.Error(ctx, "failed to map data to page", err, log.Data{"filter_id": filterID, "dataset_id": datasetID, "dimension": dimensionName})
			setStatusCode(req, w, err)
			return
		}

		b, err := json.Marshal(p)
		if err != nil {
			log.Error(ctx, "failed to marshal json", err, log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		templateBytes, err := f.Renderer.Do("dataset-filter/time", b)
		if err != nil {
			log.Error(ctx, "failed to render", err, log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		if _, err := w.Write(templateBytes); err != nil {
			log.Error(ctx, "failed to write response", err, log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}
	})

}
