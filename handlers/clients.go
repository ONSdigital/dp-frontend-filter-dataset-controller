package handlers

import "github.com/ONSdigital/dp-frontend-filter-dataset-controller/data"

// FilterClient contains the methods expected for a filter client
type FilterClient interface {
	GetDimensions(filterID string) (dims []data.FilterDimension, err error)
	GetDimensionOptions(filterID, name string) (fdv []data.DimensionOption, err error)
	GetJobState(filterID string) (f data.Filter, err error)
	GetDimension(filterID, name string) (dim data.FilterDimension, err error)
	AddDimensionValue(filterID, name, value string) error
	RemoveDimensionValue(filterID, name, value string) error
	RemoveDimension(filterID, name string) (err error)
	AddDimension(filterID, name string) (err error)
	AddDimensionValues(filterID, name string, options []string) error
}

// DatasetClient contains methods expected for a dataset client
type DatasetClient interface {
	GetDataset(id, edition, version string) (d data.Dataset, err error)
}

// CodelistClient contains methods expected for a codelist client
type CodelistClient interface {
	GetValues(id string) (vals data.DimensionValues, err error)
	GetIdNameMap(id string) (map[string]string, error)
}

// HierarchyClient contains methods expected for a heirarchy client
type HierarchyClient interface {
	GetHierarchy(path string) (h data.Hierarchy, err error)
}
