package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUseLatest(t *testing.T) {
	ctx := gomock.Any()
	mockServiceAuthToken := ""
	mockDownloadToken := ""
	mockUserAuthToken := ""
	mockCollectionID := ""
	filterID := "current-filter-id"
	mockNewFilterID := "new-filter-id"
	batchSize := 100
	maxWorkers := 25

	cfg := &config.Config{
		SearchAPIAuthToken:   mockServiceAuthToken,
		DownloadServiceURL:   "",
		BatchSizeLimit:       batchSize,
		BatchMaxWorkers:      maxWorkers,
		EnableDatasetPreview: false,
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("Test UseLatest redirects to a new filter job", t, func() {
		mockEditionLinks := dataset.Links{
			LatestVersion: dataset.Link{
				URL: "",
				ID:  "2",
			},
		}

		// Mock the calls that are expected to be made in the UseLatest method
		mockFilterClient := NewMockFilterClient(mockCtrl)
		mockDatasetClient := NewMockDatasetClient(mockCtrl)
		mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, filterID).Return(
			filter.Model{Links: filter.Links{Version: filter.Link{HRef: "/v1/datasets/95c4669b-3ae9-4ba7-b690-87e890a1c67c/editions/2016/versions/1"}}}, testETag(0), nil)
		mockFilterClient.EXPECT().GetDimensions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, nil).
			Return(
				filter.Dimensions{
					Items: []filter.Dimension{{Name: "Day"}},
				}, testETag(0), nil)
		mockDatasetClient.EXPECT().GetEdition(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016").Return(dataset.Edition{Links: mockEditionLinks}, nil)
		mockFilterClient.EXPECT().CreateBlueprint(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "2", []string{}).Return(mockNewFilterID, testETag(1), nil)
		mockFilterClient.EXPECT().AddDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockNewFilterID, "Day", testETag(1)).Return(testETag(2), nil)
		mockFilterClient.EXPECT().GetDimensionOptionsBatchProcess(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Day", gomock.Any(), batchSize, maxWorkers, true).Return(testETag(0), nil)

		mockRenderer := NewMockRenderer(mockCtrl)
		f := NewFilter(mockRenderer, mockFilterClient, mockDatasetClient, nil, nil, nil, "/v1", cfg)

		router := mux.NewRouter()
		router.Path("/filters/{filterID}/use-latest-version").HandlerFunc(f.UseLatest())
		req := httptest.NewRequest("GET", "/filters/current-filter-id/use-latest-version", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		redirectPath := "/filters/new-filter-id/dimensions"

		So(w.Code, ShouldEqual, http.StatusFound)
		So(w.Header()["Location"][0], ShouldEqual, redirectPath)
	})

	Convey("The batch processor function calls filter API patch endpoint with the expected options", t, func() {

		// mocked dimension Options batch for testing (similar to a real paginated response from filter API)
		mockDimensionOptionsBatch := filter.DimensionOptions{
			Items: []filter.DimensionOption{
				{DimensionOptionsURL: "/123", Option: "monday"},
			},
			Count:      1,
			TotalCount: 10,
			Offset:     3,
			Limit:      1,
		}

		// Mock the calls that are expected to be made by the batch processor method
		mockFilterClient := NewMockFilterClient(mockCtrl)
		mockFilterClient.EXPECT().PatchDimensionValues(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockNewFilterID, "Day", []string{mockDimensionOptionsBatch.Items[0].Option}, []string{}, batchSize, testETag(0)).Return(testETag(1), nil)
		f := NewFilter(nil, mockFilterClient, nil, nil, nil, nil, "/v1", cfg)

		batchProcessor := f.batchAddOptions(context.Background(), mockUserAuthToken, mockCollectionID, mockNewFilterID, "Day", testETag(0))
		forceAbort, err := batchProcessor(mockDimensionOptionsBatch, testETag(0))
		So(err, ShouldBeNil)
		So(forceAbort, ShouldBeFalse)
	})
}
