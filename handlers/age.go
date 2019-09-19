package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// UpdateAge is a handler which will update age values on a filter job
func (f *Filter) UpdateAge(w http.ResponseWriter, req *http.Request) {
	cfg := config.Get()
	ctx := req.Context()
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	req = forwardFlorenceTokenIfRequired(req)
	if err := f.FilterClient.RemoveDimension(req.Context(), cfg.UserAuthToken, cfg.ServiceAuthToken, filterID, "age"); err != nil {
		log.InfoCtx(ctx, "failed to remove dimension", log.Data{"error": err, "filter_id": filterID, "dimension": "age"})
		setStatusCode(req, w, err)
		return
	}
	if err := f.FilterClient.AddDimension(req.Context(), cfg.UserAuthToken, cfg.ServiceAuthToken, filterID, "age"); err != nil {
		log.InfoCtx(ctx, "failed to add dimension", log.Data{"error": err, "filter_id": filterID, "dimension": "age"})
		setStatusCode(req, w, err)
		return
	}

	if err := req.ParseForm(); err != nil {
		log.InfoCtx(ctx, "failed to parse form", log.Data{"error": err, "filter_id": filterID})
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

	log.InfoCtx(ctx, "age-selection", log.Data{"age": req.Form.Get("age-selection")})
	switch req.Form.Get("age-selection") {
	case "all":
		if err := f.FilterClient.AddDimensionValue(req.Context(), cfg.UserAuthToken, cfg.ServiceAuthToken, filterID,"age", req.Form.Get("all-ages-option")); err != nil {
			log.ErrorCtx(ctx, err, log.Data{"age_case": "all"})
		}
	case "range":
		if err := f.addAgeRange(filterID, req); err != nil {
			log.ErrorCtx(ctx, err, log.Data{"age_case": "range"})
		}
	case "list":
		if err := f.addAgeList(filterID, req); err != nil {
			log.ErrorCtx(ctx, err, log.Data{"age_case": "list"})
		}
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)
	http.Redirect(w, req, redirectURL, 302)
}

func (f *Filter) addAgeList(filterID string, req *http.Request) error {
	cfg := config.Get()
	req = forwardFlorenceTokenIfRequired(req)
	ctx := req.Context()
	opts, err := f.FilterClient.GetDimensionOptions(req.Context(), cfg.UserAuthToken, cfg.ServiceAuthToken, filterID, "age")
	if err != nil {
		return err
	}
	// Remove any unselected ages
	for _, opt := range opts {
		if _, ok := req.Form[opt.Option]; !ok {
			if err := f.FilterClient.RemoveDimensionValue(req.Context(), cfg.UserAuthToken, cfg.ServiceAuthToken, filterID, "age", opt.Option); err != nil {
				log.ErrorCtx(ctx, err, nil)
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

	if err := f.FilterClient.AddDimensionValues(req.Context(), cfg.UserAuthToken, cfg.ServiceAuthToken, filterID, "age", options); err != nil {
		log.InfoCtx(ctx, err.Error(), nil)
	}

	return nil
}

func (f *Filter) addAgeRange(filterID string, req *http.Request) error {
	cfg := config.Get()
	youngest := req.Form.Get("youngest")
	oldest := req.Form.Get("oldest")

	reg := regexp.MustCompile(`\d+\+`)

	oldestHasPlus := reg.MatchString(oldest)
	if oldestHasPlus {
		oldest = strings.Trim(oldest, "+")
	}

	req = forwardFlorenceTokenIfRequired(req)

	values, labelIDMap, err := f.getDimensionValues(req.Context(), filterID, "age")
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
	return f.FilterClient.AddDimensionValues(req.Context(), cfg.UserAuthToken, cfg.ServiceAuthToken, filterID, "age", options)
}

func (f *Filter) Age(w http.ResponseWriter, req *http.Request) {
	cfg := config.Get()
	vars := mux.Vars(req)
	filterID := vars["filterID"]
	ctx := req.Context()

	req = forwardFlorenceTokenIfRequired(req)
	fj, err := f.FilterClient.GetJobState(req.Context(), cfg.UserAuthToken, cfg.ServiceAuthToken, cfg.DownloadAuthToken, filterID)
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

	dataset, err := f.DatasetClient.Get(req.Context(), datasetID)
	if err != nil {
		log.InfoCtx(ctx, "failed to get dataset", log.Data{"error": err, "dataset_id": datasetID})
		setStatusCode(req, w, err)
		return
	}
	ver, err := f.DatasetClient.GetVersion(req.Context(), datasetID, edition, version)
	if err != nil {
		log.InfoCtx(ctx, "failed to get version", log.Data{"error": err, "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	allValues, err := f.DatasetClient.GetOptions(req.Context(), datasetID, edition, version, "age")
	if err != nil {
		log.InfoCtx(ctx, "failed to get options from dataset client",
			log.Data{"error": err, "dimension": "age", "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	if len(allValues.Items) <= 20 {
		mux.Vars(req)["name"] = "age"
		f.DimensionSelector(w, req)
		return
	}

	selValues, err := f.FilterClient.GetDimensionOptions(req.Context(), cfg.UserAuthToken, cfg.ServiceAuthToken, filterID, "age")
	if err != nil {
		log.InfoCtx(ctx, "failed to get options from filter client", log.Data{"error": err, "filter_id": filterID, "dimension": "age"})
		setStatusCode(req, w, err)
		return
	}
	dims, err := f.DatasetClient.GetDimensions(req.Context(), datasetID, edition, version)
	if err != nil {
		log.InfoCtx(ctx, "failed to get dimensions",
			log.Data{"error": err, "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	p, err := mapper.CreateAgePage(req.Context(), fj, dataset, ver, allValues, selValues, dims, datasetID)
	if err != nil {
		log.InfoCtx(ctx, "failed to map data to page", log.Data{"error": err, "filter_id": filterID, "dataset_id": datasetID, "dimension": "age"})
		setStatusCode(req, w, err)
		return
	}

	b, err := json.Marshal(p)
	if err != nil {
		log.InfoCtx(ctx, "failed to marshal json", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/age", b)
	if err != nil {
		log.InfoCtx(ctx, "failed to render", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	if _, err := w.Write(templateBytes); err != nil {
		log.InfoCtx(req.Context(), "failed to write response", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

}
