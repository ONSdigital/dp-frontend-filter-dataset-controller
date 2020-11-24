package handlers

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/hierarchy"
	"github.com/ONSdigital/dp-api-clients-go/search"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
)

// mockgen -source=clients.go -package=handlers > mock_clients.go

// ClientError implements error interface with additional code method
type ClientError interface {
	Error() string
	Code() int
}

// FilterClient contains the methods expected for a filter client
type FilterClient interface {
	Checker(ctx context.Context, check *health.CheckState) error
	GetDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID string) (dims []filter.Dimension, err error)
	GetDimensionOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string) (fdv []filter.DimensionOption, err error)
	GetJobState(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterID string) (f filter.Model, err error)
	GetOutput(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID string) (f filter.Model, err error)
	GetDimension(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string) (dim filter.Dimension, err error)
	AddDimensionValue(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name, value string) error
	RemoveDimensionValue(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name, value string) error
	RemoveDimension(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string) (err error)
	AddDimension(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string) (err error)
	SetDimensionValues(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, options []string) error
	UpdateBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID string, m filter.Model, doSubmit bool) (filter.Model, error)
	CreateBlueprint(context.Context, string, string, string, string, string, string, string, []string) (string, error)
	GetPreview(context.Context, string, string, string, string, string) (filter.Preview, error)
}

// DatasetClient is an interface with methods required for a dataset client
type DatasetClient interface {
	Checker(ctx context.Context, check *health.CheckState) error
	Get(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m dataset.DatasetDetails, err error)
	GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string) (m dataset.Version, err error)
	GetVersionDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m dataset.VersionDimensions, err error)
	GetOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition, version, dimension string) (m dataset.Options, err error)
	GetVersionMetadata(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition, version string) (m dataset.Metadata, err error)
	GetEdition(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition string) (m dataset.Edition, err error)
}

// HierarchyClient contains methods expected for a hierarchy client
type HierarchyClient interface {
	Checker(ctx context.Context, check *health.CheckState) error
	GetRoot(ctx context.Context, instanceID, name string) (hierarchy.Model, error)
	GetChild(ctx context.Context, instanceID, name, code string) (hierarchy.Model, error)
}

// SearchClient contains methods expected for a search client
type SearchClient interface {
	Checker(ctx context.Context, check *health.CheckState) error
	Dimension(ctx context.Context, datasetID, edition, version, name, query string, params ...search.Config) (m *search.Model, err error)
}

// Renderer provides an interface for a service template renderer
type Renderer interface {
	Checker(ctx context.Context, check *health.CheckState) error
	Do(path string, b []byte) ([]byte, error)
}
