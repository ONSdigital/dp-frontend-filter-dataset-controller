package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/data"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// FilterOverview controls the render of the filter overview template
// Contains stubbed data for now - page to be populated by the API
func (f *Filter) FilterOverview(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	dims, err := f.FilterClient.GetDimensions(filterID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	codeID := "64d384f1-ea3b-445c-8fb8-aa453f96e58a" // time
	idNameLookup, err := f.CodeListClient.GetIdNameMap(codeID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	codeID = "e44de4c4-d39e-4e2f-942b-3ca10584d078" // goods-and-services
	map2, err := f.CodeListClient.GetIdNameMap(codeID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for k, v := range map2 {
		idNameLookup[k] = v
	}

	var dimensions []data.Dimension
	for _, dim := range dims {
		var vals []data.DimensionOption
		vals, err = f.FilterClient.GetDimensionOptions(filterID, dim.Name)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var values []string
		for _, val := range vals {
			values = append(values, idNameLookup[val.Option])
		}

		dimensions = append(dimensions, data.Dimension{
			Name:   dim.Name,
			Values: values,
		})
	}

	filter, err := f.FilterClient.GetJobState(filterID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dataset, err := f.DatasetClient.GetDataset(filterID, "2016", "v1")
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	p := mapper.CreateFilterOverview(dimensions, filter, dataset, filterID)

	b, err := json.Marshal(p)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateBytes, err := f.Renderer.Do("dataset-filter/filter-overview", b)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateBytes)
}

// FilterOverviewClearAll removes all selected options for all dimensions
func (f *Filter) FilterOverviewClearAll(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filterID := vars["filterID"]

	dims, err := f.FilterClient.GetDimensions(filterID)
	if err != nil {
		log.ErrorR(req, err, nil)
		return
	}

	for _, dim := range dims {
		if err := f.FilterClient.RemoveDimension(filterID, dim.Name); err != nil {
			log.ErrorR(req, err, nil)
			return
		}

		if err := f.FilterClient.AddDimension(filterID, dim.Name); err != nil {
			log.ErrorR(req, err, nil)
			return
		}
	}

	redirectURL := fmt.Sprintf("/filters/%s/dimensions", filterID)

	http.Redirect(w, req, redirectURL, 302)
}
