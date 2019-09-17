package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/go-ns/clients/hierarchy"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// HierarchyUpdate controls the updating of a hierarchy job
func (f *Filter) HierarchyUpdate(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	cfg := config.Get()
	filterID := vars["filterID"]
	name := vars["name"]
	code := vars["code"]

	req = forwardFlorenceTokenIfRequired(req)
	ctx := req.Context()

	if err := req.ParseForm(); err != nil {
		log.ErrorCtx(ctx, err, nil)
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

	fil, err := f.FilterClient.GetJobState(req.Context(), cfg.ServiceAuthToken, "", filterID)
	if err != nil {
		log.InfoCtx(ctx, "failed to get job state", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	if len(req.Form["add-all"]) > 0 {
		f.addAllHierarchyLevel(w, req, fil, name, code, redirectURI)
		return
	}

	if len(req.Form["remove-all"]) > 0 {
		f.removeAllHierarchyLevel(w, req, fil, name, code, redirectURI)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		var h hierarchy.Model
		var err error
		if len(code) > 0 {
			h, err = f.HierarchyClient.GetChild(fil.InstanceID, name, code)
		} else {
			if name == "geography" {
				h, err = f.flattenGeographyTopLevel(fil.InstanceID)
			} else {
				h, err = f.HierarchyClient.GetRoot(fil.InstanceID, name)
			}

			// We include the value on the root as a selectable item, so append
			// the value on the root to the child to see if it has been removed by
			// the user
			h.Children = append(h.Children, hierarchy.Child{
				Links: h.Links,
			})
		}
		if err != nil {
			log.InfoCtx(ctx, "failed to get hierarchy node", log.Data{"error": err, "filter_id": filterID, "dimension": name, "code": code})
			setStatusCode(req, w, err)
			return
		}

		opts, err := f.FilterClient.GetDimensionOptions(req.Context(), cfg.ServiceAuthToken, filterID, name)
		if err != nil {
			log.ErrorCtx(ctx, err, nil)
		}

		for _, hv := range h.Children {
			for _, opt := range opts {
				if opt.Option == hv.Links.Self.ID {
					if _, ok := req.Form[hv.Links.Self.ID]; !ok {
						if err := f.FilterClient.RemoveDimensionValue(req.Context(), cfg.ServiceAuthToken, filterID, name, hv.Links.Self.ID); err != nil {
							log.ErrorCtx(ctx, err, nil)
						}
					}
				}
			}
		}

		wg.Done()
	}()

	for k := range req.Form {
		if k == "save-and-return" || k == ":uri" {
			continue
		}

		if strings.Contains(k, "redirect:") {
			redirectReg := regexp.MustCompile(`^redirect:(.+)$`)
			redirectSubs := redirectReg.FindStringSubmatch(k)
			redirectURI = redirectSubs[1]
			continue
		}

		if err := f.FilterClient.AddDimensionValue(req.Context(), cfg.ServiceAuthToken, filterID, name, k); err != nil {
			log.InfoCtx(ctx, err.Error(), nil)
		}
	}

	http.Redirect(w, req, redirectURI, 302)
}

func (f *Filter) addAllHierarchyLevel(w http.ResponseWriter, req *http.Request, fil filter.Model, name, code, redirectURI string) {
	cfg := config.Get()
	req = forwardFlorenceTokenIfRequired(req)
	ctx := req.Context()

	var h hierarchy.Model
	var err error
	if len(code) > 0 {
		h, err = f.HierarchyClient.GetChild(fil.InstanceID, name, code)
	} else {
		if name == "geography" {
			h, err = f.flattenGeographyTopLevel(fil.InstanceID)
		} else {
			h, err = f.HierarchyClient.GetRoot(fil.InstanceID, name)
		}
	}
	if err != nil {
		log.InfoCtx(ctx, "failed to get hierarchy node", log.Data{"error": err, "filter_id": fil.FilterID, "dimension": name, "code": code})
		setStatusCode(req, w, err)
		return
	}

	var options []string
	for _, child := range h.Children {
		options = append(options, child.Links.Self.ID)
	}
	if err := f.FilterClient.AddDimensionValues(req.Context(), cfg.ServiceAuthToken, fil.FilterID, name, options); err != nil {
		log.ErrorCtx(ctx, err, nil)
	}

	http.Redirect(w, req, redirectURI, 302)
}

func (f *Filter) removeAllHierarchyLevel(w http.ResponseWriter, req *http.Request, fil filter.Model, name, code, redirectURI string) {
	cfg := config.Get()
	req = forwardFlorenceTokenIfRequired(req)
	ctx := req.Context()

	var h hierarchy.Model
	var err error
	if len(code) > 0 {
		h, err = f.HierarchyClient.GetChild(fil.InstanceID, name, code)
	} else {
		if name == "geography" {
			h, err = f.flattenGeographyTopLevel(fil.InstanceID)
		} else {
			h, err = f.HierarchyClient.GetRoot(fil.InstanceID, name)
		}
	}
	if err != nil {
		log.InfoCtx(ctx, "failed to get hierarchy node", log.Data{"error": err, "filter_id": fil.FilterID, "dimension": name, "code": code})
		setStatusCode(req, w, err)
		return
	}

	for _, child := range h.Children {
		if err := f.FilterClient.RemoveDimensionValue(req.Context(), cfg.ServiceAuthToken, fil.FilterID, name, child.Links.Self.ID); err != nil {
			log.ErrorCtx(ctx, err, nil)
		}
	}

	http.Redirect(w, req, redirectURI, 302)
}

func (f *Filter) Hierarchy(w http.ResponseWriter, req *http.Request) {
	cfg := config.Get()
	vars := mux.Vars(req)
	filterID := vars["filterID"]
	name := vars["name"]
	code := vars["code"]
	ctx := req.Context()

	req = forwardFlorenceTokenIfRequired(req)

	fil, err := f.FilterClient.GetJobState(req.Context(), cfg.ServiceAuthToken, "", filterID)
	if err != nil {
		log.InfoCtx(ctx, "failed to get job state", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	var h hierarchy.Model
	if len(code) > 0 {
		h, err = f.HierarchyClient.GetChild(fil.InstanceID, name, code)
	} else {
		if name == "geography" {
			h, err = f.flattenGeographyTopLevel(fil.InstanceID)
		} else {
			h, err = f.HierarchyClient.GetRoot(fil.InstanceID, name)
		}
	}
	if err != nil {
		log.InfoCtx(ctx, "failed to get hierarchy node", log.Data{"error": err, "filter_id": filterID, "dimension": name, "code": code})
		setStatusCode(req, w, err)
		return
	}

	selVals, err := f.FilterClient.GetDimensionOptions(req.Context(), cfg.ServiceAuthToken, filterID, name)
	if err != nil {
		log.InfoCtx(ctx, "failed to get options from filter client", log.Data{"error": err, "filter_id": filterID, "dimension": name})
		setStatusCode(req, w, err)
		return
	}

	versionURL, err := url.Parse(fil.Links.Version.HRef)
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

	d, err := f.DatasetClient.Get(req.Context(), datasetID)
	if err != nil {
		log.InfoCtx(req.Context(), "failed to get dataset", log.Data{"error": err, "dataset_id": datasetID})
		setStatusCode(req, w, err)
		return
	}
	ver, err := f.DatasetClient.GetVersion(req.Context(), datasetID, edition, version)
	if err != nil {
		log.InfoCtx(req.Context(), "failed to get version", log.Data{"error": err, "dataset_id": datasetID, "edition": edition, "version": version})
		setStatusCode(req, w, err)
		return
	}

	allVals, err := f.DatasetClient.GetOptions(req.Context(), datasetID, edition, version, name)
	if err != nil {
		log.InfoCtx(ctx, "failed to get options from dataset client",
			log.Data{"error": err, "dataset_id": datasetID, "edition": edition, "version": version})
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

	p := mapper.CreateHierarchyPage(req.Context(), h, d, fil, selVals, allVals, dims, name, req.URL.Path, datasetID, ver.ReleaseDate)

	b, err := json.Marshal(p)
	if err != nil {
		log.InfoCtx(req.Context(), "failed to marshal json", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/hierarchy", b)
	if err != nil {
		log.InfoCtx(req.Context(), "failed to render", log.Data{"error": err, "filter_id": filterID})
		setStatusCode(req, w, err)
		return
	}

	w.Write(templateBytes)

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
func (f *Filter) flattenGeographyTopLevel(instanceID string) (h hierarchy.Model, err error) {
	root, err := f.HierarchyClient.GetRoot(instanceID, "geography")
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

			child, err := f.HierarchyClient.GetChild(instanceID, "geography", val.Links.Code.ID)
			if err != nil {
				return h, err
			}

			for _, childVal := range child.Children {

				if childVal.Links.Code.ID == nodes.order[1] {
					nodes.addWithoutChildren(childVal, 1)

					grandChild, err := f.HierarchyClient.GetChild(instanceID, "geography", childVal.Links.Code.ID)
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
