package handlers

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUpdateAge(t *testing.T) {
	t.Parallel()

	const mockUserAuthToken = "Foo"
	const mockServiceAuthToken = ""
	const mockCollectionID = "Bar"
	const mockFilterID = ""
	const batchSize = 100
	const maxWorkers = 25

	cfg := &config.Config{
		SearchAPIAuthToken:   mockServiceAuthToken,
		DownloadServiceURL:   "",
		EnableDatasetPreview: false,
		BatchSizeLimit:       batchSize,
		BatchMaxWorkers:      maxWorkers,
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	callAgeUpdate := func(formData string, mockFilterClient *MockFilterClient, mockDatasetClient *MockDatasetClient) *httptest.ResponseRecorder {
		target := fmt.Sprintf("/filters/%s/dimensions/time/update", mockFilterID)
		reader := strings.NewReader(formData)
		req := httptest.NewRequest("POST", target, reader)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)
		req.Header.Add(dprequest.CollectionIDHeaderKey, mockCollectionID)
		w := httptest.NewRecorder()
		f := NewFilter(nil, mockFilterClient, mockDatasetClient, nil, nil, nil, "/v1", cfg)
		f.UpdateAge().ServeHTTP(w, req)
		return w
	}

	Convey("Given that a user selects age options from the list, then the redirect is successful and the expected calls are made to the filter API", t, func() {
		options := []string{"30", "28", "90+"}
		mockClient := NewMockFilterClient(mockCtrl)
		mockClient.EXPECT().RemoveDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockFilterID, "age").Return(nil)
		mockClient.EXPECT().AddDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockFilterID, "age").Return(nil)
		mockClient.EXPECT().SetDimensionValues(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockFilterID, "age", ItemsEq(options)).Return(nil)
		formData := "all-ages-option=total&youngest-age=0&oldest-age=90%2B&youngest=&oldest=&age-selection=list&28=28&30=30&90%2B=90%2B&save-and-return=Save+and+return"
		w := callAgeUpdate(formData, mockClient, nil)
		So(w.Code, ShouldEqual, 302)
	})

	Convey("Given that a user selects all age options, then the redirect is successful and the expected calls are made to the filter API", t, func() {
		option := "total"
		mockClient := NewMockFilterClient(mockCtrl)
		mockClient.EXPECT().RemoveDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockFilterID, "age").Return(nil)
		mockClient.EXPECT().AddDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockFilterID, "age").Return(nil)
		mockClient.EXPECT().AddDimensionValue(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockFilterID, "age", option).Return(nil)
		formData := "age-selection=all&all-ages-option=total&youngest-age=0&oldest-age=90%2B&youngest=&oldest=&28=28&30=30&90%2B=90%2B&save-and-return=Save+and+return"
		w := callAgeUpdate(formData, mockClient, nil)
		So(w.Code, ShouldEqual, 302)
	})

	Convey("Given that a user selects a range of age options, then teh redirect is successful and the expected calls are made to the filter API", t, func() {
		expectedFilterModel := filter.Model{
			Links: filter.Links{
				Version: filter.Link{
					HRef: "http://localhost:23200/v1/datasets/mid-year-pop-est/editions/mid-2019-april-2020-geography/versions/1",
				},
			},
		}
		datasetOptions := dataset.Options{
			Items: []dataset.Option{
				{Label: "18", Option: "18"},
				{Label: "19", Option: "19"},
				{Label: "20", Option: "20"},
				{Label: "21", Option: "21"},
				{Label: "22", Option: "22"},
				{Label: "23", Option: "23"},
				{Label: "24", Option: "24"},
			},
		}
		filterOptions := []string{"18", "19", "20", "21", "22", "23", "24"}
		mockFilterClient := NewMockFilterClient(mockCtrl)
		mockDatasetClient := NewMockDatasetClient(mockCtrl)
		mockFilterClient.EXPECT().RemoveDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockFilterID, "age").Return(nil)
		mockFilterClient.EXPECT().AddDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockFilterID, "age").Return(nil)
		mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, mockServiceAuthToken, "", mockCollectionID, mockFilterID).Return(expectedFilterModel, nil)
		mockDatasetClient.EXPECT().GetOptionsInBatches(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "mid-year-pop-est", "mid-2019-april-2020-geography", "1", "age",
			batchSize, maxWorkers).Return(datasetOptions, nil)
		mockFilterClient.EXPECT().SetDimensionValues(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockFilterID, "age", filterOptions).Return(nil)
		formData := "all-ages-option=total&age-selection=range&youngest-age=0&oldest-age=90%2B&youngest=18&oldest=24&save-and-return=Save+and+return"
		w := callAgeUpdate(formData, mockFilterClient, mockDatasetClient)
		So(w.Code, ShouldEqual, 302)
	})
}
