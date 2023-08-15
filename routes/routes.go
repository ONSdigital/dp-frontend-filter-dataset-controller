package routes

import (
	"context"
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/hierarchy"
	"github.com/ONSdigital/dp-api-clients-go/v2/search"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/handlers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	render "github.com/ONSdigital/dp-renderer/v2"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

// Clients represents a list of clients
type Clients struct {
	Filter             *filter.Client
	Dataset            *dataset.Client
	Hierarchy          *hierarchy.Client
	HealthcheckHandler func(w http.ResponseWriter, req *http.Request)
	Render             *render.Render
	Search             *search.Client
	Zebedee            *zebedee.Client
}

// Init initialises routes for the service
func Init(ctx context.Context, r *mux.Router, cfg *config.Config, clients *Clients) {
	apiRouterVersion, err := helpers.GetAPIRouterVersion(cfg.APIRouterURL)
	if err != nil {
		log.Warn(ctx, "failed to obtain an api router version. Will assume that it is un-versioned", log.FormatErrors([]error{err}))
	}

	f := handlers.NewFilter(clients.Render, clients.Filter, clients.Dataset,
		clients.Hierarchy, clients.Search, clients.Zebedee, apiRouterVersion, cfg)

	r.StrictSlash(true).Path("/health").HandlerFunc(clients.HealthcheckHandler)

	r.Path("/filter-outputs/{filterOutputID}.json").Methods("GET").HandlerFunc(f.GetFilterJob())
	r.StrictSlash(true).Path("/filter-outputs/{filterOutputID}").Methods("GET").HandlerFunc(f.OutputPage())

	r.StrictSlash(true).Path("/filters/{filterID}/submit").Methods("POST").HandlerFunc(f.Submit())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions").Methods("GET").HandlerFunc(f.FilterOverview())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/clear-all").HandlerFunc(f.FilterOverviewClearAll())

	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/time").Methods("GET").HandlerFunc(f.Time())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/time/update").Methods("POST").HandlerFunc(f.UpdateTime())

	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/age").Methods("GET").HandlerFunc(f.Age())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/age/update").Methods("POST").HandlerFunc(f.UpdateAge())

	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/search").Methods("GET").HandlerFunc(f.Search())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/search/update").HandlerFunc(f.SearchUpdate())

	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}").Methods("GET").HandlerFunc(f.DimensionSelector())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/remove-all").HandlerFunc(f.DimensionRemoveAll())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/add-all").HandlerFunc(f.DimensionAddAll())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/update").HandlerFunc(f.HierarchyUpdate())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/{code}/update").HandlerFunc(f.HierarchyUpdate())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/{parent}/remove/{option}").HandlerFunc(f.DimensionRemoveOne())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/remove/{option}").HandlerFunc(f.DimensionRemoveOne())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/list").Methods("POST").HandlerFunc(f.AddList())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}{uri:.*}/remove-all").HandlerFunc(f.DimensionRemoveAll())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/options.json").HandlerFunc(f.GetSelectedDimensionOptionsJSON())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/all-options.json").HandlerFunc(f.GetAllDimensionOptionsJSON())
	r.StrictSlash(true).Path("/filters/{filterID}/dimensions/{name}/{code}").Methods("GET").HandlerFunc(f.Hierarchy())

	r.StrictSlash(true).Path("/filters/{filterID}/use-latest-version").HandlerFunc(f.UseLatest())

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
				log.Error(ctx, "invalid auth token", errors.New("invalid auth token"))
				w.WriteHeader(404)
				return
			}

			log.Info(ctx, "accessing profiling endpoint")
			h.ServeHTTP(w, req)
		})
	}
}
