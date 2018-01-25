package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/clients/hierarchy"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// HierarchyUpdate controls the updating of a hierarchy job
func (f *Filter) HierarchyUpdate(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	filterID := vars["filterID"]
	name := vars["name"]
	code := vars["code"]

	if err := req.ParseForm(); err != nil {
		log.ErrorR(req, err, nil)
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

	fil, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
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
			setStatusCode(req, w, err)
			return
		}

		opts, err := f.FilterClient.GetDimensionOptions(filterID, name)
		if err != nil {
			log.ErrorR(req, err, nil)
		}

		for _, hv := range h.Children {
			for _, opt := range opts {
				if opt.Option == hv.Links.Self.ID {
					if _, ok := req.Form[hv.Links.Self.ID]; !ok {
						if err := f.FilterClient.RemoveDimensionValue(filterID, name, hv.Links.Self.ID); err != nil {
							log.ErrorR(req, err, nil)
						}
					}
				}
			}
		}

		wg.Done()
	}()

	var options []string
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

		options = append(options, k)
	}

	if err := f.FilterClient.AddDimensionValues(filterID, name, options); err != nil {
		log.TraceR(req, err.Error(), nil)
	}

	http.Redirect(w, req, redirectURI, 302)
}

func (f *Filter) addAllHierarchyLevel(w http.ResponseWriter, req *http.Request, fil filter.Model, name, code, redirectURI string) {

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
		setStatusCode(req, w, err)
		return
	}

	var options []string
	for _, child := range h.Children {
		options = append(options, child.Links.Self.ID)
	}
	if err := f.FilterClient.AddDimensionValues(fil.FilterID, name, options); err != nil {
		log.ErrorR(req, err, nil)
	}

	http.Redirect(w, req, redirectURI, 302)
}

func (f *Filter) removeAllHierarchyLevel(w http.ResponseWriter, req *http.Request, fil filter.Model, name, code, redirectURI string) {

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
		setStatusCode(req, w, err)
		return
	}

	for _, child := range h.Children {
		if err := f.FilterClient.RemoveDimensionValue(fil.FilterID, name, child.Links.Self.ID); err != nil {
			log.ErrorR(req, err, nil)
		}
	}

	http.Redirect(w, req, redirectURI, 302)
}

func (f *Filter) Hierarchy(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]
	name := vars["name"]
	code := vars["code"]

	fil, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
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
		setStatusCode(req, w, err)
		return
	}

	selVals, err := f.FilterClient.GetDimensionOptions(filterID, name)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	versionURL, err := url.Parse(fil.Links.Version.HRef)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	d, err := f.DatasetClient.Get(datasetID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}
	ver, err := f.DatasetClient.GetVersion(datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	allVals, err := f.DatasetClient.GetOptions(datasetID, edition, version, name)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	p := mapper.CreateHierarchyPage(h, d, fil, selVals, allVals, name, req.URL.Path, datasetID, ver.ReleaseDate)

	b, err := json.Marshal(p)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/hierarchy", b)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	w.Write(templateBytes)

}

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

	h.Children = make([]hierarchy.Child, 6)

	for _, val := range root.Children {
		// K03000001 Great Britain
		if val.Links.Code.ID == "K03000001" {
			h.Children[0].Label = val.Label
			h.Children[0].Links = val.Links
			h.Children[0].HasData = false

			if val.HasData {
				child, err := f.HierarchyClient.GetChild(instanceID, "geography", val.Links.Code.ID)
				if err != nil {
					return h, err
				}

				for _, childVal := range child.Children {
					// K04000001 England and Wales
					if childVal.Links.Code.ID == "K04000001" {
						h.Children[1].Label = childVal.Label
						h.Children[1].Links = childVal.Links
						h.Children[1].HasData = false

						if childVal.HasData {
							grandChild, err := f.HierarchyClient.GetChild(instanceID, "geography", childVal.Links.Code.ID)
							if err != nil {
								return h, err
							}

							for _, grandChildVal := range grandChild.Children {
								// E92000001 England
								if grandChildVal.Links.Code.ID == "E92000001" {
									h.Children[2].Label = grandChildVal.Label
									h.Children[2].Links = grandChildVal.Links
									h.Children[2].HasData = grandChildVal.HasData
									h.Children[2].NumberofChildren = grandChildVal.NumberofChildren
								}

								// W92000004 Wales
								if grandChildVal.Links.Code.ID == "W92000004" {
									h.Children[5].Label = grandChildVal.Label
									h.Children[5].Links = grandChildVal.Links
									h.Children[5].HasData = grandChildVal.HasData
									h.Children[5].NumberofChildren = grandChildVal.NumberofChildren
								}
							}
						}
					}

					// S92000003 Scotland
					if childVal.Links.Code.ID == "S92000003" {
						h.Children[4].Label = childVal.Label
						h.Children[4].Links = childVal.Links
						h.Children[4].HasData = childVal.HasData
						h.Children[4].NumberofChildren = childVal.NumberofChildren
					}
				}
			}
		}
		// N92000002 Northern Ireland
		if val.Links.Code.ID == "N92000002" {
			h.Children[3].Label = val.Label
			h.Children[3].HasData = val.HasData
			h.Children[3].Links = val.Links
			h.Children[3].NumberofChildren = val.NumberofChildren
		}
	}

	return h, err
}
