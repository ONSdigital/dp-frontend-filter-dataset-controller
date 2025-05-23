package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/hierarchy"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dates"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	dphandlers "github.com/ONSdigital/dp-net/v3/handlers"
	"github.com/ONSdigital/log.go/v2/log"
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

		fj, _, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
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

		idNameMap, err := f.getIDNameMap(req.Context(), userAccessToken, collectionID, versionPath, name)
		if err != nil {
			log.Error(ctx, "failed to get name map", err, log.Data{"filter_id": filterID, "path": versionPath, "name": name})
			setStatusCode(req, w, err)
			return
		}

		var lids []labelID

		if name == strTime {
			var codedDates []string
			labelIDMap := make(map[string]string)
			for k, v := range idNameMap {
				codedDates = append(codedDates, v)
				labelIDMap[v] = k
			}

			readableDates, dErr := dates.ConvertToReadable(codedDates)
			if dErr != nil {
				log.Error(ctx, "failed to convert dates", dErr, log.Data{"filter_id": filterID, "dates": codedDates})
				setStatusCode(req, w, dErr)
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
			log.Error(ctx, "failed to marshal json", err, log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}

		//nolint:errcheck // ignore error
		w.Write(b)
	})
}

func (f *Filter) getIDNameMap(ctx context.Context, userAccessToken, collectionID, versionURL, dimension string) (idNameMap map[string]string, err error) {
	datasetID, edition, version, _ := helpers.ExtractDatasetInfoFromPath(ctx, versionURL)
	opts, err := f.DatasetClient.GetOptionsInBatches(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dimension, f.BatchSize, f.BatchMaxWorkers)
	if err != nil {
		return nil, err
	}

	idNameMap = make(map[string]string)
	for i := range opts.Items {
		idNameMap[opts.Items[i].Option] = opts.Items[i].Label
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

		opts, eTag0, err := f.FilterClient.GetDimensionOptionsInBatches(req.Context(), userAccessToken, "", collectionID, filterID, name, f.BatchSize, f.BatchMaxWorkers)
		if err != nil {
			log.Error(ctx, "failed to get dimension options", err, log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			// The user might want to retry this handler on ErrBatchETagMismatch
			return
		}

		fj, eTag1, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
		if err != nil {
			log.Error(ctx, "failed to get job state", err, log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		// The user might want to retry this handler if eTags don't match
		if eTag0 != eTag1 {
			conflictErr := errors.New("inconsistent filter data")
			log.Error(ctx, "data consistency cannot be guaranteed because filter was modified between calls", conflictErr,
				log.Data{"filter_id": filterID, "dimension": name, "e_tag_0": eTag0, "e_tag_1": eTag1})
			setStatusCode(req, w, conflictErr)
			return
		}

		versionURL, err := url.Parse(fj.Links.Version.HRef)
		if err != nil {
			log.Error(ctx, "failed to parse version href", err, log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}
		versionPath := strings.TrimPrefix(versionURL.Path, f.APIRouterVersion)
		idNameMap, err := f.getIDNameMap(req.Context(), userAccessToken, collectionID, versionPath, name)
		if err != nil {
			log.Error(ctx, "failed to get name map", err, log.Data{"filter_id": filterID, "path": versionPath, "name": name})
			setStatusCode(req, w, err)
			return
		}

		var lids []labelID

		if name == strTime {
			var codedDates []string
			labelIDMap := make(map[string]string)
			for _, opt := range opts.Items {
				codedDates = append(codedDates, idNameMap[opt.Option])
				labelIDMap[idNameMap[opt.Option]] = opt.Option
			}

			readableDates, dErr := dates.ConvertToReadable(codedDates)
			if dErr != nil {
				log.Error(ctx, "failed to convert dates", dErr, log.Data{"filter_id": filterID, "dates": codedDates})
				setStatusCode(req, w, dErr)
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
			log.Error(ctx, "failed to marshal json", err, log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		//nolint:errcheck // ignore error
		w.Write(b)
	})
}

// DimensionSelector controls the render of the range selector template using data from Dataset API and Filter API
func (f *Filter) DimensionSelector() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		name := vars["name"]
		filterID := vars["filterID"]
		ctx := req.Context()

		fj, eTag0, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
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

		datasetDetails, err := f.DatasetClient.Get(req.Context(), userAccessToken, "", collectionID, datasetID)
		if err != nil {
			log.Error(ctx, "failed to get dataset", err, log.Data{"dataset_id": datasetID})
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
		opts, err := f.DatasetClient.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, name, &dataset.QueryParams{Offset: 0, Limit: 0})
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
			log.Error(ctx, "failed to get dimensions", err, log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		selectedValues, eTag1, err := f.FilterClient.GetDimensionOptionsInBatches(req.Context(), userAccessToken, "", collectionID, filterID, name, f.BatchSize, f.BatchMaxWorkers)
		if err != nil {
			log.Error(ctx, "failed to get options from filter client", err, log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			// The user might want to retry this handler on ErrBatchETagMismatch
			return
		}

		// The user might want to retry this handler if eTags don't match
		if eTag0 != eTag1 {
			conflictErr := errors.New("inconsistent filter data")
			log.Error(ctx, "data consistency cannot be guaranteed because filter was modified between calls", conflictErr,
				log.Data{"filter_id": filterID, "dimension": name, "e_tag_0": eTag0, "e_tag_1": eTag1})
			setStatusCode(req, w, conflictErr)
			return
		}

		allValues, err := f.DatasetClient.GetOptionsInBatches(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version, name, f.BatchSize, f.BatchMaxWorkers)
		if err != nil {
			log.Error(ctx, "failed to get options from dataset client", err, log.Data{"dimension": name, "dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		homepageContent, err := f.ZebedeeClient.GetHomepageContent(ctx, userAccessToken, collectionID, lang, "/")
		if err != nil {
			log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
		}

		f.listSelector(w, req, name, selectedValues.Items, allValues, fj, datasetDetails, dims, datasetID, lang, homepageContent.ServiceMessage, homepageContent.EmergencyBanner)
	})
}

func (f *Filter) isHierarchicalDimension(ctx context.Context, instanceID, dimensionName string) (bool, error) {
	_, err := f.HierarchyClient.GetRoot(ctx, instanceID, dimensionName)
	if err != nil {
		var getHierarchyErr *hierarchy.ErrInvalidHierarchyAPIResponse
		if errors.As(err, &getHierarchyErr) && http.StatusNotFound == getHierarchyErr.Code() {
			return false, nil
		}

		log.Error(ctx, "unexpected error getting hierarchy root for dimension", err, log.Data{
			"instance_id":    instanceID,
			"dimension_name": dimensionName,
		})

		return false, err
	}

	return true, nil
}

// ListSelector controls the render of the age selector list template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) listSelector(w http.ResponseWriter, req *http.Request, name string, selectedValues []filter.DimensionOption, allValues dataset.Options, fm filter.Model, ds dataset.DatasetDetails, dims dataset.VersionDimensions, datasetID, lang, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) {
	bp := f.RenderClient.NewBasePageModel()
	p := mapper.CreateListSelectorPage(req, bp, name, selectedValues, allValues, fm, ds, dims, datasetID, f.APIRouterVersion, lang, serviceMessage, emergencyBannerContent)
	f.RenderClient.BuildPage(w, p, "list-selector")
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
	vars := mux.Vars(req)
	name := vars["name"]
	filterID := vars["filterID"]
	ctx := req.Context()

	fj, eTag, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
	if err != nil {
		log.Error(ctx, "failed to get job state", err, log.Data{"filter_id": filterID})
		setStatusCode(req, w, err)
		// The user might want to retry this handler on ErrBatchETagMismatch
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

	// function to add each batch of dataset dimension options to filter API
	processBatch := func(batch dataset.Options) (forceAbort bool, err error) {
		var options []string
		for i := range batch.Items {
			options = append(options, batch.Items[i].Option)
		}
		// first batch, will overwrite any existing values in filter API
		if batch.Offset == 0 {
			eTag, err = f.FilterClient.SetDimensionValues(req.Context(), userAccessToken, "", collectionID, filterID, name, options, eTag)
			return false, err
		}
		// the rest of batches will be added to the existing items in filter API via patch operations
		eTag, err = f.FilterClient.PatchDimensionValues(req.Context(), userAccessToken, "", collectionID, filterID, name, options, []string{}, f.BatchSize, eTag)
		return false, err
	}

	// call dataset API GetOptions in batches, and process each batch to add the options to filter API
	if err := f.DatasetClient.GetOptionsBatchProcess(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version, name, nil, processBatch, f.BatchSize, f.BatchMaxWorkers); err != nil {
		log.Error(ctx, "failed to process options from dataset api", err, log.Data{"filter_id": filterID, "dimension": name})
		setStatusCode(req, w, err)
		return
	}

	http.Redirect(w, req, redirectURL, http.StatusFound)
}

// AddList sets a list of values, removing any existing value.
func (f *Filter) AddList() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		name := vars["name"]
		filterID := vars["filterID"]
		ctx := req.Context()

		if err := req.ParseForm(); err != nil {
			log.Error(ctx, "failed to parse form", err, log.Data{"filter_id": filterID, "dimension": name})
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
			http.Redirect(w, req, redirectURL, http.StatusFound)
			return
		}

		var options []string
		for k := range req.Form {
			if k == ":uri" || k == "save-and-return" {
				continue
			}

			options = append(options, k)
		}

		_, err := f.FilterClient.SetDimensionValues(ctx, userAccessToken, "", collectionID, filterID, name, options, headers.IfMatchAnyETag)
		if err != nil {
			log.Warn(ctx, "failed to add dimension values", log.FormatErrors([]error{err}))
		}

		http.Redirect(w, req, redirectURL, http.StatusFound)
	})
}

func (f *Filter) getDimensionValues(ctx context.Context, userAccessToken, collectionID, filterID, name string) (values []string, labelIDMap map[string]string, err error) {
	fj, _, err := f.FilterClient.GetJobState(ctx, userAccessToken, "", "", collectionID, filterID)
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
	for i := range vals.Items {
		values = append(values, vals.Items[i].Label)
		labelIDMap[vals.Items[i].Label] = vals.Items[i].Option
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

		log.Info(ctx, "attempting to remove all options from dimension", log.Data{"dimension": name, "filterID": filterID})

		eTag, err := f.FilterClient.RemoveDimension(req.Context(), userAccessToken, "", collectionID, filterID, name, headers.IfMatchAnyETag)
		if err != nil {
			log.Error(ctx, "failed to remove dimension", err, log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}

		_, err = f.FilterClient.AddDimension(req.Context(), userAccessToken, "", collectionID, filterID, name, eTag)
		if err != nil {
			log.Error(ctx, "failed to add dimension", err, log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}

		redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
		http.Redirect(w, req, redirectURL, http.StatusFound)
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

		_, err := f.FilterClient.RemoveDimensionValue(req.Context(), userAccessToken, "", collectionID, filterID, name, option, headers.IfMatchAnyETag)
		if err != nil {
			log.Error(ctx, "failed to remove dimension option", err, log.Data{"filter_id": filterID, "dimension": name, "option": option})
			setStatusCode(req, w, err)
			return
		}

		redirectURL := fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
		http.Redirect(w, req, redirectURL, http.StatusFound)
	})
}
