package routes

import (
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/handlers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/renderer"
	"github.com/gorilla/mux"
)

// Init initialises routes for the service
func Init(r *mux.Router) {
	rend := renderer.New()
	filter := handlers.NewFilter(rend)

	r.Path("/filters/{filterID}").Methods("GET").HandlerFunc(filter.PreviewPage)
	r.Path("/filters/{filterID}/dimensions").Methods("GET").HandlerFunc(filter.FilterOverview)
	r.Path("/filters/{filterID}/dimensions/age-range").Methods("GET").HandlerFunc(filter.AgeSelectorRange)
	r.Path("/filters/{filterID}/dimensions/age-list").Methods("GET").HandlerFunc(filter.AgeSelectorList)
	r.Path("/filters/{filterID}/dimensions/geography").Methods("GET").HandlerFunc(filter.Geography)

	r.Path("/filters/{filterID}/dimensions/{dimensionType}/{hierarchyID}").Methods("GET").HandlerFunc(filter.Hierarchy)
	r.Path("/filters/{filterID}/dimensions/{dimensionType}/{hierarchyID}/remove/{value}").HandlerFunc(filter.HierarchyRemove)
	r.Path("/filters/{filterID}/dimensions/{dimensionType}/{hierarchyID}/add/{value}").HandlerFunc(filter.HierarchyAdd)
	r.Path("/filters/{filterID}/dimensions/{dimensionType}/{hierarchyID}/add-all").HandlerFunc(filter.HierarchyAddAll)
	r.Path("/filters/{filterID}/dimensions/{dimensionType}/{hierarchyID}/remove-all").HandlerFunc(filter.HierarchyRemoveAll)
}
