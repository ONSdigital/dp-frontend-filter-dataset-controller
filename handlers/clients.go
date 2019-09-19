package handlers

import (
	"context"

	"github.com/ONSdigital/go-ns/clients/codelist"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/go-ns/clients/hierarchy"
	"github.com/ONSdigital/go-ns/clients/search"
	"github.com/ONSdigital/go-ns/healthcheck"
)

// ClientError implements error interface with additional code method
type ClientError interface {
	error
	Code() int
}

// FilterClient contains the methods expected for a filter client
type FilterClient interface {
	healthcheck.Client
	GetDimensions(ctx context.Context, userAuthToken, serviceAuthToken, filterID string) (dims []filter.Dimension, err error)
	GetDimensionOptions(ctx context.Context, userAuthToken, serviceAuthToken, filterID, name string) (fdv []filter.DimensionOption, err error)
	GetJobState(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, filterID string) (f filter.Model, err error)
	GetOutput(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, filterOutputID string) (f filter.Model, err error)
	GetDimension(ctx context.Context, userAuthToken, serviceAuthToken, filterID, name string) (dim filter.Dimension, err error)
	AddDimensionValue(ctx context.Context, userAuthToken, serviceAuthToken, filterID, name, value string) error
	RemoveDimensionValue(ctx context.Context, userAuthToken, serviceAuthToken, filterID, name, value string) error
	RemoveDimension(ctx context.Context, userAuthToken, serviceAuthToken, filterID, name string) (err error)
	AddDimension(ctx context.Context, userAuthToken, serviceAuthToken, filterID, name string) (err error)
	AddDimensionValues(ctx context.Context, userAuthToken, serviceAuthToken, filterID, name string, options []string) error
	UpdateBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken string, m filter.Model, doSubmit bool) (filter.Model, error)
	CreateBlueprint(context.Context, string, string, string, string, string, string, []string) (string, error)
	GetPreview(context.Context, string, string, string, string) (filter.Preview, error)
}

// DatasetClient is an interface with methods required for a dataset client
type DatasetClient interface {
	healthcheck.Client
	Get(ctx context.Context, id string) (m dataset.Model, err error)
	GetEditions(ctx context.Context, id string) (m []dataset.Edition, err error)
	GetVersions(ctx context.Context, id, edition string) (m []dataset.Version, err error)
	GetVersion(ctx context.Context, id, edition, version string) (m dataset.Version, err error)
	GetDimensions(ctx context.Context, id, edition, version string) (m dataset.Dimensions, err error)
	GetOptions(ctx context.Context, id, edition, version, dimension string) (m dataset.Options, err error)
	GetVersionMetadata(ctx context.Context, id, edition, version string) (m dataset.Metadata, err error)
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
	GetRoot(instanceID, name string) (hierarchy.Model, error)
	GetChild(instanceID, name, code string) (hierarchy.Model, error)
}

// SearchClient contains methods expected for a search client
type SearchClient interface {
	healthcheck.Client
	Dimension(ctx context.Context, datasetID, edition, version, name, query string, params ...search.Config) (m *search.Model, err error)
}

// Renderer provides an interface for a service template renderer
type Renderer interface {
	healthcheck.Client
	Do(path string, b []byte) ([]byte, error)
}
