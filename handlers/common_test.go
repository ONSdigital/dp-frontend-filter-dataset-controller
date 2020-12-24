package handlers

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	gomock "github.com/golang/mock/gomock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetIDNameLookupFromDatasetAPI(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := context.Background()
	mockUserAuthToken := "testUserAuthToken"
	mockServiceAuthToken := "testServiceAuthToken"
	mockCollectionID := "testCollectionID"

	datasetID := "abcde"
	name := "aggregate"
	edition := "2017"
	version := "1"

	cfg := &config.Config{
		SearchAPIAuthToken:   mockServiceAuthToken,
		DownloadServiceURL:   "",
		EnableDatasetPreview: false,
	}

	Convey("Given a filter initialised with a mocked DatasetClient ", t, func() {
		mdc := NewMockDatasetClient(mockCtrl)

		Convey("a set of existing filter dimension options is correctly mapped to lables returned by dataset API GetOptions", func() {
			maxDatasetOptions := 100
			cfg.MaxDatasetOptions = maxDatasetOptions
			f := NewFilter(nil, nil, mdc, nil, nil, nil, "/v1", cfg)
			filterOptions := filter.DimensionOptions{
				Items: []filter.DimensionOption{
					{Option: "op1"},
					{Option: "op2"},
				},
			}
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				dataset.QueryParams{IDs: []string{"op1", "op2"}}).Return(datasetOptionsFromIDs([]string{"op1", "op2"}), nil)
			idLabelMap, err := f.getIDNameLookupFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name, filterOptions)
			So(err, ShouldBeNil)
			So(idLabelMap, ShouldResemble, map[string]string{
				"op1": "This is option op1",
				"op2": "This is option op2",
			})
		})

		Convey("a set of existing filter dimension options is correctly mapped to lables returned by dataset API GetOptions in multiple batches", func() {
			maxDatasetOptions := 2
			cfg.MaxDatasetOptions = maxDatasetOptions
			f := NewFilter(nil, nil, mdc, nil, nil, nil, "/v1", cfg)
			filterOptions := filter.DimensionOptions{
				Items: []filter.DimensionOption{
					{Option: "op1"}, // belongs to first batch
					{Option: "op4"}, // belongs to first batch
					{Option: "op5"}, // belongs to second batch
				},
			}
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				dataset.QueryParams{IDs: []string{"op1", "op4"}}).Return(datasetOptionsFromIDs([]string{"op1", "op4"}), nil)
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				dataset.QueryParams{IDs: []string{"op5"}}).Return(datasetOptionsFromIDs([]string{"op5"}), nil)
			idLabelMap, err := f.getIDNameLookupFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name, filterOptions)
			So(err, ShouldBeNil)
			So(idLabelMap, ShouldResemble, map[string]string{
				"op1": "This is option op1",
				"op4": "This is option op4",
				"op5": "This is option op5",
			})
		})

		Convey("a set of filter dimension options containing inexistent options returns the expected error, but the existing dimensions are correctly mapped", func() {
			maxDatasetOptions := 100
			cfg.MaxDatasetOptions = maxDatasetOptions
			f := NewFilter(nil, nil, mdc, nil, nil, nil, "/v1", cfg)
			filterOptions := filter.DimensionOptions{
				Items: []filter.DimensionOption{
					{Option: "op1"},
					{Option: "inexistent"},
				},
			}
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				dataset.QueryParams{IDs: []string{"op1", "inexistent"}}).Return(datasetOptionsFromIDs([]string{"op1"}), nil)
			expectedErr := errors.New("could not find all required filter options in dataset API")
			idLabelMap, err := f.getIDNameLookupFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name, filterOptions)
			So(err, ShouldResemble, expectedErr)
			So(idLabelMap, ShouldResemble, map[string]string{
				"op1":        "This is option op1",
				"inexistent": "",
			})
		})

		Convey("if dataset API GetOptions returns an error, the same error is returned and the next batch is not requested", func() {
			maxDatasetOptions := 1
			cfg.MaxDatasetOptions = maxDatasetOptions
			f := NewFilter(nil, nil, mdc, nil, nil, nil, "/v1", cfg)
			filterOptions := filter.DimensionOptions{
				Items: []filter.DimensionOption{{Option: "op1"}, {Option: "op2"}},
			}
			expectedErr := errors.New("error getting options from Dataset API")
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				dataset.QueryParams{IDs: []string{"op1"}}).Return(dataset.Options{}, expectedErr)
			idLabelMap, err := f.getIDNameLookupFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name, filterOptions)
			So(err, ShouldResemble, expectedErr)
			So(idLabelMap, ShouldResemble, map[string]string{"op1": "", "op2": ""})
		})
	})
}

func TestGetDimensionOptionsFromFilterAPI(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := context.Background()
	mockUserAuthToken := "testUserAuthToken"
	mockServiceAuthToken := "testServiceAuthToken"
	mockCollectionID := "testCollectionID"

	filterID := "testFilter"
	name := "aggregate"

	cfg := &config.Config{
		SearchAPIAuthToken:   mockServiceAuthToken,
		DownloadServiceURL:   "",
		EnableDatasetPreview: false,
	}

	Convey("Given a filter initialised with a mocked FilterClient ", t, func() {
		mfc := NewMockFilterClient(mockCtrl)

		Convey("given that the number of filter options is smaller than a batch, then all options are returned after a single batch getDimensionOptions call", func() {
			batchSize := 100
			cfg.BatchSizeLimit = batchSize
			f := NewFilter(nil, mfc, nil, nil, nil, nil, "/v1", cfg)
			mfc.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name,
				filter.QueryParams{Offset: 0, Limit: batchSize}).Return(filterOptions(0, batchSize), nil)
			opts, err := f.GetDimensionOptionsFromFilterAPI(ctx, mockUserAuthToken, mockCollectionID, filterID, name)
			So(err, ShouldBeNil)
			So(opts, ShouldResemble, filterOptions(0, 0))
		})

		Convey("given that the number of filter options is greater than a batch, then all options are returned after multiple batch getDimensionOptions calls", func() {
			batchSize := 3
			cfg.BatchSizeLimit = batchSize
			f := NewFilter(nil, mfc, nil, nil, nil, nil, "/v1", cfg)
			mfc.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name,
				filter.QueryParams{Offset: 0, Limit: batchSize}).Return(filterOptions(0, batchSize), nil)
			mfc.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name,
				filter.QueryParams{Offset: batchSize, Limit: batchSize}).Return(filterOptions(batchSize, batchSize), nil)
			opts, err := f.GetDimensionOptionsFromFilterAPI(ctx, mockUserAuthToken, mockCollectionID, filterID, name)
			So(err, ShouldBeNil)
			So(opts, ShouldResemble, filterOptions(0, 0))
		})

		Convey("if filter API GetDimensionOptions returns an error, the same error is returned and the next batch is not requested", func() {
			batchSize := 2
			cfg.BatchSizeLimit = batchSize
			f := NewFilter(nil, mfc, nil, nil, nil, nil, "/v1", cfg)
			expectedErr := errors.New("error getting options from Filter API")
			mfc.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name,
				filter.QueryParams{Offset: 0, Limit: batchSize}).Return(filter.DimensionOptions{}, expectedErr)
			opts, err := f.GetDimensionOptionsFromFilterAPI(ctx, mockUserAuthToken, mockCollectionID, filterID, name)
			So(err, ShouldResemble, expectedErr)
			So(opts, ShouldResemble, filter.DimensionOptions{})
		})

	})
}

func TestGetDimensionOptionsFromDataseAPI(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := context.Background()
	mockUserAuthToken := "testUserAuthToken"
	mockServiceAuthToken := "testServiceAuthToken"
	mockCollectionID := "testCollectionID"

	datasetID := "abcde"
	name := "aggregate"
	edition := "2017"
	version := "1"

	cfg := &config.Config{
		SearchAPIAuthToken:   mockServiceAuthToken,
		DownloadServiceURL:   "",
		EnableDatasetPreview: false,
	}

	Convey("Given a filter initialised with a mocked DatasetClient ", t, func() {
		mdc := NewMockDatasetClient(mockCtrl)

		Convey("given that the number of dataset options is smaller than a batch, then all options are returned after a single batch getOptions call", func() {
			batchSize := 100
			cfg.BatchSizeLimit = batchSize
			f := NewFilter(nil, nil, mdc, nil, nil, nil, "/v1", cfg)
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				dataset.QueryParams{Offset: 0, Limit: batchSize}).Return(datasetOptions(0, batchSize), nil)
			opts, err := f.GetDimensionOptionsFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name)
			So(err, ShouldBeNil)
			So(opts, ShouldResemble, datasetOptions(0, 0))
		})

		Convey("given that the number of dataset options is greater than a batch, then all options are returned after multiple batch getOptions calls", func() {
			batchSize := 3
			cfg.BatchSizeLimit = batchSize
			f := NewFilter(nil, nil, mdc, nil, nil, nil, "/v1", cfg)
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				dataset.QueryParams{Offset: 0, Limit: batchSize}).Return(datasetOptions(0, batchSize), nil)
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				dataset.QueryParams{Offset: batchSize, Limit: batchSize}).Return(datasetOptions(batchSize, batchSize), nil)
			opts, err := f.GetDimensionOptionsFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name)
			So(err, ShouldBeNil)
			So(opts, ShouldResemble, datasetOptions(0, 0))
		})

		Convey("if dataset API GetOptions returns an error, the same error is returned and the next batch is not requested", func() {
			batchSize := 2
			cfg.BatchSizeLimit = batchSize
			f := NewFilter(nil, nil, mdc, nil, nil, nil, "/v1", cfg)
			expectedErr := errors.New("error getting options from Dataset API")
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				dataset.QueryParams{Offset: 0, Limit: batchSize}).Return(dataset.Options{}, expectedErr)
			opts, err := f.GetDimensionOptionsFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name)
			So(err, ShouldResemble, expectedErr)
			So(opts, ShouldResemble, dataset.Options{})
		})

	})
}

// datasetOptionsFromIDs returns a mocked dataset.Options struct according to the provided list of IDs
func datasetOptionsFromIDs(ids []string) dataset.Options {
	items := []dataset.Option{}
	for _, id := range ids {
		items = append(items, dataset.Option{
			Label:  fmt.Sprintf("This is option %s", id),
			Option: id,
		})
	}
	o := dataset.Options{
		Offset:     0,
		Limit:      0,
		TotalCount: len(items),
	}
	o.Items = items
	o.Count = len(o.Items)
	return o
}

// datasetOptions returns a mocked dataset.Options struct according to the provided offset and limit
func datasetOptions(offset, limit int) dataset.Options {
	allItems := []dataset.Option{
		{
			Label:  "This is option op1",
			Option: "op1",
		},
		{
			Label:  "This is option op2",
			Option: "op2",
		},
		{
			Label:  "This is option op3",
			Option: "op3",
		},
		{
			Label:  "This is option op4",
			Option: "op4",
		},
		{
			Label:  "This is option op5",
			Option: "op5",
		},
	}
	o := dataset.Options{
		Offset:     offset,
		Limit:      limit,
		TotalCount: len(allItems),
	}
	o.Items = sliceDatasetOptions(allItems, offset, limit)
	o.Count = len(o.Items)
	return o
}

// filterOptions returns a mocked filter.Options struct according to the provided offset and limit
func filterOptions(offset, limit int) filter.DimensionOptions {
	allItems := []filter.DimensionOption{
		{Option: "op1"},
		{Option: "op2"},
		{Option: "op3"},
		{Option: "op4"},
		{Option: "op5"},
	}
	o := filter.DimensionOptions{
		Offset:     offset,
		Limit:      limit,
		TotalCount: len(allItems),
	}
	o.Items = sliceFilterOptions(allItems, offset, limit)
	o.Count = len(o.Items)
	return o
}

func sliceDatasetOptions(full []dataset.Option, offset, limit int) (sliced []dataset.Option) {
	if offset > len(full) {
		return []dataset.Option{}
	}
	end := offset + limit
	if limit == 0 || end > len(full) {
		end = len(full)
	}
	return full[offset:end]
}

func sliceFilterOptions(full []filter.DimensionOption, offset, limit int) (sliced []filter.DimensionOption) {
	if offset > len(full) {
		return []filter.DimensionOption{}
	}
	end := offset + limit
	if limit == 0 || end > len(full) {
		end = len(full)
	}
	return full[offset:end]
}

// go-mock tailored matcher to compare lists of strings ignoring order
type itemsEq struct{ expected []string }

// ItemsEq checks if 2 slices contain the same items in any order
func ItemsEq(expected []string) gomock.Matcher {
	return &itemsEq{expected}
}

func (i *itemsEq) Matches(x interface{}) bool {
	if len(x.([]string)) != len(i.expected) {
		return false
	}
	mExpected := make(map[string]struct{})
	for _, e := range i.expected {
		mExpected[e] = struct{}{}
	}
	for _, val := range x.([]string) {
		if _, found := mExpected[val]; !found {
			return false
		}
	}
	return true
}

func (i *itemsEq) String() string {
	return fmt.Sprintf("%v (in any order)", i.expected)
}
