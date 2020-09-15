package routes

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/hierarchy"
	"github.com/ONSdigital/dp-api-clients-go/renderer"
	"github.com/ONSdigital/dp-api-clients-go/search"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/handlers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	"github.com/ONSdigital/go-ns/validator"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

// Clients represents a list of clients
type Clients struct {
	Filter             *filter.Client
	Dataset            *dataset.Client
	Hierarchy          *hierarchy.Client
	HealthcheckHandler func(w http.ResponseWriter, req *http.Request)
	Renderer           *renderer.Renderer
	Search             *search.Client
}

// Init initialises routes for the service
func Init(ctx context.Context, r *mux.Router, cfg *config.Config, clients *Clients) {

	fi, err := os.Open("rules.json")
	if err != nil {
		log.Event(ctx, "unable to open date rules", log.WARN, log.Error(err))
	}
	defer fi.Close()

	v, err := validator.New(fi)
	if err != nil {
		log.Event(ctx, "failed to validate date rules", log.WARN, log.Error(err))
	}

	apiRouterVersion, err := helpers.GetAPIRouterVersion(cfg.APIRouterURL)
	if err != nil {
		log.Event(ctx, "failed to obtain an api router version. Will assume that it is un-versioned", log.WARN, log.Error(err))
	}

	filter := handlers.NewFilter(clients.Renderer, clients.Filter, clients.Dataset,
		clients.Hierarchy, clients.Search, v, cfg.SearchAPIAuthToken, cfg.DownloadServiceURL, apiRouterVersion,
		cfg.EnableDatasetPreview)

	r.StrictSlash(true).Path("/health").HandlerFunc(clients.HealthcheckHandler)

	r.Path("/filter-outputs/{filterOutputID}.json").Methods("GET").HandlerFunc(filter.GetFilterJob)
	r.StrictSlash(true).Path("/filter-outputs/{filterOutputID}").Methods("GET").HandlerFunc(filter.OutputPage)

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

	// Enable profiling endpoint for authorised users
	if cfg.EnableProfiler {
		middlewareChain := alice.New(profileMiddleware(cfg.PprofToken)).Then(http.DefaultServeMux)
		r.PathPrefix("/debug").Handler(middlewareChain)
	}
}

// profileMiddleware to validate auth token before accessing endpoint
func profileMiddleware(token string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()

			pprofToken := req.Header.Get("Authorization")
			if pprofToken == "Bearer " || pprofToken != "Bearer "+token {
				log.Event(ctx, "invalid auth token", log.ERROR, log.Error(errors.New("invalid auth token")))
				w.WriteHeader(404)
				return
			}

			log.Event(ctx, "accessing profiling endpoint", log.INFO)
			h.ServeHTTP(w, req)
		})
	}
}
