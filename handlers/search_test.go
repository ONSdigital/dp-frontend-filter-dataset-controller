package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/clients/search"
	gomock "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitSearch(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	filterID := "12345"
	datasetID := "abcde"
	name := "aggregate"
	edition := "2017"
	version := "1"
	query := "Newport"
	expectedHTML := "<html>Search Results</html>"

	Convey("test Search", t, func() {
		Convey("test search can successfully load a page", func() {

			mfc := NewMockFilterClient(mockCtrl)
			mdc := NewMockDatasetClient(mockCtrl)
			msc := NewMockSearchClient(mockCtrl)
			mrc := NewMockRenderer(mockCtrl)

			mfc.EXPECT().GetJobState(filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:22000/datasets/abcde/editions/2017/versions/1",
					},
				},
			}, nil)
			mfc.EXPECT().GetDimensionOptions(filterID, name).Return([]filter.DimensionOption{}, nil)
			mdc.EXPECT().Get(datasetID).Return(dataset.Model{}, nil)
			mdc.EXPECT().GetVersion(datasetID, edition, version).Return(dataset.Version{}, nil)
			mdc.EXPECT().GetOptions(datasetID, edition, version, name).Return(dataset.Options{}, nil)
			msc.EXPECT().Dimension(datasetID, edition, version, name, query).Return(&search.Model{}, nil)
			mrc.EXPECT().Do("dataset-filter/hierarchy", gomock.Any()).Return([]byte(expectedHTML), nil)

			req := httptest.NewRequest("GET", "/filters/12345/dimensions/aggregate/search?q=Newport", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(mrc, mfc, mdc, nil, nil, msc, nil)
			router.Path("/filters/{filterID}/dimensions/{name}/search").Methods("GET").HandlerFunc(f.Search)

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, expectedHTML)

		})

		Convey("test search can returns server error if GetJobState errors", func() {

			mfc := NewMockFilterClient(mockCtrl)
			mdc := NewMockDatasetClient(mockCtrl)
			msc := NewMockSearchClient(mockCtrl)
			mrc := NewMockRenderer(mockCtrl)

			mfc.EXPECT().GetJobState(filterID).Return(filter.Model{}, errors.New("get job state error"))

			req := httptest.NewRequest("GET", "/filters/12345/dimensions/aggregate/search?q=Newport", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(mrc, mfc, mdc, nil, nil, msc, nil)
			router.Path("/filters/{filterID}/dimensions/{name}/search").Methods("GET").HandlerFunc(f.Search)

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("test search can returns server error if GetDimensionOptions errors", func() {

			mfc := NewMockFilterClient(mockCtrl)
			mdc := NewMockDatasetClient(mockCtrl)
			msc := NewMockSearchClient(mockCtrl)
			mrc := NewMockRenderer(mockCtrl)

			mfc.EXPECT().GetJobState(filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:22000/datasets/abcde/editions/2017/versions/1",
					},
				},
			}, nil)
			mfc.EXPECT().GetDimensionOptions(filterID, name).Return([]filter.DimensionOption{}, errors.New("get dimensions options error"))

			req := httptest.NewRequest("GET", "/filters/12345/dimensions/aggregate/search?q=Newport", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(mrc, mfc, mdc, nil, nil, msc, nil)
			router.Path("/filters/{filterID}/dimensions/{name}/search").Methods("GET").HandlerFunc(f.Search)

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)

		})
		Convey("test search can returns server error if Dataset Get errors", func() {

			mfc := NewMockFilterClient(mockCtrl)
			mdc := NewMockDatasetClient(mockCtrl)
			msc := NewMockSearchClient(mockCtrl)
			mrc := NewMockRenderer(mockCtrl)

			mfc.EXPECT().GetJobState(filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:22000/datasets/abcde/editions/2017/versions/1",
					},
				},
			}, nil)
			mfc.EXPECT().GetDimensionOptions(filterID, name).Return([]filter.DimensionOption{}, nil)
			mdc.EXPECT().Get(datasetID).Return(dataset.Model{}, errors.New("dataset get error"))

			req := httptest.NewRequest("GET", "/filters/12345/dimensions/aggregate/search?q=Newport", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(mrc, mfc, mdc, nil, nil, msc, nil)
			router.Path("/filters/{filterID}/dimensions/{name}/search").Methods("GET").HandlerFunc(f.Search)

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)

		})
		Convey("test search can returns server error if GetVersion errors", func() {

			mfc := NewMockFilterClient(mockCtrl)
			mdc := NewMockDatasetClient(mockCtrl)
			msc := NewMockSearchClient(mockCtrl)
			mrc := NewMockRenderer(mockCtrl)

			mfc.EXPECT().GetJobState(filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:22000/datasets/abcde/editions/2017/versions/1",
					},
				},
			}, nil)
			mfc.EXPECT().GetDimensionOptions(filterID, name).Return([]filter.DimensionOption{}, nil)
			mdc.EXPECT().Get(datasetID).Return(dataset.Model{}, nil)
			mdc.EXPECT().GetVersion(datasetID, edition, version).Return(dataset.Version{}, errors.New("get version error"))

			req := httptest.NewRequest("GET", "/filters/12345/dimensions/aggregate/search?q=Newport", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(mrc, mfc, mdc, nil, nil, msc, nil)
			router.Path("/filters/{filterID}/dimensions/{name}/search").Methods("GET").HandlerFunc(f.Search)

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)

		})
		Convey("test search can returns server error if GetOptions errors", func() {

			mfc := NewMockFilterClient(mockCtrl)
			mdc := NewMockDatasetClient(mockCtrl)
			msc := NewMockSearchClient(mockCtrl)
			mrc := NewMockRenderer(mockCtrl)

			mfc.EXPECT().GetJobState(filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:22000/datasets/abcde/editions/2017/versions/1",
					},
				},
			}, nil)
			mfc.EXPECT().GetDimensionOptions(filterID, name).Return([]filter.DimensionOption{}, nil)
			mdc.EXPECT().Get(datasetID).Return(dataset.Model{}, nil)
			mdc.EXPECT().GetVersion(datasetID, edition, version).Return(dataset.Version{}, nil)
			mdc.EXPECT().GetOptions(datasetID, edition, version, name).Return(dataset.Options{}, errors.New("get options error"))

			req := httptest.NewRequest("GET", "/filters/12345/dimensions/aggregate/search?q=Newport", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(mrc, mfc, mdc, nil, nil, msc, nil)
			router.Path("/filters/{filterID}/dimensions/{name}/search").Methods("GET").HandlerFunc(f.Search)

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test search can returns server error if search api call errors", func() {

			mfc := NewMockFilterClient(mockCtrl)
			mdc := NewMockDatasetClient(mockCtrl)
			msc := NewMockSearchClient(mockCtrl)
			mrc := NewMockRenderer(mockCtrl)

			mfc.EXPECT().GetJobState(filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:22000/datasets/abcde/editions/2017/versions/1",
					},
				},
			}, nil)
			mfc.EXPECT().GetDimensionOptions(filterID, name).Return([]filter.DimensionOption{}, nil)
			mdc.EXPECT().Get(datasetID).Return(dataset.Model{}, nil)
			mdc.EXPECT().GetVersion(datasetID, edition, version).Return(dataset.Version{}, nil)
			mdc.EXPECT().GetOptions(datasetID, edition, version, name).Return(dataset.Options{}, nil)
			msc.EXPECT().Dimension(datasetID, edition, version, name, query).Return(&search.Model{}, errors.New("search api error"))

			req := httptest.NewRequest("GET", "/filters/12345/dimensions/aggregate/search?q=Newport", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(mrc, mfc, mdc, nil, nil, msc, nil)
			router.Path("/filters/{filterID}/dimensions/{name}/search").Methods("GET").HandlerFunc(f.Search)

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)

		})
		Convey("test search can returns server error if renderer errors", func() {

			mfc := NewMockFilterClient(mockCtrl)
			mdc := NewMockDatasetClient(mockCtrl)
			msc := NewMockSearchClient(mockCtrl)
			mrc := NewMockRenderer(mockCtrl)

			mfc.EXPECT().GetJobState(filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:22000/datasets/abcde/editions/2017/versions/1",
					},
				},
			}, nil)
			mfc.EXPECT().GetDimensionOptions(filterID, name).Return([]filter.DimensionOption{}, nil)
			mdc.EXPECT().Get(datasetID).Return(dataset.Model{}, nil)
			mdc.EXPECT().GetVersion(datasetID, edition, version).Return(dataset.Version{}, nil)
			mdc.EXPECT().GetOptions(datasetID, edition, version, name).Return(dataset.Options{}, nil)
			msc.EXPECT().Dimension(datasetID, edition, version, name, query).Return(&search.Model{}, nil)
			mrc.EXPECT().Do("dataset-filter/hierarchy", gomock.Any()).Return([]byte(expectedHTML), errors.New("renderer error"))

			req := httptest.NewRequest("GET", "/filters/12345/dimensions/aggregate/search?q=Newport", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(mrc, mfc, mdc, nil, nil, msc, nil)
			router.Path("/filters/{filterID}/dimensions/{name}/search").Methods("GET").HandlerFunc(f.Search)

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)

		})

		Convey("test error returned if version url cannot be parsed", func() {

			mfc := NewMockFilterClient(mockCtrl)
			mdc := NewMockDatasetClient(mockCtrl)
			msc := NewMockSearchClient(mockCtrl)
			mrc := NewMockRenderer(mockCtrl)

			mfc.EXPECT().GetJobState(filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:22000/datasets",
					},
				},
			}, nil)
			mfc.EXPECT().GetDimensionOptions(filterID, name).Return([]filter.DimensionOption{}, nil)

			req := httptest.NewRequest("GET", "/filters/12345/dimensions/aggregate/search?q=Newport", nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(mrc, mfc, mdc, nil, nil, msc, nil)
			router.Path("/filters/{filterID}/dimensions/{name}/search").Methods("GET").HandlerFunc(f.Search)

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)

		})

	})

}