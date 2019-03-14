package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

type labelID struct {
	Label string `json:"label"`
	ID    string `json:"id"`
}

// GetAllDimensionOptionsJSON will return a list of all options from the dataset api
func (f *Filter) GetAllDimensionOptionsJSON(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	ctx := req.Context()

	req = forwardFlorenceTokenIfRequired(req)

	fj, err := f.FilterClient.GetJobState(req.Context(), filterID)
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

	idNameMap, err := f.getIDNameMap(req.Context(), versionURL.Path, name)
	if err != nil {
		log.InfoCtx(ctx, "failed to get name map", log.Data{"error": err, "filter_id": filterID, "path": versionURL, "name": name})
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

		readbleDates, err := dates.ConvertToReadable(codedDates)
		if err != nil {
			log.InfoCtx(ctx, "failed to convert dates", log.Data{"error": err, "filter_id": filterID, "dates": codedDates})
			setStatusCode(req, w, err)
			return
		}

		readbleDates = dates.Sort(readbleDates)

		for _, date := range readbleDates {
			lid := labelID{
				Label: fmt.Sprintf("%s %d", date.Month(), date.Year()),
				ID:    labelIDMap[date.Format("Jan-06")],
			}

			lids = append(lids, lid)
		}
	}

	b, err := json.Marshal(lids)
	if err != nil {
		log.InfoCtx(ctx, "failed to marshal json", log.Data{"error": err, "filter_id": filterID, "dimension": name})
		setStatusCode(req, w, err)
		return
	}

	w.Write(b)
}

func (f *Filter) getIDNameMap(ctx context.Context, versionURL, dimension string) (map[string]string, error) {
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL)
	if err != nil {
		return nil, err
	}

	idNameMap := make(map[string]string)
	opts, err := f.DatasetClient.GetOptions(ctx, datasetID, edition, version, dimension)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts.Items {
		idNameMap[opt.Option] = opt.Label
	}

	return idNameMap, nil
}

// GetSelectedDimensionOptionsJSON will return a list of selected options from the filter api with corresponding label
func (f *Filter) GetSelectedDimensionOptionsJSON(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	ctx := req.Context()

	req = forwardFlorenceTokenIfRequired(req)

	opts, err := f.FilterClient.GetDimensionOptions(req.Context(), filterID, name)
	if err != nil {
		log.InfoCtx(ctx, "failed to get dimension options", log.Data{"error": err, "filter_id": filterID, "dimension": name})
		setStatusCode(req, w, err)
		return
	}

	fj, err := f.FilterClient.GetJobState(req.Context(), filterID)
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
	idNameMap, err := f.getIDNameMap(req.Context(), versionURL.Path, name)
	if err != nil {
		log.InfoCtx(ctx, "failed to get name map", log.Data{"error": err, "filter_id": filterID, "path": versionURL, "name": name})
		setStatusCode(req, w, err)
		return
	}

	var lids []labelID

	if name == "time" {

		var codedDates []string
		labelIDMap := make(map[string]string)
		for _, opt := range opts {
			codedDates = append(codedDates, idNameMap[opt.Option])
			labelIDMap[idNameMap[opt.Option]] = opt.Option
		}

		readbleDates, err := dates.ConvertToReadable(codedDates)
		if err != nil {
			log.InfoCtx(ctx, "failed to convert dates", log.Data{"error": err, "filter_id": filterID, "dates": codedDates})
			setStatusCode(req, w, err)
			return
		}

		readbleDates = dates.Sort(readbleDates)

		for _, date := range readbleDates {
			lid := labelID{
				Label: fmt.Sprintf("%s %d", date.Month(), date.Year()),
				ID:    labelIDMap[date.Format("Jan-06")],
			}

			lids = append(lids, lid)
		}
	}

	b, err := json.Marshal(lids)
	if err != nil {
		log.InfoCtx(ctx, "failed to marshal json", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	w.Write(b)
}

// DimensionSelector controls the render of the range selector template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) DimensionSelector(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	ctx := req.Context()

	req = forwardFlorenceTokenIfRequired(req)

	fj, err := f.FilterClient.GetJobState(req.Context(), filterID)
	if err != nil {
		log.InfoCtx(ctx, "failed to get job state", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	selectedValues, err := f.FilterClient.GetDimensionOptions(req.Context(), filterID, name)
	if err != nil {
		log.InfoCtx(ctx, "failed to get options from filter client", log.Data{"error": err, "filter_id": filterID, "dimension": name})
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

	allValues, err := f.DatasetClient.GetOptions(req.Context(), datasetID, edition, version, name)
	if err != nil {
		log.InfoCtx(ctx, "failed to get options from dataset client",
			log.Data{"error": err, "dimension": name, "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	// TODO: This is a shortcut for now, if the hierarchy api returns a status 200
	// then the dimension should be populated as a hierarchy
	if _, err = f.HierarchyClient.GetRoot(fj.InstanceID, name); err == nil && len(allValues.Items) > 20 {
		f.Hierarchy(w, req)
		return
	}

	dims, err := f.DatasetClient.GetDimensions(req.Context(), datasetID, edition, version)
	if err != nil {
		log.InfoCtx(ctx, "failed to get dimensions", log.Data{"error": err, "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	f.listSelector(w, req, name, selectedValues, allValues, fj, dataset, dims, datasetID, ver.ReleaseDate)
}

// ListSelector controls the render of the age selector list template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) listSelector(w http.ResponseWriter, req *http.Request, name string, selectedValues []filter.DimensionOption, allValues dataset.Options, filter filter.Model, dataset dataset.Model, dims dataset.Dimensions, datasetID, releaseDate string) {
	p := mapper.CreateListSelectorPage(req.Context(), name, selectedValues, allValues, filter, dataset, dims, datasetID, releaseDate)

	b, err := json.Marshal(p)
	if err != nil {
		log.InfoCtx(req.Context(), "failed to marshal json", log.Data{"error": err, "filter_id": filter.FilterID})
		setStatusCode(req, w, err)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/list-selector", b)
	if err != nil {
		log.InfoCtx(req.Context(), "failed to render", log.Data{"error": err, "filter_id": filter.FilterID})
		setStatusCode(req, w, err)
		return
	}

	w.Write(templateBytes)
}

// DimensionAddAll will add all dimension values to a basket
func (f *Filter) DimensionAddAll(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	f.addAll(w, req, fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name))
}

func (f *Filter) addAll(w http.ResponseWriter, req *http.Request, redirectURL string) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	ctx := req.Context()

	req = forwardFlorenceTokenIfRequired(req)

	fj, err := f.FilterClient.GetJobState(req.Context(), filterID)
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

	vals, err := f.DatasetClient.GetOptions(req.Context(), datasetID, edition, version, name)
	if err != nil {
		log.InfoCtx(ctx, "failed to get options from dataset client",
			log.Data{"error": err, "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	var options []string
	for _, val := range vals.Items {
		options = append(options, val.Option)
	}

	if err := f.FilterClient.AddDimensionValues(req.Context(), filterID, name, options); err != nil {
		log.InfoCtx(ctx, "failed to add dimension values", log.Data{"error": err, "filter_id": filterID, "dimension": name})
		setStatusCode(req, w, err)
		return
	}

	http.Redirect(w, req, redirectURL, 302)
}

// AddList adds a list of values
func (f *Filter) AddList(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	ctx := req.Context()

	req = forwardFlorenceTokenIfRequired(req)

	if err := req.ParseForm(); err != nil {
		log.InfoCtx(ctx, "failed to parse form", log.Data{"error": err, "filter_id": filterID, "dimension": name})
		setStatusCode(req, w, err)
		return
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)

	if len(req.Form["add-all"]) > 0 {
		redirectURL = fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
		f.addAll(w, req, redirectURL)
		return
	}

	if len(req.Form["remove-all"]) > 0 {
		redirectURL = fmt.Sprintf("/filters/%s/dimensions/%s/remove-all", filterID, name)
		http.Redirect(w, req, redirectURL, 302)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// concurrently remove any fields that have been deselected
	go func() {
		opts, err := f.FilterClient.GetDimensionOptions(ctx, filterID, name)
		if err != nil {
			log.InfoCtx(ctx, "failed to get options from filter client", log.Data{"error": err, "filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}

		for _, opt := range opts {
			if _, ok := req.Form[opt.Option]; !ok {
				if err := f.FilterClient.RemoveDimensionValue(ctx, filterID, name, opt.Option); err != nil {
					log.ErrorCtx(ctx, err, nil)
				}
			}
		}

		wg.Done()
	}()

	wg.Wait()

	var options []string
	for k := range req.Form {
		if k == ":uri" || k == "save-and-return" {
			continue
		}

		options = append(options, k)
	}

	if err := f.FilterClient.AddDimensionValues(ctx, filterID, name, options); err != nil {
		log.InfoCtx(ctx, err.Error(), nil)
	}

	http.Redirect(w, req, redirectURL, 302)
}

func (f *Filter) getDimensionValues(ctx context.Context, filterID, name string) (values []string, labelIDMap map[string]string, err error) {
	fj, err := f.FilterClient.GetJobState(ctx, filterID)
	if err != nil {
		return
	}

	versionURL, err := url.Parse(fj.Links.Version.HRef)
	if err != nil {
		return
	}

	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		return
	}

	vals, err := f.DatasetClient.GetOptions(ctx, datasetID, edition, version, name)
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
func (f *Filter) DimensionRemoveAll(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	ctx := req.Context()

	log.InfoCtx(ctx, "attempting to remove all options from dimension", log.Data{"dimension": name, "filterID": filterID})

	req = forwardFlorenceTokenIfRequired(req)

	if err := f.FilterClient.RemoveDimension(req.Context(), filterID, name); err != nil {
		log.InfoCtx(ctx, "failed to remove dimension", log.Data{"error": err, "filter_id": filterID, "dimension": name})
		setStatusCode(req, w, err)
		return
	}

	if err := f.FilterClient.AddDimension(req.Context(), filterID, name); err != nil {
		log.InfoCtx(ctx, "failed to add dimension", log.Data{"error": err, "filter_id": filterID, "dimension": name})
		setStatusCode(req, w, err)
		return
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
	http.Redirect(w, req, redirectURL, 302)
}

// DimensionRemoveOne removes an individual option on a dimensions
func (f *Filter) DimensionRemoveOne(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	option := vars["option"]

	req = forwardFlorenceTokenIfRequired(req)

	if err := f.FilterClient.RemoveDimensionValue(req.Context(), filterID, name, option); err != nil {
		log.InfoCtx(req.Context(), "failed to remove dimension option", log.Data{"error": err, "filter_id": filterID, "dimension": name, "option": option})
		setStatusCode(req, w, err)
		return
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
	http.Redirect(w, req, redirectURL, 302)
}
