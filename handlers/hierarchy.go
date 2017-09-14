package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

// HierarchyRemoveAll allows the removing of all selected values in a hierarchy
func (f *Filter) HierarchyRemoveAll(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	filterID := vars["filterID"]
	name := vars["name"]

	log.Debug("name", log.Data{"name": name})

	if name == "CPI" {
		name = "goods-and-services"
	}

	if err := f.FilterClient.RemoveDimension(filterID, name); err != nil {
		log.ErrorR(req, err, nil)
	}

	if err := f.FilterClient.AddDimension(filterID, name); err != nil {
		log.ErrorR(req, err, nil)
	}

	curPath := req.URL.Path

	pathReg := regexp.MustCompile(`^(\/filters\/.+\/hierarchies\/.+)\/remove-all$`)
	pathSubs := pathReg.FindStringSubmatch(curPath)

	redirectURI := pathSubs[1]

	http.Redirect(w, req, redirectURI, 302)
}

// HierarchyUpdate ...
func (f *Filter) HierarchyUpdate(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	filterID := vars["filterID"]
	name := vars["name"]

	if err := req.ParseForm(); err != nil {
		log.ErrorR(req, err, nil)
		return
	}

	curPath := req.URL.Path

	pathReg := regexp.MustCompile(`^\/filters\/.+\/hierarchies\/(.+)\/update$`)
	pathSubs := pathReg.FindStringSubmatch(curPath)

	hierarchyPath := pathSubs[1]
	hierarchyPath = strings.Replace(hierarchyPath, "goods-and-services", "CPI", -1)

	var redirectURI string
	if len(req.Form["save-and-return"]) > 0 {
		redirectURI = fmt.Sprintf("/filters/%s/dimensions", filterID)
	} else {
		pathReg := regexp.MustCompile(`^(\/filters\/.+\/hierarchies\/.+)\/update$`)
		pathSubs := pathReg.FindStringSubmatch(req.URL.Path)
		if len(pathSubs) > 1 {
			redirectURI = pathSubs[1]
		}
	}

	if name == "CPI" {
		name = "goods-and-services"
	}

	if len(req.Form["add-all"]) > 0 {
		f.addAllHierarchyLevel(w, req, filterID, name, redirectURI, hierarchyPath)
		return
	}

	if len(req.Form["remove-all"]) > 0 {
		f.removeAllHierarchyLevel(w, req, filterID, name, redirectURI, hierarchyPath)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		h, err := f.HierarchyClient.GetHierarchy(hierarchyPath)
		if err != nil {
			log.ErrorR(req, err, nil)
		}

		opts, err := f.FilterClient.GetDimensionOptions(filterID, name)
		if err != nil {
			log.ErrorR(req, err, nil)
		}

		for _, hv := range h.Children {
			for _, opt := range opts {
				if opt.Option == hv.ID {
					if _, ok := req.Form[hv.ID]; !ok {
						if err := f.FilterClient.RemoveDimensionValue(filterID, name, hv.ID); err != nil {
							log.ErrorR(req, err, nil)
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

		if err := f.FilterClient.AddDimensionValue(filterID, name, k); err != nil {
			log.TraceR(req, err.Error(), nil)
			continue
		}
	}

	http.Redirect(w, req, redirectURI, 302)
}

func (f *Filter) addAllHierarchyLevel(w http.ResponseWriter, req *http.Request, filterID, name, redirectURI, hierarchyPath string) {

	h, err := f.HierarchyClient.GetHierarchy(hierarchyPath)
	if err != nil {
		log.ErrorR(req, err, nil)
		return
	}

	for _, child := range h.Children {
		if err := f.FilterClient.AddDimensionValue(filterID, name, child.ID); err != nil {
			log.ErrorR(req, err, nil)
		}
	}

	http.Redirect(w, req, redirectURI, 302)
}

func (f *Filter) removeAllHierarchyLevel(w http.ResponseWriter, req *http.Request, filterID, name, redirectURI, hierarchyPath string) {

	h, err := f.HierarchyClient.GetHierarchy(hierarchyPath)
	if err != nil {
		log.ErrorR(req, err, nil)
		return
	}

	for _, child := range h.Children {
		if err := f.FilterClient.RemoveDimensionValue(filterID, name, child.ID); err != nil {
			log.ErrorR(req, err, nil)
		}
	}

	http.Redirect(w, req, redirectURI, 302)
}

// HierarchyRemove removes a single value from a hierarchy
func (f *Filter) HierarchyRemove(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	filterID := vars["filterID"]
	name := vars["name"]
	option := vars["option"]

	if name == "CPI" {
		name = "goods-and-services"
	}

	if err := f.FilterClient.RemoveDimensionValue(filterID, name, option); err != nil {
		log.ErrorR(req, err, nil)
		return
	}

	curPath := req.URL.Path

	pathReg := regexp.MustCompile(`^(\/filters\/.+\/hierarchies\/.+)\/remove\/.+$`)
	pathSubs := pathReg.FindStringSubmatch(curPath)

	redirectURI := pathSubs[1]

	http.Redirect(w, req, redirectURI, 302)
}

// Hierarchy controls the rendering of the hierarchy template
func (f *Filter) Hierarchy(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	dimensionType := vars["name"]
	filterID := vars["filterID"]

	path := req.URL.Path

	pathReg := regexp.MustCompile(`^\/filters\/.+\/hierarchies\/(.+)$`)
	pathsubs := pathReg.FindStringSubmatch(path)
	if len(pathsubs) < 2 {
		log.Info("could not get hierarchy path", nil)
		return
	}

	hierarchyPath := pathsubs[1]

	// TODO: This will need to be removed when the hierarchy is updated
	hierarchyPath = strings.Replace(hierarchyPath, "goods-and-services", "CPI", -1)

	h, err := f.HierarchyClient.GetHierarchy(hierarchyPath)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	parents, err := getHierarchyParents(h.Parent)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if dimensionType == "CPI" {
		dimensionType = "goods-and-services"
	}

	selectedValues, err := f.FilterClient.GetDimensionOptions(filterID, dimensionType)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	codeid := "e44de4c4-d39e-4e2f-942b-3ca10584d078" // TODO: Remove this when the real code id becomes available
	idLabelMap, err := f.CodeListClient.GetIDNameMap(codeid)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var selectedLabels, selectedIDs []string
	for _, opt := range selectedValues {
		selectedIDs = append(selectedIDs, opt.Option)
		selectedLabels = append(selectedLabels, idLabelMap[opt.Option])
	}

	fil, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	versionURL := fil.DatasetFilterID
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dataset, err := f.DatasetClient.Get(datasetID)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ver, err := f.DatasetClient.GetVersion(datasetID, edition, version)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fil.Dimensions = []filter.ModelDimension{
		{
			Name:   dimensionType,
			Values: selectedLabels,
			IDs:    selectedIDs,
		},
	}

	p := mapper.CreateHierarchyPage(h, parents, dataset, fil, req.URL.Path, dimensionType, datasetID, ver.ReleaseDate)

	body, err := json.Marshal(p)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := f.Renderer.Do("dataset-filter/hierarchy", body)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(b); err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func getHierarchyParents(p hierarchy.Parent) ([]hierarchy.Parent, error) {
	var parents []hierarchy.Parent

	if p.URL != "" {
		parents = append(parents, p)

		resp, err := http.Get("http://localhost:22600" + p.URL)
		if err != nil {
			return parents, err
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return parents, err
		}
		defer resp.Body.Close()

		var h hierarchy.Model
		if err = json.Unmarshal(b, &h); err != nil {
			return parents, err
		}

		grandParents, err := getHierarchyParents(h.Parent)
		if err != nil {
			return parents, nil
		}

		parents = append(parents, grandParents...)
		return parents, nil
	}

	return parents, nil
}
