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
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
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
			h, err = f.HierarchyClient.GetRoot(fil.InstanceID, name)
		}
		if err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
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

func (f *Filter) addAllHierarchyLevel(w http.ResponseWriter, req *http.Request, fil filter.Model, name, code, redirectURI string) {

	var h hierarchy.Model
	var err error
	if len(code) > 0 {
		h, err = f.HierarchyClient.GetChild(fil.InstanceID, name, code)
	} else {
		h, err = f.HierarchyClient.GetRoot(fil.InstanceID, name)
	}
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, child := range h.Children {
		if err := f.FilterClient.AddDimensionValue(fil.FilterID, name, child.Links.Self.ID); err != nil {
			log.ErrorR(req, err, nil)
		}
	}

	http.Redirect(w, req, redirectURI, 302)
}

func (f *Filter) removeAllHierarchyLevel(w http.ResponseWriter, req *http.Request, fil filter.Model, name, code, redirectURI string) {

	var h hierarchy.Model
	var err error
	if len(code) > 0 {
		h, err = f.HierarchyClient.GetChild(fil.InstanceID, name, code)
	} else {
		h, err = f.HierarchyClient.GetRoot(fil.InstanceID, name)
	}
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
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
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var h hierarchy.Model
	if len(code) > 0 {
		h, err = f.HierarchyClient.GetChild(fil.InstanceID, name, code)
	} else {
		h, err = f.HierarchyClient.GetRoot(fil.InstanceID, name)
	}
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	selVals, err := f.FilterClient.GetDimensionOptions(filterID, name)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	versionURL, err := url.Parse(fil.Links.Version.HRef)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	datasetID, edition, version, err := helpers.ExtractDatasetInfoFromPath(versionURL.Path)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	d, err := f.DatasetClient.Get(datasetID)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ver, err := f.DatasetClient.GetVersion(datasetID, edition, version)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	allVals, err := f.DatasetClient.GetOptions(datasetID, edition, version, name)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	p := mapper.CreateHierarchyPage(h, d, fil, selVals, allVals, name, req.URL.Path, datasetID, ver.ReleaseDate)

	b, err := json.Marshal(p)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/hierarchy", b)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)

}
