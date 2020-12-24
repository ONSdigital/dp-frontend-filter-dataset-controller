package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	dphandlers "github.com/ONSdigital/dp-net/handlers"

	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/hierarchy"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
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
			log.Event(ctx, "failed to parse request", log.ERROR, log.Error(err))
			return
		}

		var redirectURI string
		if len(req.Form["save-and-return"]) > 0 {
			redirectURI = fmt.Sprintf("/filters/%s/dimensions", filterID)
		} else {
			if len(code) > 0 {
				redirectURI = fmt.Sprintf("/filters/%s/dimensions/%s/%s", filterID, name, code)
			} else {
				redirectURI = fmt.Sprintf("/filters/%s/dimensions/%s", filterID, name)
			}
		}

		fil, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
		if err != nil {
			log.Event(ctx, "failed to get job state", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		if len(req.Form["add-all"]) > 0 {
			f.addAllHierarchyLevel(w, req, fil, name, code, redirectURI, userAccessToken, collectionID)
			return
		}

		if len(req.Form["remove-all"]) > 0 {
			f.removeAllHierarchyLevel(w, req, fil, name, code, redirectURI, userAccessToken, collectionID)
			return
		}

		h, err := f.buildHierarchyModel(ctx, fil, name, code)
		if err != nil {
			log.Event(ctx, "failed to get hierarchy node", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name, "code": code})
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
		var addOptions []string
		addOptions = getOptionsAndRedirect(req.Form, &redirectURI)

		err = f.FilterClient.PatchDimensionValues(ctx, userAccessToken, "", collectionID, filterID, name, addOptions, removeOptions, f.BatchSize)
		if err != nil {
			log.Event(ctx, "failed to patch dimension values", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name, "code": code})
			setStatusCode(req, w, err)
			return
		}
		http.Redirect(w, req, redirectURI, 302)
	})
}

func (f *Filter) buildHierarchyModel(ctx context.Context, fil filter.Model, name, code string) (h hierarchy.Model, err error) {
	if len(code) > 0 {
		return f.HierarchyClient.GetChild(ctx, fil.InstanceID, name, code)
	}
	if name == "geography" {
		h, err = f.flattenGeographyTopLevel(ctx, fil.InstanceID)
	} else {
		h, err = f.HierarchyClient.GetRoot(ctx, fil.InstanceID, name)
	}

	// We include the value on the root as a selectable item, so append
	// the value on the root to the child to see if it has been removed by
	// the user
	h.Children = append(h.Children, hierarchy.Child{
		Links: h.Links,
	})
	return h, err
}

func (f *Filter) addAllHierarchyLevel(w http.ResponseWriter, req *http.Request, fil filter.Model, name, code, redirectURI, userAccessToken, collectionID string) {

	ctx := req.Context()
	var err error

	var h hierarchy.Model
	if len(code) > 0 {
		h, err = f.HierarchyClient.GetChild(ctx, fil.InstanceID, name, code)
	} else {
		if name == "geography" {
			h, err = f.flattenGeographyTopLevel(ctx, fil.InstanceID)
		} else {
			h, err = f.HierarchyClient.GetRoot(ctx, fil.InstanceID, name)
		}
	}
	if err != nil {
		log.Event(ctx, "failed to get hierarchy node", log.ERROR, log.Error(err), log.Data{"filter_id": fil.FilterID, "dimension": name, "code": code})
		setStatusCode(req, w, err)
		return
	}

	var options []string
	for _, child := range h.Children {
		options = append(options, child.Links.Code.ID)
	}
	if err := f.FilterClient.SetDimensionValues(req.Context(), userAccessToken, "", collectionID, fil.FilterID, name, options); err != nil {
		log.Event(ctx, "failed to add dimension values", log.ERROR, log.Error(err))
	}

	http.Redirect(w, req, redirectURI, 302)
}

func (f *Filter) removeAllHierarchyLevel(w http.ResponseWriter, req *http.Request, fil filter.Model, name, code, redirectURI, userAccessToken, collectionID string) {
	ctx := req.Context()
	var h hierarchy.Model
	var err error

	if len(code) > 0 {
		h, err = f.HierarchyClient.GetChild(ctx, fil.InstanceID, name, code)
	} else {
		if name == "geography" {
			h, err = f.flattenGeographyTopLevel(ctx, fil.InstanceID)
		} else {
			h, err = f.HierarchyClient.GetRoot(ctx, fil.InstanceID, name)
		}
	}
	if err != nil {
		log.Event(ctx, "failed to get hierarchy node", log.ERROR, log.Error(err), log.Data{"filter_id": fil.FilterID, "dimension": name, "code": code})
		setStatusCode(req, w, err)
		return
	}

	// list of dimensions to remove created from children code links
	removeOptions := []string{}
	for _, child := range h.Children {
		removeOptions = append(removeOptions, child.Links.Code.ID)
	}

	// remove all items
	if err := f.FilterClient.PatchDimensionValues(ctx, userAccessToken, "", collectionID, fil.FilterID, name, []string{}, removeOptions, f.BatchSize); err != nil {
		log.Event(ctx, "failed to remove dimension values using a patch", log.ERROR, log.Error(err), log.Data{"filter_id": fil.FilterID, "dimension": name, "code": code, "options": removeOptions})
	}

	http.Redirect(w, req, redirectURI, 302)
}

// Hierarchy controls the creation of a hierarchy page
func (f *Filter) Hierarchy() http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		filterID := vars["filterID"]
		name := vars["name"]
		code := vars["code"]
		ctx := req.Context()

		fil, err := f.FilterClient.GetJobState(req.Context(), userAccessToken, "", "", collectionID, filterID)
		if err != nil {
			log.Event(ctx, "failed to get job state", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		var h hierarchy.Model
		if len(code) > 0 {
			h, err = f.HierarchyClient.GetChild(ctx, fil.InstanceID, name, code)
		} else {
			if name == "geography" {
				h, err = f.flattenGeographyTopLevel(ctx, fil.InstanceID)
			} else {
				h, err = f.HierarchyClient.GetRoot(ctx, fil.InstanceID, name)
			}
		}

		if err != nil {
			log.Event(ctx, "failed to get hierarchy node", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name, "code": code})
			setStatusCode(req, w, err)
			return
		}

		selVals, err := f.GetDimensionOptionsFromFilterAPI(req.Context(), userAccessToken, collectionID, filterID, name)
		if err != nil {
			log.Event(ctx, "failed to get options from filter client", log.ERROR, log.Error(err), log.Data{"filter_id": filterID, "dimension": name})
			setStatusCode(req, w, err)
			return
		}

		versionURL, err := url.Parse(fil.Links.Version.HRef)
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

		d, err := f.DatasetClient.Get(req.Context(), userAccessToken, "", collectionID, datasetID)
		if err != nil {
			log.Event(req.Context(), "failed to get dataset", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID})
			setStatusCode(req, w, err)
			return
		}
		ver, err := f.DatasetClient.GetVersion(req.Context(), userAccessToken, "", "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Event(req.Context(), "failed to get version", log.ERROR, log.Error(err), log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		dims, err := f.DatasetClient.GetVersionDimensions(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version)
		if err != nil {
			log.Event(ctx, "failed to get dimensions", log.ERROR, log.Error(err),
				log.Data{"dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		selValsLabelMap, err := f.getIDNameLookupFromDatasetAPI(ctx, userAccessToken, collectionID, datasetID, edition, version, name, selVals)
		if err != nil {
			log.Event(ctx, "failed to get options from dataset client for the selected values", log.ERROR, log.Error(err),
				log.Data{"dimension": name, "dataset_id": datasetID, "edition": edition, "version": version})
			setStatusCode(req, w, err)
			return
		}

		p := mapper.CreateHierarchyPage(req, h, d, fil, selValsLabelMap, dims, name, req.URL.Path, datasetID, ver.ReleaseDate, f.APIRouterVersion, lang)

		b, err := json.Marshal(p)
		if err != nil {
			log.Event(req.Context(), "failed to marshal json", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		templateBytes, err := f.Renderer.Do("dataset-filter/hierarchy", b)
		if err != nil {
			log.Event(req.Context(), "failed to render", log.ERROR, log.Error(err), log.Data{"filter_id": filterID})
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateBytes)
	})

}

type flatNodes struct {
	list  []*hierarchy.Child
	order []string
}

func (n flatNodes) addWithoutChildren(val hierarchy.Child, i int) {
	if !val.HasData {
		return
	}

	n.list[i] = &hierarchy.Child{
		Label:   val.Label,
		Links:   val.Links,
		HasData: val.HasData,
	}
}

func (n flatNodes) addWithChildren(val hierarchy.Child, i int) {
	n.list[i] = &hierarchy.Child{
		Label:            val.Label,
		Links:            val.Links,
		HasData:          val.HasData,
		NumberofChildren: val.NumberofChildren,
	}
}

// Flatten the geography hierarchy - please note this will only work for this particular hierarchy,
// need helper functions for other geog hierarchies too.
func (f *Filter) flattenGeographyTopLevel(ctx context.Context, instanceID string) (h hierarchy.Model, err error) {
	root, err := f.HierarchyClient.GetRoot(ctx, instanceID, "geography")
	if err != nil {
		return
	}

	if root.HasData {
		h.Label = root.Label
		h.Links = root.Links
		h.HasData = root.HasData
	}

	// Order: Great Britain, England and Wales, England, Northern Ireland, Scotland, Wales
	nodes := flatNodes{
		list:  make([]*hierarchy.Child, 6),
		order: []string{"K03000001", "K04000001", "E92000001", "N92000002", "S92000003", "W92000004"},
	}

	for _, val := range root.Children {
		if val.Links.Code.ID == nodes.order[0] {
			nodes.addWithoutChildren(val, 0)

			child, err := f.HierarchyClient.GetChild(ctx, instanceID, "geography", val.Links.Code.ID)
			if err != nil {
				return h, err
			}

			for _, childVal := range child.Children {

				if childVal.Links.Code.ID == nodes.order[1] {
					nodes.addWithoutChildren(childVal, 1)

					grandChild, err := f.HierarchyClient.GetChild(ctx, instanceID, "geography", childVal.Links.Code.ID)
					if err != nil {
						return h, err
					}

					for _, grandChildVal := range grandChild.Children {
						if grandChildVal.Links.Code.ID == nodes.order[2] {
							nodes.addWithChildren(grandChildVal, 2)
						}

						if grandChildVal.Links.Code.ID == nodes.order[5] {
							nodes.addWithChildren(grandChildVal, 5)
						}

					}
				}

				if childVal.Links.Code.ID == nodes.order[4] {
					nodes.addWithChildren(childVal, 4)
				}
			}
		}

		if val.Links.Code.ID == nodes.order[3] {
			nodes.addWithChildren(val, 3)
		}
	}

	//remove nil elements from list
	children := []hierarchy.Child{}
	for _, c := range nodes.list {
		if c != nil {
			children = append(children, *c)
		}
	}

	if len(children) == 0 {
		children = root.Children
	}

	h.Children = children
	return h, err
}
