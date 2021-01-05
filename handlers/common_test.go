package handlers

import (
	"context"
	"errors"
	"fmt"
	"testing"

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

	maxDatasetOptions := 100
	maxWorkers := 25

	cfg := &config.Config{
		SearchAPIAuthToken:   mockServiceAuthToken,
		DownloadServiceURL:   "",
		EnableDatasetPreview: false,
		BatchMaxWorkers:      maxWorkers,
		MaxDatasetOptions:    maxDatasetOptions,
	}

	Convey("Given a filter initialised with a mocked DatasetClient", t, func() {
		mdc := NewMockDatasetClient(mockCtrl)
		f := NewFilter(nil, nil, mdc, nil, nil, nil, "/v1", cfg)

		Convey("a set of existing filter dimension options is correctly requested in batches to dataset API by their option IDs", func() {
			filterOptions := filter.DimensionOptions{
				Items: []filter.DimensionOption{{Option: "op1"}, {Option: "op2"}},
			}
			mdc.EXPECT().GetOptionsBatchProcess(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				&[]string{"op1", "op2"}, gomock.Any(), maxDatasetOptions, maxWorkers).Return(nil)

			_, err := f.getIDNameLookupFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name, filterOptions)
			So(err, ShouldBeNil)
		})

		Convey("an empty set of filter dimension options does not call dataset API is not called", func() {
			filterOptions := filter.DimensionOptions{}
			_, err := f.getIDNameLookupFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name, filterOptions)
			So(err, ShouldBeNil)
		})

		Convey("if dataset API GetOptions returns an error, the same error is returned and the next batch is not requested", func() {
			filterOptions := filter.DimensionOptions{
				Items: []filter.DimensionOption{{Option: "op1"}, {Option: "op2"}},
			}
			expectedErr := errors.New("error getting options from Dataset API")
			mdc.EXPECT().GetOptionsBatchProcess(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				&[]string{"op1", "op2"}, gomock.Any(), maxDatasetOptions, maxWorkers).Return(expectedErr)

			_, err := f.getIDNameLookupFromDatasetAPI(ctx, mockUserAuthToken, mockCollectionID, datasetID, edition, version, name, filterOptions)
			So(err, ShouldResemble, expectedErr)
		})
	})
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
