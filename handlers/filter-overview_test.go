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
			mockCodeListClient := NewMockCodelistClient(mockCtrl)
			mockCodeListClient.EXPECT().GetIDNameMap("64d384f1-ea3b-445c-8fb8-aa453f96e58a").Return(map[string]string{
				"1234567": "Monday",
			}, nil)
			mockCodeListClient.EXPECT().GetIDNameMap("e44de4c4-d39e-4e2f-942b-3ca10584d078").Return(map[string]string{
				"678910": "Travel",
			}, nil)
			mockFilterClient.EXPECT().GetDimensionOptions("12345", "Day").Return([]filter.DimensionOption{}, nil)
			mockFilterClient.EXPECT().GetDimensionOptions("12345", "Goods and Services").Return([]filter.DimensionOption{}, nil)
			mockFilterClient.EXPECT().GetJobState("12345").Return(filter.Model{DatasetFilterID: "/datasets/12345/editions/2016/versions/1"}, nil)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockDatasetClient.EXPECT().Get("12345").Return(dataset.Model{}, nil)
			mockDatasetClient.EXPECT().GetVersion("12345", "2016", "1").Return(dataset.Version{}, nil)
			mockRenderer := NewMockRenderer(mockCtrl)
			mockRenderer.EXPECT().Do("dataset-filter/filter-overview", gomock.Any()).Return([]byte("some-bytes"), nil)

			req := httptest.NewRequest("GET", "/filters/12345/dimensions", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(mockRenderer, mockFilterClient, mockDatasetClient, mockCodeListClient, nil, nil)
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
			f := NewFilter(nil, mockFilterClient, nil, nil, nil, nil)
			router.Path("/filters/{filterID}/dimensions/clear-all").HandlerFunc(f.FilterOverviewClearAll)

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Header().Get("Location"), ShouldEqual, "/filters/12345/dimensions")
		})
	})
}
