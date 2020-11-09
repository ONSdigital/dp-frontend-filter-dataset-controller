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

func TestUseLatest(t *testing.T) {
	ctx := gomock.Any()
	mockServiceAuthToken := ""
	mockDownloadToken := ""
	mockUserAuthToken := ""
	mockCollectionID := ""
	filterID := "current-filter-id"
	mockNewFilterID := "new-filter-id"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("Test UseLatest redirects to a new filter job", t, func() {
		mockEditionLinks := dataset.Links{
			LatestVersion: dataset.Link{
				URL: "",
				ID:  "2",
			},
		}

		mockDimensionOption := []filter.DimensionOption{
			{DimensionOptionsURL: "/123", Option: "monday"},
		}

		// Mock the calls that are expected to be made in the UseLatest method
		mockFilterClient := NewMockFilterClient(mockCtrl)
		mockDatasetClient := NewMockDatasetClient(mockCtrl)
		mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, filterID).Return(filter.Model{Links: filter.Links{Version: filter.Link{HRef: "/v1/datasets/95c4669b-3ae9-4ba7-b690-87e890a1c67c/editions/2016/versions/1"}}}, nil)
		mockFilterClient.EXPECT().GetDimensions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID).Return([]filter.Dimension{{Name: "Day"}}, nil)
		mockDatasetClient.EXPECT().GetEdition(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016").Return(dataset.Edition{Links: mockEditionLinks}, nil)
		mockFilterClient.EXPECT().CreateBlueprint(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "2", []string{}).Return(mockNewFilterID, nil)
		mockFilterClient.EXPECT().AddDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockNewFilterID, "Day").Return(nil)
		mockFilterClient.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Day").Return(mockDimensionOption, nil)
		mockFilterClient.EXPECT().AddDimensionValues(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockNewFilterID, "Day", []string{mockDimensionOption[0].Option}).Return(nil)

		mockRenderer := NewMockRenderer(mockCtrl)
		f := NewFilter(mockRenderer, mockFilterClient, mockDatasetClient, nil, nil, nil, mockServiceAuthToken, "", "/v1", false)

		router := mux.NewRouter()
		router.Path("/filters/{filterID}/use-latest-version").HandlerFunc(f.UseLatest())
		req := httptest.NewRequest("GET", "/filters/current-filter-id/use-latest-version", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		redirectPath := "/filters/new-filter-id/dimensions"

		So(w.Code, ShouldEqual, http.StatusFound)
		So(w.Header()["Location"][0], ShouldEqual, redirectPath)
	})

}
