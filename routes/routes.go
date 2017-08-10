package routes

import (
	"os"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/codelist"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dataset"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/filter"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/handlers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/renderer"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/validator"
	"github.com/gorilla/mux"
)

// Init initialises routes for the service
func Init(r *mux.Router) {
	cfg := config.Get()

	fi, err := os.Open("rules.json")
	if err != nil {
		log.ErrorC("could not open rules for validation", err, nil)
	}
	defer fi.Close()

	v, err := validator.New(fi)
	if err != nil {
		log.ErrorC("failed to crare form validator", err, nil)
	}

	rend := renderer.New()
	fc := filter.New(cfg.FilterAPIURL)
	dc := dataset.New(cfg.DatasetAPIURL)
	clc := codelist.New(cfg.CodeListAPIURL)
	filter := handlers.NewFilter(rend, fc, dc, clc, v)

	r.Path("/filters/{filterID}").Methods("GET").HandlerFunc(filter.PreviewPage)
	r.Path("/filters/{filterID}/dimensions").Methods("GET").HandlerFunc(filter.FilterOverview)
	r.Path("/filters/{filterID}/dimensions/geography").Methods("GET").HandlerFunc(filter.Geography)

	r.Path("/filters/{filterID}/dimensions/{name}").Methods("GET").HandlerFunc(filter.DimensionSelector)
	r.Path("/filters/{filterID}/dimensions/{name}/remove-all").HandlerFunc(filter.DimensionRemoveAll)
	r.Path("/filters/{filterID}/dimensions/{name}/remove/{option}").HandlerFunc(filter.DimensionRemoveOne)
	r.Path("/filters/{filterID}/dimensions/{name}/range").Methods("POST").HandlerFunc(filter.AddRange)
	r.Path("/filters/{filterID}/dimensions/{name}/list").Methods("POST").HandlerFunc(filter.AddList)

	r.Path("/filters/{filterID}/{uri:.*}/{name}{uri:.*}/remove-all").HandlerFunc(filter.DimensionRemoveAll)

	r.Path("/filters/{filterID}/hierarchies/{name}").Methods("GET").HandlerFunc(filter.Hierarchy)
	r.Path("/filters/{filterID}/hierarchies/{name}/{hierarchyID}").Methods("GET").HandlerFunc(filter.Hierarchy)
	r.Path("/filters/{filterID}/hierarchies/{name}/{hierarchyID}/remove/{value}").HandlerFunc(filter.HierarchyRemove)
	r.Path("/filters/{filterID}/hierarchies/{name}/{hierarchyID}/add/{value}").HandlerFunc(filter.HierarchyAdd)
	r.Path("/filters/{filterID}/hierarchies/{uri:.*}/add-all").HandlerFunc(filter.HierarchyAddAll)

}
