package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/headers"
	dphandlers "github.com/ONSdigital/dp-net/handlers"
	"github.com/ONSdigital/log.go/log"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/gorilla/mux"
)

// UpdateAge is a handler which will update age values on a filter job
func (f *Filter) UpdateAge() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		ctx := req.Context()
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		dimensionName := "age"

		eTag, err := f.FilterClient.RemoveDimension(ctx, userAccessToken, "", collectionID, filterID, dimensionName, headers.IfMatchAnyETag)
		if err != nil {
			log.Event(ctx, "failed to remove dimension", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dimensionName})
			setStatusCode(req, w, err)
			return
		}
		eTag, err = f.FilterClient.AddDimension(ctx, userAccessToken, "", collectionID, filterID, dimensionName, eTag)
		if err != nil {
			log.Event(ctx, "failed to add dimension", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dimensionName})
			setStatusCode(req, w, err)
			return
		}

		if err := req.ParseForm(); err != nil {
			log.Event(ctx, "failed to parse form", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "e_tag": eTag})
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

		log.Event(ctx, "age-selection", log.INFO, log.Data{dimensionName: req.Form.Get("age-selection")})
		switch req.Form.Get("age-selection") {
		case "all":
			eTag, err = f.FilterClient.AddDimensionValue(ctx, userAccessToken, "", collectionID, filterID, dimensionName, req.Form.Get("all-ages-option"), eTag)
			if err != nil {
				log.Event(ctx, "failed to add all ages option", log.WARN, log.Error(err), log.Data{"age_case": "all"})
			}
		case "range":
			eTag, err = f.addAgeRange(filterID, userAccessToken, collectionID, req, eTag)
			if err != nil {
				log.Event(ctx, "failed to add age range", log.WARN, log.Error(err), log.Data{"age_case": "range"})
			}
		case "list":
			eTag, err = f.addAgeList(filterID, userAccessToken, collectionID, req, eTag)
			if err != nil {
				log.Event(ctx, "failed to add age list", log.WARN, log.Error(err), log.Data{"age_case": "list"})
			}
		}

		redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)
		http.Redirect(w, req, redirectURL, 302)
	})
}

func (f *Filter) addAgeList(filterID, userAccessToken, collectionID string, req *http.Request, eTag string) (newETag string, err error) {
	ctx := req.Context()
	dimensionName := "age"

	var options []string
	for k := range req.Form {
		if _, err := strconv.Atoi(k); err != nil {
			if !strings.Contains(k, "+") {
				continue
			}
		}

		options = append(options, k)
	}

	newETag, err = f.FilterClient.SetDimensionValues(ctx, userAccessToken, "", collectionID, filterID, dimensionName, options, eTag)
	if err != nil {
		log.Event(ctx, "failed to add dimension options", log.ERROR, log.Error(err))
	}

	return newETag, nil
}

func (f *Filter) addAgeRange(filterID, userAccessToken, collectionID string, req *http.Request, eTag string) (newETag string, err error) {
	youngest := req.Form.Get("youngest")
	oldest := req.Form.Get("oldest")

	reg := regexp.MustCompile(`\d+\+`)
	ctx := req.Context()

	dimensionName := "age"

	oldestHasPlus := reg.MatchString(oldest)
	if oldestHasPlus {
		oldest = strings.Trim(oldest, "+")
	}

	values, labelIDMap, err := f.getDimensionValues(ctx, userAccessToken, collectionID, filterID, dimensionName)
	if err != nil {
		return "", err
	}

	var intValues []int
	for _, val := range values {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			hasPlus := reg.MatchString(val)
			if hasPlus {
				val = strings.Trim(oldest, "+")
				intVal, err = strconv.Atoi(val)
				if err != nil {
					break
				}
			}
		}

		intValues = append(intValues, intVal)
	}

	sort.Ints(intValues)

	for i, val := range intValues {
		values[i] = strconv.Itoa(val)
	}

	var isInRange bool
	var options []string
	for i, age := range values {
		if i == len(values)-1 && oldestHasPlus {
			age = age + "+"
		}

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
	return f.FilterClient.SetDimensionValues(ctx, userAccessToken, "", collectionID, filterID, dimensionName, options, eTag)
}

// Age is a handler which will create age values on a filter job
func (f *Filter) Age() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		ctx := req.Context()
		dimensionName := "age"

		fj, eTag0, err := f.FilterClient.GetJobState(ctx, userAccessToken, "", "", collectionID, filterID)
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

		datasetDetails, err := f.DatasetClient.Get(ctx, userAccessToken, "", collectionID, datasetID)
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

		// count number of options for the dimension in dataset API
		opts, err := f.DatasetClient.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dimensionName,
			&dataset.QueryParams{Offset: 0, Limit: 0})
		if err != nil {
			log.Event(ctx, "failed to get options from dataset client", log.ERROR, log.Error(err),
				log.Data{"dimension": dimensionName, "dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		if opts.TotalCount <= MaxNumOptionsOnPage {
			mux.Vars(req)["name"] = dimensionName
			f.DimensionSelector().ServeHTTP(w, req)
			return
		}

		dims, err := f.DatasetClient.GetVersionDimensions(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Event(ctx, "failed to get dimensions", log.ERROR, log.Error(err),
				log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		selValues, eTag1, err := f.FilterClient.GetDimensionOptionsInBatches(ctx, userAccessToken, "", collectionID, filterID, dimensionName, f.BatchSize, f.BatchMaxWorkers)
		if err != nil {
			log.Event(ctx, "failed to get options from filter client", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": dimensionName})
			setStatusCode(req, w, err)
			return
		}

		// The user might want to retry this handler if eTags don't match
		if eTag0 != eTag1 {
			err := errors.New("inconsistent filter data")
			log.Event(ctx, "data consistency cannot be guaranteed because filter was modified between calls", log.ERROR, log.Error(err),
				log.Data{"filter_id": filterID, "dimension": dimensionName, "e_tag_0": eTag0, "e_tag_1": eTag1})
			setStatusCode(req, w, err)
			return
		}

		allValues, err := f.DatasetClient.GetOptionsInBatches(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dimensionName, f.BatchSize, f.BatchMaxWorkers)
		if err != nil {
			log.Event(ctx, "failed to get options from dataset client", log.ERROR, log.Error(err),
				log.Data{"dimension": dimensionName, "dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		p, err := mapper.CreateAgePage(req, fj, datasetDetails, ver, allValues, selValues, dims, datasetID, f.APIRouterVersion, lang)
		if err != nil {
			log.Event(ctx, "failed to map data to page", log.ERROR, log.Error(err),
				log.Data{"filter_id": filterID, "dataset_id": datasetID, "dimension": dimensionName})
			setStatusCode(req, w, err)
			return
		}

		b, err := json.Marshal(p)
		if err != nil {
			log.Event(ctx, "failed to marshal json", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		templateBytes, err := f.Renderer.Do("dataset-filter/age", b)
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
	})
}
