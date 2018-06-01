package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	dataset "github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitFilterOverview(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("test FilterOverview", t, func() {
		Convey("test FilterOverview can successfully load a page", func() {
			mockFilterClient := NewMockFilterClient(mockCtrl)
			mockFilterClient.EXPECT().GetDimensions("12345").Return([]filter.Dimension{filter.Dimension{Name: "Day"}, filter.Dimension{Name: "Goods and Services"}}, nil)
			mockFilterClient.EXPECT().GetDimensionOptions("12345", "Day").Return([]filter.DimensionOption{}, nil)
			mockFilterClient.EXPECT().GetDimensionOptions("12345", "Goods and Services").Return([]filter.DimensionOption{}, nil)
			mockFilterClient.EXPECT().GetJobState("12345").Return(filter.Model{Links: filter.Links{Version: filter.Link{HRef: "/datasets/95c4669b-3ae9-4ba7-b690-87e890a1c67c/editions/2016/versions/1"}}}, nil)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockDatasetClient.EXPECT().GetDimensions("95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1").Return(dataset.Dimensions{Items: []dataset.Dimension{{ID: "geography"}}}, nil)
			mockDatasetClient.EXPECT().GetOptions("95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1", "geography")
			mockDatasetClient.EXPECT().Get("95c4669b-3ae9-4ba7-b690-87e890a1c67c").Return(dataset.Model{Contacts: []dataset.Contact{{Name: "Matt"}}}, nil)
			mockDatasetClient.EXPECT().GetVersion("95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1").Return(dataset.Version{}, nil)
			mockRenderer := NewMockRenderer(mockCtrl)
			mockRenderer.EXPECT().Do("dataset-filter/filter-overview", gomock.Any()).Return([]byte("some-bytes"), nil)

			req := httptest.NewRequest("GET", "/filters/12345/dimensions", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(mockRenderer, mockFilterClient, mockDatasetClient, nil, nil, nil, nil, "")
			router.Path("/filters/{filterID}/dimensions").HandlerFunc(f.FilterOverview)

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, "some-bytes")
		})

		Convey("test sucessful FilterOverviewClearAll", func() {
			mockFilterClient := NewMockFilterClient(mockCtrl)
			mockFilterClient.EXPECT().GetDimensions("12345").Return([]filter.Dimension{filter.Dimension{Name: "Day"}, filter.Dimension{Name: "Goods and Services"}}, nil)
			mockFilterClient.EXPECT().RemoveDimension("12345", "Day")
			mockFilterClient.EXPECT().AddDimension("12345", "Day")
			mockFilterClient.EXPECT().RemoveDimension("12345", "Goods and Services")
			mockFilterClient.EXPECT().AddDimension("12345", "Goods and Services")

			req := httptest.NewRequest("GET", "/filters/12345/dimensions/clear-all", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(nil, mockFilterClient, nil, nil, nil, nil, nil, "")
			router.Path("/filters/{filterID}/dimensions/clear-all").HandlerFunc(f.FilterOverviewClearAll)

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Header().Get("Location"), ShouldEqual, "/filters/12345/dimensions")
		})
	})
}
