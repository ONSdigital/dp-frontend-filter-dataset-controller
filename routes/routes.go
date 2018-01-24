package routes

import (
	"os"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/handlers"
	"github.com/ONSdigital/go-ns/clients/codelist"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/clients/hierarchy"
	"github.com/ONSdigital/go-ns/clients/renderer"
	"github.com/ONSdigital/go-ns/clients/search"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/validator"
	"github.com/gorilla/mux"
)

// Init initialises routes for the service
func Init(r *mux.Router) (*renderer.Renderer, *filter.Client, *dataset.Client, *codelist.Client, *hierarchy.Client, *search.Client) {
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
	sc := search.New(cfg.SearchAPIURL)
	filter := handlers.NewFilter(rend, fc, dc, clc, hc, sc, v)

	r.StrictSlash(true).Path("/healthcheck").HandlerFunc(healthcheck.Do)

	r.StrictSlash(true).Path("/filter-outputs/{filterOutputID}").Methods("GET").HandlerFunc(filter.PreviewPage)

	r.StrictSlash(true).Path("/filters/{filterID}/submit").Methods("POST").HandlerFunc(filter.Submit)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions").Methods("GET").HandlerFunc(filter.FilterOverview)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/clear-all").HandlerFunc(filter.FilterOverviewClearAll)

	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/time").Methods("GET").HandlerFunc(filter.Time)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/time/update").Methods("POST").HandlerFunc(filter.UpdateTime)

	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/age").Methods("GET").HandlerFunc(filter.Age)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/age/update").Methods("POST").HandlerFunc(filter.UpdateAge)

	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/search").Methods("GET").HandlerFunc(filter.Search)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/search/update").HandlerFunc(filter.SearchUpdate)

	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}").Methods("GET").HandlerFunc(filter.DimensionSelector)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/remove-all").HandlerFunc(filter.DimensionRemoveAll)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/add-all").HandlerFunc(filter.DimensionAddAll)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/update").HandlerFunc(filter.HierarchyUpdate)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/{code}/update").HandlerFunc(filter.HierarchyUpdate)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/{parent}/remove/{option}").HandlerFunc(filter.DimensionRemoveOne)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/remove/{option}").HandlerFunc(filter.DimensionRemoveOne)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/list").Methods("POST").HandlerFunc(filter.AddList)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}{uri:.*}/remove-all").HandlerFunc(filter.DimensionRemoveAll)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/options.json").HandlerFunc(filter.GetSelectedDimensionOptionsJSON)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/all-options.json").HandlerFunc(filter.GetAllDimensionOptionsJSON)
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/{code}").Methods("GET").HandlerFunc(filter.Hierarchy)

	r.StrictSlash(true).Path("/filters/{filterID}/use-latest-version").HandlerFunc(filter.UseLatest)

	return rend, fc, dc, clc, hc, sc
}
