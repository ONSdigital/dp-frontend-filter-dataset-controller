package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitFilterOverview(t *testing.T) {
	ctx := gomock.Any()
	mockServiceAuthToken := ""
	mockDownloadToken := ""
	mockUserAuthToken := ""
	mockCollectionID := ""
	filterID := "12345"
	batchSize := 100
	maxWorkers := 25
	maxDatasetOptions := 3

	cfg := &config.Config{
		SearchAPIAuthToken:   mockServiceAuthToken,
		DownloadServiceURL:   "",
		BatchSizeLimit:       batchSize,
		MaxDatasetOptions:    maxDatasetOptions,
		EnableDatasetPreview: false,
		BatchMaxWorkers:      maxWorkers,
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("test FilterOverview", t, func() {

		filterGeographyOptions := filter.DimensionOptions{
			Items: []filter.DimensionOption{
				{
					Option: "geoUK",
				},
			},
			Count:      1,
			TotalCount: 1,
			Offset:     0,
			Limit:      0,
		}

		Convey("test FilterOverview can successfully load a page, with expected calls", func() {
			// expected calls to filter api: get dimension and get options for each dimension. Get job state.
			mockFilterClient := NewMockFilterClient(mockCtrl)
			mockFilterClient.EXPECT().GetDimensions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, nil).
				Return(
					filter.Dimensions{
						Items: []filter.Dimension{{Name: "geography"}, {Name: "Day"}, {Name: "Goods and Services"}},
					}, testETag(0), nil)
			mockFilterClient.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "geography", batchSize, maxWorkers).Return(filterGeographyOptions, testETag(0), nil)
			mockFilterClient.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Day", batchSize, maxWorkers).Return(filter.DimensionOptions{}, testETag(0), nil)
			mockFilterClient.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Goods and Services", batchSize, maxWorkers).Return(filter.DimensionOptions{}, testETag(0), nil)
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, filterID).Return(filter.Model{Links: filter.Links{Version: filter.Link{HRef: "/v1/datasets/95c4669b-3ae9-4ba7-b690-87e890a1c67c/editions/2016/versions/1"}}}, testETag(0), nil)

			// expected calls to dataset api: get options only for options that were found in filter api. Get, GetVersion and GetEdition
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockDatasetClient.EXPECT().GetVersionDimensions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1").Return(dataset.VersionDimensions{
				Items: []dataset.VersionDimension{{Name: "geography"}, {Name: "Day"}, {Name: "Goods and services"}, {Name: "unused"}}}, nil)
			mockDatasetClient.EXPECT().GetOptionsBatchProcess(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1", "geography",
				&[]string{"geoUK"}, gomock.Any(), maxDatasetOptions, maxWorkers).Return(nil)
			mockDatasetClient.EXPECT().Get(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Matt"}}}, nil)
			mockDatasetClient.EXPECT().GetVersion(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1").Return(dataset.Version{}, nil)
			mockDatasetClient.EXPECT().GetEdition(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016").Return(dataset.Edition{}, nil)

			mockRenderer := NewMockRenderer(mockCtrl)
			mockRenderer.EXPECT().Do("dataset-filter/filter-overview", gomock.Any()).Return([]byte("some-bytes"), nil)

			req := httptest.NewRequest("GET", "/filters/12345/dimensions", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(mockRenderer, mockFilterClient, mockDatasetClient, nil, nil, "/v1", cfg)
			router.Path("/filters/{filterID}/dimensions").HandlerFunc(f.FilterOverview())

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, "some-bytes")
		})

		Convey("test successful FilterOverviewClearAll", func() {
			mockFilterClient := NewMockFilterClient(mockCtrl)
			mockFilterClient.EXPECT().GetDimensions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, nil).
				Return(
					filter.Dimensions{
						Items: []filter.Dimension{{Name: "geography"}, {Name: "Day"}, {Name: "Goods and Services"}},
					}, testETag(0), nil)
			mockFilterClient.EXPECT().RemoveDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "geography", testETag(0)).Return(testETag(1), nil)
			mockFilterClient.EXPECT().AddDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "geography", testETag(1)).Return(testETag(2), nil)
			mockFilterClient.EXPECT().RemoveDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Day", testETag(2)).Return(testETag(3), nil)
			mockFilterClient.EXPECT().AddDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Day", testETag(3)).Return(testETag(4), nil)
			mockFilterClient.EXPECT().RemoveDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Goods and Services", testETag(4)).Return(testETag(5), nil)
			mockFilterClient.EXPECT().AddDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Goods and Services", testETag(5)).Return(testETag(6), nil)

			req := httptest.NewRequest("GET", "/filters/12345/dimensions/clear-all", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(nil, mockFilterClient, nil, nil, nil, "/v1", cfg)
			router.Path("/filters/{filterID}/dimensions/clear-all").HandlerFunc(f.FilterOverviewClearAll())

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Header().Get("Location"), ShouldEqual, "/filters/12345/dimensions")
		})
	})
}
