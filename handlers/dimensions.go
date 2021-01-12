package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/hierarchy"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	dphandlers "github.com/ONSdigital/dp-net/handlers"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

type labelID struct {
	Label string `json:"label"`
	ID    string `json:"id"`
}

// GetAllDimensionOptionsJSON will return a list of all options from the dataset api
func (f *Filter) GetAllDimensionOptionsJSON() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		name := vars["name"]
		filterID := vars["filterID"]
		ctx := req.Context()

		fj, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
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

		idNameMap, err := f.getIDNameMap(req.Context(), userAccessToken, collectionID, versionPath, name)
		if err != nil {
			log.Event(ctx, "failed to get name map", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "path": versionPath, "name": name})
			setStatusCode(req, w, err)
			return
		}

		var lids []labelID

		if name == "time" {

			var codedDates []string
			labelIDMap := make(map[string]string)
			for k, v := range idNameMap {
				codedDates = append(codedDates, v)
				labelIDMap[v] = k
			}

			readableDates, err := dates.ConvertToReadable(codedDates)
			if err != nil {
				log.Event(ctx, "failed to convert dates", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dates": codedDates})
				setStatusCode(req, w, err)
				return
			}

			readableDates = dates.Sort(readableDates)

			for _, date := range readableDates {
				lid := labelID{
					Label: fmt.Sprintf("%s %d", date.Month(), date.Year()),
					ID:    labelIDMap[date.Format("Jan-06")],
				}

				lids = append(lids, lid)
			}
		}

		b, err := json.Marshal(lids)
		if err != nil {
			log.Event(ctx, "failed to marshal json", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}

		w.Write(b)
	})

}

func (f *Filter) getIDNameMap(ctx context.Context, userAccessToken, collectionID, versionURL, dimension string) (map[string]string, error) {
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(ctx, versionURL)

	idNameMap := make(map[string]string)

	opts, err := f.DatasetClient.GetOptionsInBatches(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dimension, f.BatchSize, f.BatchMaxWorkers)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts.Items {
		idNameMap[opt.Option] = opt.Label
	}

	return idNameMap, nil
}

// GetSelectedDimensionOptionsJSON will return a list of selected options from the filter api with corresponding label
func (f *Filter) GetSelectedDimensionOptionsJSON() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		name := vars["name"]
		filterID := vars["filterID"]
		ctx := req.Context()

		opts, err := f.FilterClient.GetDimensionOptionsInBatches(req.Context(), userAccessToken, "", collectionID, filterID, name, f.BatchSize, f.BatchMaxWorkers)
		if err != nil {
			log.Event(ctx, "failed to get dimension options", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}

		fj, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
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
		idNameMap, err := f.getIDNameMap(req.Context(), userAccessToken, collectionID, versionPath, name)
		if err != nil {
			log.Event(ctx, "failed to get name map", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "path": versionPath, "name": name})
			setStatusCode(req, w, err)
			return
		}

		var lids []labelID

		if name == "time" {

			var codedDates []string
			labelIDMap := make(map[string]string)
			for _, opt := range opts.Items {
				codedDates = append(codedDates, idNameMap[opt.Option])
				labelIDMap[idNameMap[opt.Option]] = opt.Option
			}

			readableDates, err := dates.ConvertToReadable(codedDates)
			if err != nil {
				log.Event(ctx, "failed to convert dates", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dates": codedDates})
				setStatusCode(req, w, err)
				return
			}

			readableDates = dates.Sort(readableDates)

			for _, date := range readableDates {
				lid := labelID{
					Label: fmt.Sprintf("%s %d", date.Month(), date.Year()),
					ID:    labelIDMap[date.Format("Jan-06")],
				}

				lids = append(lids, lid)
			}
		}

		b, err := json.Marshal(lids)
		if err != nil {
			log.Event(ctx, "failed to marshal json", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		w.Write(b)
	})

}

// DimensionSelector controls the render of the range selector template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) DimensionSelector() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {

		var tAllVals, tSelVals time.Duration
		t0 := time.Now()

		logTime := func() {
			log.Event(nil, "+++ PERFORMANCE TEST", log.Data{
				"method":              "dimensions.DimensionSelector",
				"whole":               time.Since(t0),
				"get_dataset_options": tAllVals,
				"get_filter_options":  tSelVals,
			})
		}

		defer logTime()

		vars := mux.Vars(req)
		name := vars["name"]
		filterID := vars["filterID"]
		ctx := req.Context()

		fj, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
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

		datasetDetails, err := f.DatasetClient.Get(req.Context(), userAccessToken, "", collectionID, datasetID)
		if err != nil {
			log.Event(ctx, "failed to get dataset", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID})
			setStatusCode(req, w, err)
			return
		}

		ver, err := f.DatasetClient.GetVersion(req.Context(), userAccessToken, "", "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Event(ctx, "failed to get version", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		// TODO: This is a shortcut for now, if the hierarchy api returns a status 200
		// then the dimension should be populated as a hierarchy
		isHierarchy, err := f.isHierarchicalDimension(ctx, fj.InstanceID, name)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		// count number of options for the dimension in dataset API
		opts, err := f.DatasetClient.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, name, dataset.QueryParams{Offset: 0, Limit: 1})
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		// if there are more than maxNumOptionsOnPage, then we need to use the hierarchy model
		if isHierarchy && opts.TotalCount > MaxNumOptionsOnPage {
			f.Hierarchy().ServeHTTP(w, req)
			return
		}

		dims, err := f.DatasetClient.GetVersionDimensions(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Event(ctx, "failed to get dimensions", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		t1 := time.Now()
		selectedValues, err := f.FilterClient.GetDimensionOptionsInBatches(req.Context(), userAccessToken, "", collectionID, filterID, name, f.BatchSize, f.BatchMaxWorkers)
		if err != nil {
			log.Event(ctx, "failed to get options from filter client", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}
		tSelVals = time.Since(t1)

		t2 := time.Now()
		allValues, err := f.DatasetClient.GetOptionsInBatches(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version, name, f.BatchSize, f.BatchMaxWorkers)
		if err != nil {
			log.Event(ctx, "failed to get options from dataset client", log.ERROR, log.Error(err), log.Data{"dimension": name, "dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}
		tAllVals = time.Since(t2)

		if name == "time" {
			allValues = sortedTime(ctx, allValues)
		}

		f.listSelector(w, req, name, selectedValues.Items, allValues, fj, datasetDetails, dims, datasetID, ver.ReleaseDate, lang)
	})

}

func (f *Filter) isHierarchicalDimension(ctx context.Context, instanceID, dimensionName string) (bool, error) {
	_, err := f.HierarchyClient.GetRoot(ctx, instanceID, dimensionName)
	if err != nil {

		var getHierarchyErr *hierarchy.ErrInvalidHierarchyAPIResponse
		if errors.As(err, &getHierarchyErr) && http.StatusNotFound == getHierarchyErr.Code() {
			return false, nil
		}

		log.Event(ctx, "unexpected error getting hierarchy root for dimension", log.ERROR, log.Error(err), log.Data{
			"instance_id":    instanceID,
			"dimension_name": dimensionName,
		})

		return false, err
	}

	return true, nil
}

type sorting struct {
	substring string
	option    dataset.Option
}

//sort time chronologically, for code lists which match the format mmm-mmm-yyyy
func sortedTime(ctx context.Context, opts dataset.Options) dataset.Options {
	if &opts == nil || len(opts.Items) == 0 {
		return opts
	}

	validMonths := map[string]int{
		"jan":       1,
		"january":   1,
		"feb":       2,
		"february":  2,
		"mar":       3,
		"march":     3,
		"apr":       4,
		"april":     4,
		"may":       5,
		"jun":       6,
		"june":      6,
		"jul":       7,
		"july":      7,
		"aug":       8,
		"august":    8,
		"sep":       9,
		"september": 9,
		"oct":       10,
		"october":   10,
		"nov":       11,
		"november":  11,
		"dec":       12,
		"december":  12,
	}

	output := make(map[string][]sorting)
	for _, o := range opts.Items {
		if &o.Links == nil || &o.Links.Code == nil {
			log.Event(ctx, "options list does not contain code ids so cannot be sorted", log.WARN)
			break
		}
		// these codes are mmm-mmm-yyyy where the second month relates to the year
		// e.g. `nov-jan-2014` means november 2013 - january 2014
		// so to sort chronologically we must refer to the second month mentioned
		month, year, err := splitCode(o.Links.Code.ID)
		if err != nil {
			log.Event(ctx, "option format is not sortable, returning flat list", log.WARN, log.Data{"code": o.Links.Code.ID})
			break
		}

		var monthOrder int
		var ok bool
		if monthOrder, ok = validMonths[month]; !ok {
			log.Event(ctx, "time does not follow an understood format so cannot be sorted", log.WARN, log.Data{"lookup": month, "code": o.Links.Code.ID})
			break
		}

		sortedOptions := output[year]

		if len(sortedOptions) == 0 {
			sortedOptions = append(sortedOptions, sorting{month, o})
			output[year] = sortedOptions
			continue
		}

		//insert the month in the correct place according to its order number
		for i, s := range sortedOptions {
			if validMonths[s.substring] < monthOrder {
				if len(sortedOptions)-1 > i {
					continue
				}

				//at end of list, add to end
				sortedOptions = append(sortedOptions, sorting{month, o})
				break
			}

			sortedOptions = append(sortedOptions, sorting{})
			copy(sortedOptions[i+1:], sortedOptions[i:])
			sortedOptions[i] = sorting{month, o}
			break
		}

		output[year] = sortedOptions
	}

	//get the years from the map and sort them
	keys := []string{}
	for y := range output {
		keys = append(keys, y)
	}
	sort.Strings(keys)

	//flatten the map of sorted lists into one super-sorted list
	newList := []dataset.Option{}
	for _, key := range keys {
		l := output[key]

		for _, o := range l {
			newList = append(newList, o.option)
		}
	}
	//check lists are the same length (so contain the same data) and only
	//return the sorted list if it's complete
	if len(newList) == len(opts.Items) {
		opts.Items = newList
	}
	return opts
}

func splitCode(id string) (string, string, error) {
	code := strings.Split(id, "-")
	if len(code) == 1 {
		return "", "", errors.New("code cannot be split")
	}

	if len(code) < 3 {
		return "", "", errors.New("code does not match expected format")
	}

	month := code[len(code)-2]
	month = strings.ToLower(month)

	year := code[len(code)-1]
	year = strings.ToLower(year)

	return month, year, nil
}

// ListSelector controls the render of the age selector list template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) listSelector(w http.ResponseWriter, req *http.Request, name string, selectedValues []filter.DimensionOption, allValues dataset.Options, filter filter.Model, dataset dataset.DatasetDetails, dims dataset.VersionDimensions, datasetID, releaseDate, lang string) {
	ctx := req.Context()
	p := mapper.CreateListSelectorPage(req, name, selectedValues, allValues, filter, dataset, dims, datasetID, releaseDate, f.APIRouterVersion, lang)

	b, err := json.Marshal(p)
	if err != nil {
		log.Event(ctx, "failed to marshal json", log.ERROR, log.Error(err), log.Data{"filter_id": filter.FilterID})
		setStatusCode(req, w, err)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/list-selector", b)
	if err != nil {
		log.Event(ctx, "failed to render", log.ERROR, log.Error(err), log.Data{"filter_id": filter.FilterID})
		setStatusCode(req, w, err)
		return
	}

	w.Write(templateBytes)
}

// DimensionAddAll will add all dimension values to a basket
func (f *Filter) DimensionAddAll() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		name := vars["name"]
		filterID := vars["filterID"]
		f.addAll(w, req, fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name), collectionID, userAccessToken)
	})
}

func (f *Filter) addAll(w http.ResponseWriter, req *http.Request, redirectURL, userAccessToken, collectionID string) {

	var tOptionsBatchPatch time.Duration
	t0 := time.Now()

	logTime := func() {
		log.Event(nil, "+++ PERFORMANCE TEST", log.Data{
			"method":              "dimensions.addAll",
			"whole":               time.Since(t0),
			"options_batch_patch": tOptionsBatchPatch,
		})
	}

	defer logTime()

	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	ctx := req.Context()

	fj, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
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

	// function to add each batch of dataset dimension options to filter API
	processBatch := func(batch dataset.Options) (forceAbort bool, err error) {
		var options []string
		for _, item := range batch.Items {
			options = append(options, item.Option)
		}
		// first batch, will overwrite any existing values in filter API
		if batch.Offset == 0 {
			return false, f.FilterClient.SetDimensionValues(req.Context(), userAccessToken, "", collectionID, filterID, name, options)
		}
		// the rest of batches will be added to the existing items in filter API via patch operations
		return false, f.FilterClient.PatchDimensionValues(req.Context(), userAccessToken, "", collectionID, filterID, name, options, []string{}, f.BatchSize)
	}

	t1 := time.Now()
	// call dataset API GetOptions in batches, and process each batch to add the options to filter API
	if err := f.DatasetClient.GetOptionsBatchProcess(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version, name, nil, processBatch, f.BatchSize, f.BatchMaxWorkers); err != nil {
		log.Event(ctx, "failed to process options from dataset api", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
		setStatusCode(req, w, err)
		return
	}
	tOptionsBatchPatch = time.Since(t1)

	http.Redirect(w, req, redirectURL, 302)

}

// AddList sets a list of values, removing any existing value.
func (f *Filter) AddList() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		name := vars["name"]
		filterID := vars["filterID"]
		ctx := req.Context()

		if err := req.ParseForm(); err != nil {
			log.Event(ctx, "failed to parse form", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}

		redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)

		if len(req.Form["add-all"]) > 0 {
			redirectURL = fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
			f.addAll(w, req, redirectURL, userAccessToken, collectionID)
			return
		}

		if len(req.Form["remove-all"]) > 0 {
			redirectURL = fmt.Sprintf("/filters/%s/dimensions/%s/remove-all", filterID, name)
			http.Redirect(w, req, redirectURL, 302)
			return
		}

		var options []string
		for k := range req.Form {
			if k == ":uri" || k == "save-and-return" {
				continue
			}

			options = append(options, k)
		}

		if err := f.FilterClient.SetDimensionValues(ctx, userAccessToken, "", collectionID, filterID, name, options); err != nil {
			log.Event(ctx, "failed to add dimension values", log.WARN, log.Error(err))
		}

		http.Redirect(w, req, redirectURL, 302)
	})

}

func (f *Filter) getDimensionValues(ctx context.Context, userAccessToken, collectionID, filterID, name string) (values []string, labelIDMap map[string]string, err error) {

	fj, err := f.FilterClient.GetJobState(ctx, userAccessToken, "", "", collectionID, filterID)
	if err != nil {
		return
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		return
	}
	versionPath := strings.TrimPrefix(versionURL.Path, f.APIRouterVersion)

	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(ctx, versionPath)
	if err != nil {
		return
	}

	vals, err := f.DatasetClient.GetOptionsInBatches(ctx, userAccessToken, "", collectionID, datasetID, edition, version, name, f.BatchSize, f.BatchMaxWorkers)
	if err != nil {
		return
	}

	labelIDMap = make(map[string]string)
	for _, val := range vals.Items {
		values = append(values, val.Label)
		labelIDMap[val.Label] = val.Option
	}

	return
}

// DimensionRemoveAll removes all options on a particular dimensions
func (f *Filter) DimensionRemoveAll() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		name := vars["name"]
		filterID := vars["filterID"]
		ctx := req.Context()

		log.Event(ctx, "attempting to remove all options from dimension", log.INFO, log.Data{"dimension": name, "filterID": filterID})

		if err := f.FilterClient.RemoveDimension(req.Context(), userAccessToken, "", collectionID, filterID, name); err != nil {
			log.Event(ctx, "failed to remove dimension", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}

		if err := f.FilterClient.AddDimension(req.Context(), userAccessToken, "", collectionID, filterID, name); err != nil {
			log.Event(ctx, "failed to add dimension", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}

		redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
		http.Redirect(w, req, redirectURL, 302)
	})

}

// DimensionRemoveOne removes an individual option on a dimensions
func (f *Filter) DimensionRemoveOne() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		name := vars["name"]
		filterID := vars["filterID"]
		option := vars["option"]
		ctx := req.Context()

		if err := f.FilterClient.RemoveDimensionValue(req.Context(), userAccessToken, "", collectionID, filterID, name, option); err != nil {
			log.Event(ctx, "failed to remove dimension option", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name, "option": option})
			setStatusCode(req, w, err)
			return
		}

		redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
		http.Redirect(w, req, redirectURL, 302)
	})

}
