package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	dphandlers "github.com/ONSdigital/dp-net/v3/handlers"

	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/hierarchy"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// codes for geography nodes that need to be flattened in a single layer
const (
	Uk              = "K02000001"
	GreatBritain    = "K03000001"
	EnglandAndWales = "K04000001"
	England         = "E92000001"
	NorthernIreland = "N92000002"
	Scotland        = "S92000003"
	Wales           = "W92000004"
)

// HierarchyUpdate controls the updating of a hierarchy job
func (f *Filter) HierarchyUpdate() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		name := vars["name"]
		code := vars["code"]

		ctx := req.Context()

		if err := req.ParseForm(); err != nil {
			log.Error(ctx, "failed to parse request", err)
			return
		}

		var redirectURI string
		if len(req.Form["save-and-return"]) > 0 {
			redirectURI = fmt.Sprintf("/filters/%s/dimensions", filterID)
		} else {
			if code != "" {
				redirectURI = fmt.Sprintf("/filters/%s/dimensions/%s/%s", filterID, name, code)
			} else {
				redirectURI = fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
			}
		}

		fil, eTag, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
		if err != nil {
			log.Error(ctx, "failed to get job state", err, log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		if len(req.Form["add-all"]) > 0 {
			f.addAllHierarchyLevel(w, req, fil, name, code, redirectURI, userAccessToken, collectionID, eTag)
			return
		}

		if len(req.Form["remove-all"]) > 0 {
			f.removeAllHierarchyLevel(w, req, fil, name, code, redirectURI, userAccessToken, collectionID, eTag)
			return
		}

		h, err := f.buildHierarchyModel(ctx, fil, name, code)
		if err != nil {
			log.Error(ctx, "failed to get hierarchy node", err, log.Data{"filter_id": filterID, "dimension": name, "code": code})
			setStatusCode(req, w, err)
			return
		}

		// obtain options to remove from unselected values (not provided in form)
		removeOptions := []string{}
		for _, hv := range h.Children {
			if _, ok := req.Form[hv.Links.Code.ID]; !ok {
				removeOptions = append(removeOptions, hv.Links.Code.ID)
			}
		}

		// get options to add and overwrite redirectURI, if provided in the form
		addOptions := getOptionsAndRedirect(req.Form, &redirectURI)

		_, err = f.FilterClient.PatchDimensionValues(ctx, userAccessToken, "", collectionID, filterID, name, addOptions, removeOptions, f.BatchSize, eTag)
		if err != nil {
			log.Error(ctx, "failed to patch dimension values", err, log.Data{"filter_id": filterID, "dimension": name, "code": code})
			setStatusCode(req, w, err)
			return
		}
		http.Redirect(w, req, redirectURI, http.StatusFound)
	})
}

func (f *Filter) buildHierarchyModel(ctx context.Context, fil filter.Model, name, code string) (h hierarchy.Model, err error) {
	if code != "" {
		return f.HierarchyClient.GetChild(ctx, fil.InstanceID, name, code)
	}
	if name == geography {
		h, err = f.flattenGeographyTopLevel(ctx, fil.InstanceID)
	} else {
		h, err = f.HierarchyClient.GetRoot(ctx, fil.InstanceID, name)
	}

	if err != nil {
		return h, err
	}

	// We include the value on the root as a selectable item, so append
	// the value on the root to the child to see if it has been removed by
	// the user
	h.Children = append(h.Children, hierarchy.Child{
		Links: h.Links,
	})
	return h, err
}

func (f *Filter) addAllHierarchyLevel(w http.ResponseWriter, req *http.Request, fil filter.Model, name, code, redirectURI, userAccessToken, collectionID, eTag string) {
	ctx := req.Context()
	var err error

	var h hierarchy.Model
	if code != "" {
		h, err = f.HierarchyClient.GetChild(ctx, fil.InstanceID, name, code)
	} else {
		if name == geography {
			h, err = f.flattenGeographyTopLevel(ctx, fil.InstanceID)
		} else {
			h, err = f.HierarchyClient.GetRoot(ctx, fil.InstanceID, name)
		}
	}
	if err != nil {
		log.Error(ctx, "failed to get hierarchy node", err, log.Data{"filter_id": fil.FilterID, "dimension": name, "code": code})
		setStatusCode(req, w, err)
		return
	}

	options := make([]string, 0, len(h.Children))
	for _, child := range h.Children {
		options = append(options, child.Links.Code.ID)
	}
	_, err = f.FilterClient.SetDimensionValues(req.Context(), userAccessToken, "", collectionID, fil.FilterID, name, options, eTag)
	if err != nil {
		log.Error(ctx, "failed to add dimension values", err)
	}

	http.Redirect(w, req, redirectURI, http.StatusFound)
}

func (f *Filter) removeAllHierarchyLevel(w http.ResponseWriter, req *http.Request, fil filter.Model, name, code, redirectURI, userAccessToken, collectionID, eTag string) {
	ctx := req.Context()
	var h hierarchy.Model
	var err error

	if code != "" {
		h, err = f.HierarchyClient.GetChild(ctx, fil.InstanceID, name, code)
	} else {
		if name == geography {
			h, err = f.flattenGeographyTopLevel(ctx, fil.InstanceID)
		} else {
			h, err = f.HierarchyClient.GetRoot(ctx, fil.InstanceID, name)
		}
	}
	if err != nil {
		log.Error(ctx, "failed to get hierarchy node", err, log.Data{"filter_id": fil.FilterID, "dimension": name, "code": code})
		setStatusCode(req, w, err)
		return
	}

	// list of dimensions to remove created from children code links
	removeOptions := []string{}
	for _, child := range h.Children {
		removeOptions = append(removeOptions, child.Links.Code.ID)
	}

	// remove all items
	_, err = f.FilterClient.PatchDimensionValues(ctx, userAccessToken, "", collectionID, fil.FilterID, name, []string{}, removeOptions, f.BatchSize, eTag)
	if err != nil {
		log.Error(ctx, "failed to remove dimension values using a patch", err, log.Data{"filter_id": fil.FilterID, "dimension": name, "code": code, "options": removeOptions})
	}

	http.Redirect(w, req, redirectURI, http.StatusFound)
}

// Hierarchy controls the creation of a hierarchy page
func (f *Filter) Hierarchy() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		name := vars["name"]
		code := vars["code"]
		ctx := req.Context()

		fil, eTag0, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
		if err != nil {
			log.Error(ctx, "failed to get job state", err, log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		var h hierarchy.Model
		if code != "" {
			h, err = f.HierarchyClient.GetChild(ctx, fil.InstanceID, name, code)
		} else {
			if name == geography {
				h, err = f.flattenGeographyTopLevel(ctx, fil.InstanceID)
			} else {
				h, err = f.HierarchyClient.GetRoot(ctx, fil.InstanceID, name)
			}
		}

		if err != nil {
			log.Error(ctx, "failed to get hierarchy node", err, log.Data{"filter_id": filterID, "dimension": name, "code": code})
			setStatusCode(req, w, err)
			return
		}

		selVals, eTag1, err := f.FilterClient.GetDimensionOptionsInBatches(req.Context(), userAccessToken, "", collectionID, filterID, name, f.BatchSize, f.BatchMaxWorkers)
		if err != nil {
			log.Error(ctx, "failed to get options from filter client", err, log.Data{"filter_id": filterID, "dimension": name})
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

		versionURL, err := url.Parse(fil.Links.Version.HRef)
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

		d, err := f.DatasetClient.Get(req.Context(), userAccessToken, "", collectionID, datasetID)
		if err != nil {
			log.Error(req.Context(), "failed to get dataset", err, log.Data{"dataset_id": datasetID})
			setStatusCode(req, w, err)
			return
		}

		dims, err := f.DatasetClient.GetVersionDimensions(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Error(ctx, "failed to get dimensions", err,
				log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		selValsLabelMap, err := f.getIDNameLookupFromDatasetAPI(ctx, userAccessToken, collectionID, datasetID, edition, version, name, selVals)
		if err != nil {
			log.Error(ctx, "failed to get options from dataset client for the selected values", err,
				log.Data{"dimension": name, "dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		homepageContent, err := f.ZebedeeClient.GetHomepageContent(ctx, userAccessToken, collectionID, lang, "/")
		if err != nil {
			log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
		}

		bp := f.RenderClient.NewBasePageModel()
		p := mapper.CreateHierarchyPage(req, bp, h, d, fil, selValsLabelMap, dims, name, req.URL.Path, datasetID, f.APIRouterVersion, lang, homepageContent.ServiceMessage, homepageContent.EmergencyBanner)
		f.RenderClient.BuildPage(w, p, "hierarchy")
	})
}

type flatNodes struct {
	list         []hierarchy.Child
	defaultOrder map[string]int
}

func (n *flatNodes) addWithoutChildren(val hierarchy.Child) {
	if !val.HasData {
		return
	}

	n.list = append(n.list, hierarchy.Child{
		Label:   val.Label,
		Links:   val.Links,
		HasData: val.HasData,
		Order:   val.Order,
	})
}

func (n *flatNodes) addWithChildren(val hierarchy.Child) {
	n.list = append(n.list, hierarchy.Child{
		Label:            val.Label,
		Links:            val.Links,
		HasData:          val.HasData,
		Order:            val.Order,
		NumberofChildren: val.NumberofChildren,
	})
}

// hasOrder returns true if and only if all child items in the list have a non-nil order values
func (n *flatNodes) hasOrder() bool {
	if n.list == nil {
		return false
	}

	for _, child := range n.list {
		if child.Order == nil {
			return false
		}
	}
	return true
}

// getOrder obtains the order value, with paramater checking, and assuming that it's not nil
// returns the order value, or -1 if any parameter check failed or the order was nil
func (n *flatNodes) getOrder(i int) int {
	// split checks over separate if's so that code coverage can be related to unit test code
	if n.list == nil {
		return -1
	}
	if i >= len(n.list) {
		return -1
	}
	if n.list[i].Order == nil {
		return -1
	}

	return *n.list[i].Order
}

// getDefaultOrder obtains the default order value according to the defaultOrder slice, with parameter checking
// returns the default order value corresponding to the child item in the provided index, or -1 if not defined
func (n *flatNodes) getDefaultOrder(i int) int {
	// split checks over separate if's so that code coverage can be related to unit test code
	if n.list == nil {
		return -1
	}
	if i >= len(n.list) {
		return -1
	}
	if n.list[i].Links.Code.ID == "" {
		return -1
	}

	order, ok := n.defaultOrder[n.list[i].Links.Code.ID]
	if !ok {
		return -1
	}
	return order
}

// sort child items by order property, or by default values as fallback (if order is not defined in all items)
func (n *flatNodes) sort() {
	// split checks over separate if's so that code coverage can be related to unit test code
	if n.list == nil {
		return
	}
	if len(n.list) == 0 {
		return
	}

	if n.hasOrder() {
		sort.Slice(n.list, func(i, j int) bool {
			return n.getOrder(i) < n.getOrder(j)
		})
	} else {
		sort.Slice(n.list, func(i, j int) bool {
			return n.getDefaultOrder(i) < n.getDefaultOrder(j)
		})
	}
}

// Flatten the geography hierarchy - please note this will only work for this particular hierarchy,
// need helper functions for other geog hierarchies too.
func (f *Filter) flattenGeographyTopLevel(ctx context.Context, instanceID string) (h hierarchy.Model, err error) {
	// obtain root element
	root, err := f.HierarchyClient.GetRoot(ctx, instanceID, geography)
	if err != nil {
		return h, err
	}

	// if root has data, we need to copy label and links to the model
	if root.HasData {
		h.Label = root.Label
		h.Links = root.Links
		h.HasData = root.HasData
	}

	// create nodes struct with default order
	nodes := flatNodes{
		list:         []hierarchy.Child{},
		defaultOrder: map[string]int{GreatBritain: 0, EnglandAndWales: 1, England: 2, NorthernIreland: 3, Scotland: 4, Wales: 5},
	}

	// add items to flatNodes
	for _, val := range root.Children {
		if val.Links.Code.ID == GreatBritain {
			nodes.addWithoutChildren(val)

			child, childErr := f.HierarchyClient.GetChild(ctx, instanceID, geography, GreatBritain)
			if childErr != nil {
				return h, childErr
			}

			for _, childVal := range child.Children {
				if childVal.Links.Code.ID == EnglandAndWales {
					nodes.addWithoutChildren(childVal)

					grandChild, cErr := f.HierarchyClient.GetChild(ctx, instanceID, geography, childVal.Links.Code.ID)
					if cErr != nil {
						return h, cErr
					}

					for _, grandChildVal := range grandChild.Children {
						if grandChildVal.Links.Code.ID == England {
							nodes.addWithChildren(grandChildVal)
						}

						if grandChildVal.Links.Code.ID == Wales {
							nodes.addWithChildren(grandChildVal)
						}
					}
				}

				if childVal.Links.Code.ID == Scotland {
					nodes.addWithChildren(childVal)
				}
			}
		}

		if val.Links.Code.ID == NorthernIreland {
			nodes.addWithChildren(val)
		}
	}

	// sort nodes according to their defined order, or the defaultOrder as a fallback
	nodes.sort()

	// use nodes.list only if the list is not empty. Otherwise, use the Children items under the root node
	if len(nodes.list) == 0 {
		h.Children = root.Children
	} else {
		h.Children = nodes.list
	}

	return h, err
}
