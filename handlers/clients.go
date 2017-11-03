package handlers

import (
	"github.com/ONSdigital/go-ns/clients/codelist"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/clients/hierarchy"
	"github.com/ONSdigital/go-ns/healthcheck"
)

// FilterClient contains the methods expected for a filter client
type FilterClient interface {
	healthcheck.Client
	GetDimensions(filterID string) (dims []filter.Dimension, err error)
	GetDimensionOptions(filterID, name string) (fdv []filter.DimensionOption, err error)
	GetJobState(filterID string) (f filter.Model, err error)
	GetOutput(filterOutputID string) (f filter.Model, err error)
	GetDimension(filterID, name string) (dim filter.Dimension, err error)
	AddDimensionValue(filterID, name, value string) error
	RemoveDimensionValue(filterID, name, value string) error
	RemoveDimension(filterID, name string) (err error)
	AddDimension(filterID, name string) (err error)
	AddDimensionValues(filterID, name string, options []string) error
	UpdateBlueprint(m filter.Model, doSubmit bool) (filter.Model, error)
	CreateBlueprint(string, []string) (string, error)
}

// DatasetClient is an interface with methods required for a dataset client
type DatasetClient interface {
	healthcheck.Client
	Get(id string) (m dataset.Model, err error)
	GetEditions(id string) (m []dataset.Edition, err error)
	GetVersions(id, edition string) (m []dataset.Version, err error)
	GetVersion(id, edition, version string) (m dataset.Version, err error)
	GetDimensions(id, edition, version string) (m dataset.Dimensions, err error)
	GetOptions(id, edition, version, dimension string) (m dataset.Options, err error)
}

// CodelistClient contains methods expected for a codelist client
type CodelistClient interface {
	healthcheck.Client
	GetValues(id string) (vals codelist.DimensionValues, err error)
	GetIDNameMap(id string) (map[string]string, error)
}

// HierarchyClient contains methods expected for a heirarchy client
type HierarchyClient interface {
	healthcheck.Client
	GetHierarchy(path string) (h hierarchy.Model, err error)
}

// Renderer provides an interface for a service template renderer
type Renderer interface {
	healthcheck.Client
	Do(path string, b []byte) ([]byte, error)
}
