package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
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

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("test FilterOverview", t, func() {

		datasetGeographyOptions := dataset.Options{
			Items: []dataset.Option{
				{
					DimensionID: "geography",
					Label:       "UK",
				},
			},
			Count:      1,
			TotalCount: 1,
			Offset:     0,
			Limit:      batchSize,
		}

		Convey("test FilterOverview can successfully load a page", func() {
			mockFilterClient := NewMockFilterClient(mockCtrl)
			mockFilterClient.EXPECT().GetDimensions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID).Return([]filter.Dimension{{Name: "geography"}, {Name: "Day"}, {Name: "Goods and Services"}}, nil)
			mockFilterClient.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Day", 0, 0).Return(filter.DimensionOptions{}, nil)
			mockFilterClient.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Goods and Services", 0, 0).Return(filter.DimensionOptions{}, nil)
			mockFilterClient.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "geography", 0, 0).Return(filter.DimensionOptions{}, nil)
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, filterID).Return(filter.Model{Links: filter.Links{Version: filter.Link{HRef: "/v1/datasets/95c4669b-3ae9-4ba7-b690-87e890a1c67c/editions/2016/versions/1"}}}, nil)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockDatasetClient.EXPECT().GetVersionDimensions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1").Return(dataset.VersionDimensions{
				Items: []dataset.VersionDimension{{Name: "geography"}, {Name: "Day"}, {Name: "Goods and services"}, {Name: "unused"}}}, nil)
			mockDatasetClient.EXPECT().GetOptions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1", "geography", 0, batchSize).Return(datasetGeographyOptions, nil)
			mockDatasetClient.EXPECT().GetOptions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1", "Day", 0, batchSize).Return(dataset.Options{}, nil)
			mockDatasetClient.EXPECT().GetOptions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1", "Goods and Services", 0, batchSize).Return(dataset.Options{}, nil)
			mockDatasetClient.EXPECT().Get(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Matt"}}}, nil)
			mockDatasetClient.EXPECT().GetVersion(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1").Return(dataset.Version{}, nil)
			mockDatasetClient.EXPECT().GetEdition(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016").Return(dataset.Edition{}, nil)
			mockRenderer := NewMockRenderer(mockCtrl)
			mockRenderer.EXPECT().Do("dataset-filter/filter-overview", gomock.Any()).Return([]byte("some-bytes"), nil)

			req := httptest.NewRequest("GET", "/filters/12345/dimensions", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(mockRenderer, mockFilterClient, mockDatasetClient, nil, nil, nil, mockServiceAuthToken, "", "/v1", false, batchSize)
			router.Path("/filters/{filterID}/dimensions").HandlerFunc(f.FilterOverview())

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, "some-bytes")
		})

		Convey("test successful FilterOverviewClearAll", func() {
			mockFilterClient := NewMockFilterClient(mockCtrl)
			mockFilterClient.EXPECT().GetDimensions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID).Return([]filter.Dimension{{Name: "geography"}, {Name: "Day"}, {Name: "Goods and Services"}}, nil)
			mockFilterClient.EXPECT().RemoveDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Day")
			mockFilterClient.EXPECT().AddDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Day")
			mockFilterClient.EXPECT().RemoveDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Goods and Services")
			mockFilterClient.EXPECT().AddDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Goods and Services")
			mockFilterClient.EXPECT().RemoveDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "geography")
			mockFilterClient.EXPECT().AddDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "geography")

			req := httptest.NewRequest("GET", "/filters/12345/dimensions/clear-all", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(nil, mockFilterClient, nil, nil, nil, nil, mockServiceAuthToken, "", "/v1", false, batchSize)
			router.Path("/filters/{filterID}/dimensions/clear-all").HandlerFunc(f.FilterOverviewClearAll())

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Header().Get("Location"), ShouldEqual, "/filters/12345/dimensions")
		})
	})
}
