package handlers

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
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

	Convey("Given a filter initialised with a mocked DatasetClient ", t, func() {
		mdc := NewMockDatasetClient(mockCtrl)

		Convey("a set of existing filter dimension options is correctly mapped to lables returned by dataset API GetOptions", func() {
			batchSize := 100
			f := NewFilter(nil, nil, mdc, nil, nil, nil, mockServiceAuthToken, "", "/v1", false, batchSize)
			filterOptions := filter.DimensionOptions{
				Items: []filter.DimensionOption{
					{Option: "op1"},
					{Option: "op2"},
				},
			}
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name, 0, batchSize).Return(datasetOptions(0, batchSize), nil)
			idLabelMap, err := f.getIDNameLookupFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name, filterOptions)
			So(err, ShouldBeNil)
			So(idLabelMap, ShouldResemble, map[string]string{
				"op1": "This is option 1",
				"op2": "This is option 2",
			})
		})

		Convey("a set of existing filter dimension options is correctly mapped to lables returned by dataset API GetOptions in multiple batches", func() {
			batchSize := 3
			f := NewFilter(nil, nil, mdc, nil, nil, nil, mockServiceAuthToken, "", "/v1", false, batchSize)
			filterOptions := filter.DimensionOptions{
				Items: []filter.DimensionOption{
					{Option: "op1"}, // belongs to first batch
					{Option: "op4"}, // belongs to second batch
					{Option: "op5"}, // belongs to second batch
				},
			}
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name, 0, batchSize).Return(datasetOptions(0, batchSize), nil)
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name, batchSize, batchSize).Return(datasetOptions(batchSize, batchSize), nil)
			idLabelMap, err := f.getIDNameLookupFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name, filterOptions)
			So(err, ShouldBeNil)
			So(idLabelMap, ShouldResemble, map[string]string{
				"op1": "This is option 1",
				"op4": "This is option 4",
				"op5": "This is option 5",
			})
		})

		Convey("a set of existing filter dimension options is correctly mapped to lables returned by dataset API GetOptions and no further batches are executed if all items have been found", func() {
			batchSize := 3
			f := NewFilter(nil, nil, mdc, nil, nil, nil, mockServiceAuthToken, "", "/v1", false, batchSize)
			filterOptions := filter.DimensionOptions{
				Items: []filter.DimensionOption{
					{Option: "op1"}, // belongs to first batch
					{Option: "op2"}, // belongs to first batch
				},
			}
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name, 0, batchSize).Return(datasetOptions(0, batchSize), nil)
			idLabelMap, err := f.getIDNameLookupFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name, filterOptions)
			So(err, ShouldBeNil)
			So(idLabelMap, ShouldResemble, map[string]string{
				"op1": "This is option 1",
				"op2": "This is option 2",
			})
		})

		Convey("a set of filter dimension options containing inexistent options returns the expected error, but the existing dimensions are correctly mapped", func() {
			batchSize := 100
			f := NewFilter(nil, nil, mdc, nil, nil, nil, mockServiceAuthToken, "", "/v1", false, batchSize)
			filterOptions := filter.DimensionOptions{
				Items: []filter.DimensionOption{
					{Option: "op1"},
					{Option: "inexistent"},
				},
			}
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name, 0, batchSize).Return(datasetOptions(0, batchSize), nil)
			expectedErr := errors.New("could not find all required filter options in dataset API")
			idLabelMap, err := f.getIDNameLookupFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name, filterOptions)
			So(err, ShouldResemble, expectedErr)
			So(idLabelMap, ShouldResemble, map[string]string{
				"op1":        "This is option 1",
				"inexistent": "",
			})
		})

		Convey("if dataset API GetOptions returns an error, the same error is returned and the next batch is not requested", func() {
			batchSize := 2
			f := NewFilter(nil, nil, mdc, nil, nil, nil, mockServiceAuthToken, "", "/v1", false, batchSize)
			filterOptions := filter.DimensionOptions{
				Items: []filter.DimensionOption{{Option: "op1"}},
			}
			expectedErr := errors.New("error getting options from Dataset API")
			mdc.EXPECT().GetOptions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name, 0, batchSize).Return(dataset.Options{}, expectedErr)
			idLabelMap, err := f.getIDNameLookupFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name, filterOptions)
			So(err, ShouldResemble, expectedErr)
			So(idLabelMap, ShouldResemble, map[string]string{"op1": ""})
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

	Convey("Given a filter initialised with a mocked FilterClient ", t, func() {
		mfc := NewMockFilterClient(mockCtrl)

		Convey("given that the number of filter options is smaller than a batch, then all options are returned after a single batch getDimensionOptions call", func() {
			batchSize := 100
			f := NewFilter(nil, mfc, nil, nil, nil, nil, mockServiceAuthToken, "", "/v1", false, batchSize)
			mfc.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name, 0, batchSize).Return(filterOptions(0, batchSize), nil)
			opts, err := f.GetDimensionOptionsFromFilterAPI(ctx, mockUserAuthToken, mockCollectionID, filterID, name)
			So(err, ShouldBeNil)
			So(opts, ShouldResemble, filterOptions(0, 0))
		})

		Convey("given that the number of filter options is greater than a batch, then all options are returned after multiple batch getDimensionOptions calls", func() {
			batchSize := 3
			f := NewFilter(nil, mfc, nil, nil, nil, nil, mockServiceAuthToken, "", "/v1", false, batchSize)
			mfc.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name, 0, batchSize).Return(filterOptions(0, batchSize), nil)
			mfc.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name, batchSize, batchSize).Return(filterOptions(batchSize, batchSize), nil)
			opts, err := f.GetDimensionOptionsFromFilterAPI(ctx, mockUserAuthToken, mockCollectionID, filterID, name)
			So(err, ShouldBeNil)
			So(opts, ShouldResemble, filterOptions(0, 0))
		})

		Convey("if filter API GetDimensionOptions returns an error, the same error is returned and the next batch is not requested", func() {
			batchSize := 2
			f := NewFilter(nil, mfc, nil, nil, nil, nil, mockServiceAuthToken, "", "/v1", false, batchSize)
			expectedErr := errors.New("error getting options from Filter API")
			mfc.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name, 0, batchSize).Return(filter.DimensionOptions{}, expectedErr)
			opts, err := f.GetDimensionOptionsFromFilterAPI(ctx, mockUserAuthToken, mockCollectionID, filterID, name)
			So(err, ShouldResemble, expectedErr)
			So(opts, ShouldResemble, filter.DimensionOptions{})
		})

	})
}

// datasetOptions returns a mocked dataset.Options struct according to the provided offset and limit
func datasetOptions(offset, limit int) dataset.Options {
	allItems := []dataset.Option{
		{
			Label:  "This is option 1",
			Option: "op1",
		},
		{
			Label:  "This is option 2",
			Option: "op2",
		},
		{
			Label:  "This is option 3",
			Option: "op3",
		},
		{
			Label:  "This is option 4",
			Option: "op4",
		},
		{
			Label:  "This is option 5",
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
