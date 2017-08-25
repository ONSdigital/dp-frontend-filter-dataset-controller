package routes

import (
	"os"
	"time"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/handlers"
	"github.com/ONSdigital/go-ns/clients/codelist"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/clients/hierarchy"
	"github.com/ONSdigital/go-ns/clients/renderer"
	"github.com/ONSdigital/go-ns/healthcheck"
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
		log.ErrorC("failed to create form validator", err, nil)
	}

	rend := renderer.New(cfg.RendererURL)
	fc := filter.New(cfg.FilterAPIURL)
	dc := dataset.New(cfg.DatasetAPIURL)
	clc := codelist.New(cfg.CodeListAPIURL)
	hc := hierarchy.New(cfg.HierarchyAPIURL)
	filter := handlers.NewFilter(rend, fc, dc, clc, hc, v)

	go func() {
		for {
			timer := time.NewTimer(time.Second * 60)

			healthcheck.MonitorExternal(fc, dc, clc, hc, rend)

			<-timer.C
		}
	}()

	r.Path("/healthcheck").HandlerFunc(healthcheck.Do)

	r.Path("/filters/{filterID}").Methods("GET").HandlerFunc(filter.PreviewPage)
	r.Path("/filters/{filterID}/dimensions").Methods("GET").HandlerFunc(filter.FilterOverview)
	r.Path("/filters/{filterID}/dimensions/clear-all").HandlerFunc(filter.FilterOverviewClearAll)
	r.Path("/filters/{filterID}/dimensions/geography").Methods("GET").HandlerFunc(filter.Geography)

	r.Path("/filters/{filterID}/dimensions/{name}").Methods("GET").HandlerFunc(filter.DimensionSelector)
	r.Path("/filters/{filterID}/dimensions/{name}/remove-all").HandlerFunc(filter.DimensionRemoveAll)
	r.Path("/filters/{filterID}/dimensions/{name}/remove/{option}").HandlerFunc(filter.DimensionRemoveOne)
	r.Path("/filters/{filterID}/dimensions/{name}/range").Methods("POST").HandlerFunc(filter.AddRange)
	r.Path("/filters/{filterID}/dimensions/{name}/list").Methods("POST").HandlerFunc(filter.AddList)
	r.Path("/filters/{filterID}/dimensions/{name}{uri:.*}/remove-all").HandlerFunc(filter.DimensionRemoveAll)

	r.Path("/filters/{filterID}/hierarchies/{name}/update").Methods("POST").HandlerFunc(filter.HierarchyUpdate)
	r.Path("/filters/{filterID}/hierarchies/{name}/{uri:.*}/update").Methods("POST").HandlerFunc(filter.HierarchyUpdate)
	r.Path("/filters/{filterID}/hierarchies/{name}/remove/{option}").HandlerFunc(filter.HierarchyRemove)
	r.Path("/filters/{filterID}/hierarchies/{name}/{uri:.*}/remove/{option}").HandlerFunc(filter.HierarchyRemove)
	r.Path("/filters/{filterID}/hierarchies/{name}/remove-all").HandlerFunc(filter.HierarchyRemoveAll)
	r.Path("/filters/{filterID}/hierarchies/{name}/{uri:.*}/remove-all").HandlerFunc(filter.HierarchyRemoveAll)
	r.Path("/filters/{filterID}/hierarchies/{name}/{uri:.*}").Methods("GET").HandlerFunc(filter.Hierarchy)
	r.Path("/filters/{filterID}/hierarchies/{name}").Methods("GET").HandlerFunc(filter.Hierarchy)
}
