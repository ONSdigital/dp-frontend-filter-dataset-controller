package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/data"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// HierarchyAddAll allows the adding of all values in a hierarchy
func (f *Filter) HierarchyAddAll(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	filterID := vars["filterID"]

	//TODO: Get all fields from the hierarchy api - post them to the filter api
	//TODO: Use the regex: \/filters\/\w+(\/hierarchies\/.+)\/add-all to extract the url call to the hierarchy api

	uri := fmt.Sprintf("/filters/%s/dimensions/", filterID)

	http.Redirect(w, req, uri, 301)
}

// HierarchyRemoveAll allows the removing of all selected values in a hierarchy
func (f *Filter) HierarchyRemoveAll(w http.ResponseWriter, req *http.Request) {
	// TODO: Needs to make a call to the filter api to update job

	vars := mux.Vars(req)

	filterID := vars["filterID"]

	//TODO: Get all fields from the hierarchy api - post them to the filter api
	//TODO: Use the regex: \/filters\/\w+(\/hierarchies\/.+)\/add-all to extract the url call to the hierarchy api

	uri := fmt.Sprintf("/filters/%s/dimensions", filterID)

	http.Redirect(w, req, uri, 301)
}

// HierarchyAdd adds a single hierarchy value to a hierarchy
func (f *Filter) HierarchyAdd(w http.ResponseWriter, req *http.Request) {
	// TODO: Needs to make a call to the filter api to update job

	vars := mux.Vars(req)

	filterID := vars["filterID"]
	hierarchyID := vars["hierarchyID"]
	dimensionType := vars["name"]

	uri := fmt.Sprintf("/filters/%s/dimensions/%s/%s", filterID, dimensionType, hierarchyID)

	http.Redirect(w, req, uri, 301)
}

// HierarchyRemove removes a single value from a hierarchy
func (f *Filter) HierarchyRemove(w http.ResponseWriter, req *http.Request) {
	// TODO: Needs to make a call to the filter api to update job

	vars := mux.Vars(req)

	filterID := vars["filterID"]
	hierarchyID := vars["hierarchyID"]
	dimensionType := vars["name"]

	uri := fmt.Sprintf("/filters/%s/dimensions/%s/%s", filterID, dimensionType, hierarchyID)

	http.Redirect(w, req, uri, 301)
}

// Hierarchy controls the rendering of the hierarchy template
func (f *Filter) Hierarchy(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	dimensionType := vars["name"]
	hierarchyID := vars["hierarchyID"]
	filterID := vars["filterID"]

	var uri string
	if hierarchyID != "" {
		uri = "http://localhost:22600/hierarchies/CPI/" + hierarchyID
	} else {
		uri = "http://localhost:22600/hierarchies/CPI"
	}

	resp, err := http.Get(uri)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var h data.Hierarchy
	if err = json.Unmarshal(b, &h); err != nil {
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

	d := data.Dataset{
		ID:          "849209",
		ReleaseDate: "17 January 2017",
		NextRelease: "17 February 2017",
		Contact: data.Contact{
			Name:      "Matt Rout",
			Telephone: "07984593234",
			Email:     "matt@gmail.com",
		},
		Title: "Consumer Prices Index (COICOP): 2016",
	}

	fil := data.Filter{
		FilterID: filterID,
		Edition:  "12345",
		Dataset:  "849209",
		Version:  "2017",
		Dimensions: []data.Dimension{
			{
				Name:   dimensionType,
				Values: []string{"03.1 Clothing", "03.1.2 Garments", "03.2 Footwear including repairs"},
			},
		},
		Downloads: map[string]data.Download{
			"csv": {
				Size: "362783",
				URL:  "/",
			},
			"xls": {
				Size: "373929",
				URL:  "/",
			},
		},
	}

	met := data.Metadata{
		Name:        "goods and services",
		Description: "Goods and services provides information ....",
	}

	p := mapper.CreateHierarchyPage(h, parents, d, fil, met, req.URL.Path, dimensionType)

	body, err := json.Marshal(p)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err = f.r.Do("dataset-filter/hierarchy", body)
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

func getHierarchyParents(p data.Parent) ([]data.Parent, error) {
	var parents []data.Parent

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

		var h data.Hierarchy
		if err := json.Unmarshal(b, &h); err != nil {
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
