package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

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
			if err := f.FilterClient.AddDimensionValue(ctx, userAccessToken, "", collectionID, filterID, dimensionName, req.Form.Get("all-ages-option")); err != nil {
				log.Event(ctx, "failed to add all ages option", log.WARN, log.Error(err), log.Data{"age_case": "all"})
			}
		case "range":
			if err := f.addAgeRange(filterID, req); err != nil {
				log.Event(ctx, "failed to add age range", log.WARN, log.Error(err), log.Data{"age_case": "range"})
			}
		case "list":
			if err := f.addAgeList(filterID, req); err != nil {
				log.Event(ctx, "failed to add age list", log.WARN, log.Error(err), log.Data{"age_case": "list"})
			}
		}

		redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)
		http.Redirect(w, req, redirectURL, 302)
	})
}

func (f *Filter) addAgeList(filterID string, req *http.Request) error {
	ctx := req.Context()
	collectionID := getCollectionIDFromContext(ctx)
	dimensionName := "age"
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
	// Remove any unselected ages
	for _, opt := range opts {
		if _, ok := req.Form[opt.Option]; !ok {
			if err := f.FilterClient.RemoveDimensionValue(ctx, userAccessToken, "", collectionID, filterID, dimensionName, opt.Option); err != nil {
				log.Event(ctx, "failed to remove dimension options", log.WARN, log.Error(err))
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

	if err := f.FilterClient.AddDimensionValues(ctx, userAccessToken, "", collectionID, filterID, dimensionName, options); err != nil {
		log.Event(ctx, "failed to add dimension options", log.ERROR, log.Error(err))
	}

	return nil
}

func (f *Filter) addAgeRange(filterID string, req *http.Request) error {
	youngest := req.Form.Get("youngest")
	oldest := req.Form.Get("oldest")

	reg := regexp.MustCompile(`\d+\+`)
	ctx := req.Context()

	dimensionName := "age"

	oldestHasPlus := reg.MatchString(oldest)
	if oldestHasPlus {
		oldest = strings.Trim(oldest, "+")
	}
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
	return f.FilterClient.AddDimensionValues(ctx, userAccessToken, "", collectionID, filterID, dimensionName, options)
}

// Age is a handler which will create age values on a filter job
func (f *Filter) Age() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		ctx := req.Context()
		dimensionName := "age"

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
		datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(ctx, versionURL.Path)
		if err != nil {
			log.Event(ctx, "failed to extract dataset info from path", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "path": versionURL})
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

		if len(allValues.Items) <= 20 {
			mux.Vars(req)["name"] = dimensionName
			f.DimensionSelector()
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

		p, err := mapper.CreateAgePage(req, fj, dataset, ver, allValues, selValues, dims, datasetID)
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
