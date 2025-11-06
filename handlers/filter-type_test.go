package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	core "github.com/ONSdigital/dis-design-system-go/model"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/helpers"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	mockSearchAPIAuthToken = "testServiceAuthToken"

	filterID          = "12345"
	dimensionName     = "myDimension"
	batchSize         = 100
	maxWorkers        = 25
	maxDatasetOptions = 10

	mockServiceAuthToken = ""
	mockDownloadToken    = ""
	mockUserAuthToken    = ""
	mockCollectionID     = ""
)

func SetUpMockFilterFlexServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Mock Census Response"))
		if err != nil {
			panic(err)
		}
	}))
}

func SetUpGeographyOptions() filter.DimensionOptions {
	return filter.DimensionOptions{
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
}

func TestFilterType(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := gomock.Any()

	mockFilterFlexServer := SetUpMockFilterFlexServer()

	defer mockFilterFlexServer.Close()

	cfg := &config.Config{
		SearchAPIAuthToken:          mockSearchAPIAuthToken,
		DownloadServiceURL:          "",
		BatchSizeLimit:              batchSize,
		BatchMaxWorkers:             maxWorkers,
		MaxDatasetOptions:           maxDatasetOptions,
		EnableDatasetPreview:        false,
		FilterFlexDatasetServiceURL: mockFilterFlexServer.URL,
	}

	Convey("Given a set of CMD mocked clients and models", t, func() {
		filterGeographyOptions := SetUpGeographyOptions()

		mockFilterClient := NewMockFilterClient(mockCtrl)
		mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
		mockDatasetClient := NewMockDatasetClient(mockCtrl)
		mockRend := NewMockRenderClient(mockCtrl)
		mockZebedeeClient := NewMockZebedeeClient(mockCtrl)

		mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, filterID).Return(filter.Model{Links: filter.Links{Version: filter.Link{HRef: "/v1/datasets/95c4669b-3ae9-4ba7-b690-87e890a1c67c/editions/2016/versions/1"}}}, testETag(0), nil)
		mockDatasetClient.EXPECT().Get(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Matt"}}}, nil)
		mockFilterClient.EXPECT().GetDimensions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, nil).
			Return(
				filter.Dimensions{
					Items: []filter.Dimension{{Name: "geography"}, {Name: "Day"}, {Name: "Goods and Services"}},
				}, testETag(0), nil)
		mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, filterID).Return(filter.Model{Links: filter.Links{Version: filter.Link{HRef: "/v1/datasets/95c4669b-3ae9-4ba7-b690-87e890a1c67c/editions/2016/versions/1"}}}, testETag(0), nil)
		mockDatasetClient.EXPECT().GetVersionDimensions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1").Return(dataset.VersionDimensions{
			Items: []dataset.VersionDimension{{Name: "geography"}, {Name: "Day"}, {Name: "Goods and services"}, {Name: "unused"}}}, nil)
		mockFilterClient.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "geography", batchSize, maxWorkers).Return(filterGeographyOptions, testETag(0), nil)
		mockFilterClient.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Day", batchSize, maxWorkers).Return(filter.DimensionOptions{}, testETag(0), nil)
		mockFilterClient.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Goods and Services", batchSize, maxWorkers).Return(filter.DimensionOptions{}, testETag(0), nil)
		mockDatasetClient.EXPECT().GetOptionsBatchProcess(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1", "geography",
			&[]string{"geoUK"}, gomock.Any(), maxDatasetOptions, maxWorkers).Return(nil)
		mockDatasetClient.EXPECT().Get(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Matt"}}}, nil)
		mockDatasetClient.EXPECT().GetEdition(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016").Return(dataset.Edition{}, nil)

		mockZebedeeClient.EXPECT().GetHomepageContent(ctx, mockUserAuthToken, mockCollectionID, "en", "/")

		mockRend.EXPECT().NewBasePageModel().Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
		mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "filter-overview")

		Convey("When a CMD filter page is requested", func() {
			callFilterDimensionPage := func(url2 string) *httptest.ResponseRecorder {
				req, err := http.NewRequest(http.MethodGet, url2, http.NoBody)
				So(err, ShouldBeNil)
				req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)

				u, err := url.Parse(cfg.FilterFlexDatasetServiceURL)
				So(err, ShouldBeNil)

				filterFlexHandler := helpers.CreateReverseProxy("flex", u)
				router := mux.NewRouter()
				w := httptest.NewRecorder()
				f := NewFilter(mockRend, mockFilterClient, mockDatasetClient, mockHierarchyClient, nil, mockZebedeeClient, "/v1", cfg)
				router.Path("/filters/{filterID}/dimensions/{name}").HandlerFunc(f.FilterType(mockDatasetClient, f.FilterOverview(), filterFlexHandler))
				router.ServeHTTP(w, req)
				return w
			}

			w := callFilterDimensionPage(fmt.Sprintf("/filters/%s/dimensions/%s", filterID, dimensionName))
			So(w.Code, ShouldEqual, http.StatusOK)

			b, err := io.ReadAll(w.Body)
			So(err, ShouldBeNil)
			So(string(b), ShouldNotEqual, "Mock Census Response")
		})
	})

	Convey("Given a set of Cantabular mocked clients and models", t, func() {
		mockFilterClient := NewMockFilterClient(mockCtrl)
		mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
		mockDatasetClient := NewMockDatasetClient(mockCtrl)
		mockRend := NewMockRenderClient(mockCtrl)
		mockZebedeeClient := NewMockZebedeeClient(mockCtrl)

		mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, filterID).Return(filter.Model{Links: filter.Links{Version: filter.Link{HRef: "/v1/datasets/RM1111/editions/2021/versions/1"}}}, testETag(0), nil)
		mockDatasetClient.EXPECT().Get(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Matt"}}, Type: "cantabular_flexible_table"}, nil)

		Convey("When a cantabular filter page is requested", func() {
			callFilterDimensionPage := func(url2 string) *httptest.ResponseRecorder {
				req, err := http.NewRequest(http.MethodGet, url2, http.NoBody)
				So(err, ShouldBeNil)
				req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)
				u, err := url.Parse(cfg.FilterFlexDatasetServiceURL)
				So(err, ShouldBeNil)
				filterFlexHandler := helpers.CreateReverseProxy("flex", u)
				router := mux.NewRouter()
				w := httptest.NewRecorder()
				f := NewFilter(mockRend, mockFilterClient, mockDatasetClient, mockHierarchyClient, nil, mockZebedeeClient, "/v1", cfg)
				router.Path("/filters/{filterID}/dimensions/{name}").HandlerFunc(f.FilterType(mockDatasetClient, f.FilterOverview(), filterFlexHandler))
				router.ServeHTTP(w, req)
				return w
			}

			w := callFilterDimensionPage(fmt.Sprintf("/filters/%s/dimensions/%s", filterID, dimensionName))

			b, err := io.ReadAll(w.Body)
			So(err, ShouldBeNil)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(string(b), ShouldEqual, "Mock Census Response")
		})
	})
}

func TestFilterTypeErrors(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := gomock.Any()

	mockFilterFlexServer := SetUpMockFilterFlexServer()

	defer mockFilterFlexServer.Close()

	cfg := &config.Config{
		SearchAPIAuthToken:          mockSearchAPIAuthToken,
		DownloadServiceURL:          "",
		BatchSizeLimit:              batchSize,
		BatchMaxWorkers:             maxWorkers,
		MaxDatasetOptions:           maxDatasetOptions,
		EnableDatasetPreview:        false,
		FilterFlexDatasetServiceURL: mockFilterFlexServer.URL,
	}

	Convey("Given a set of CMD mocked clients and models", t, func() {
		filterGeographyOptions := SetUpGeographyOptions()

		mockFilterClient := NewMockFilterClient(mockCtrl)
		mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
		mockDatasetClient := NewMockDatasetClient(mockCtrl)
		mockRend := NewMockRenderClient(mockCtrl)
		mockZebedeeClient := NewMockZebedeeClient(mockCtrl)

		Convey("When a filter page is requested with no filter id", func() {
			callFilterDimensionPage := func(url2 string) *httptest.ResponseRecorder {
				req, err := http.NewRequest(http.MethodGet, url2, http.NoBody)
				So(err, ShouldBeNil)
				req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)

				u, err := url.Parse(cfg.FilterFlexDatasetServiceURL)
				So(err, ShouldBeNil)

				filterFlexHandler := helpers.CreateReverseProxy("flex", u)
				router := mux.NewRouter()
				w := httptest.NewRecorder()
				f := NewFilter(mockRend, mockFilterClient, mockDatasetClient, mockHierarchyClient, nil, mockZebedeeClient, "/v1", cfg)
				router.Path("/filters").HandlerFunc(f.FilterType(mockDatasetClient, f.FilterOverview(), filterFlexHandler))
				router.ServeHTTP(w, req)
				return w
			}

			w := callFilterDimensionPage("/filters")
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When the retrieval of the filter job fails", func() {
			errGetJobState := errors.New("error getting job state")
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, filterID).Return(filter.Model{}, "", errGetJobState)
			mockFilterClient.EXPECT().GetDimensions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, nil).
				Return(
					filter.Dimensions{
						Items: []filter.Dimension{{Name: "geography"}, {Name: "Day"}, {Name: "Goods and Services"}},
					}, testETag(0), nil)

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, filterID).Return(filter.Model{}, "", errGetJobState)

			Convey("When a CMD filter page is requested", func() {
				callFilterDimensionPage := func(url2 string) *httptest.ResponseRecorder {
					req, err := http.NewRequest(http.MethodGet, url2, http.NoBody)
					So(err, ShouldBeNil)
					req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)

					u, err := url.Parse("http://localhost")
					So(err, ShouldBeNil)

					filterFlexHandler := helpers.CreateReverseProxy("flex", u)
					router := mux.NewRouter()
					w := httptest.NewRecorder()
					f := NewFilter(mockRend, mockFilterClient, mockDatasetClient, mockHierarchyClient, nil, mockZebedeeClient, "/v1", cfg)
					router.Path("/filters/{filterID}/dimensions/{name}").HandlerFunc(f.FilterType(mockDatasetClient, f.FilterOverview(), filterFlexHandler))
					router.ServeHTTP(w, req)
					return w
				}

				w := callFilterDimensionPage(fmt.Sprintf("/filters/%s/dimensions/%s", filterID, dimensionName))
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When the dataset client fails", func() {
			mockDatasetClient.EXPECT().Get(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "").Return(dataset.DatasetDetails{}, errors.New("dataset get error"))
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, filterID).Return(filter.Model{Links: filter.Links{Version: filter.Link{HRef: "/v1/datasets/95c4669b-3ae9-4ba7-b690-87e890a1c67c/editions/2016/versions/1"}}}, testETag(0), nil)

			mockFilterClient.EXPECT().GetDimensions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, nil).
				Return(
					filter.Dimensions{
						Items: []filter.Dimension{{Name: "geography"}, {Name: "Day"}, {Name: "Goods and Services"}},
					}, testETag(0), nil)

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, mockServiceAuthToken, mockDownloadToken, mockCollectionID, filterID).Return(filter.Model{Links: filter.Links{Version: filter.Link{HRef: "/v1/datasets/95c4669b-3ae9-4ba7-b690-87e890a1c67c/editions/2016/versions/1"}}}, testETag(0), nil)

			mockDatasetClient.EXPECT().GetVersionDimensions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1").Return(dataset.VersionDimensions{
				Items: []dataset.VersionDimension{{Name: "geography"}, {Name: "Day"}, {Name: "Goods and services"}, {Name: "unused"}}}, nil)
			mockFilterClient.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "geography", batchSize, maxWorkers).Return(filterGeographyOptions, testETag(0), nil)
			mockFilterClient.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Day", batchSize, maxWorkers).Return(filter.DimensionOptions{}, testETag(0), nil)
			mockFilterClient.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, filterID, "Goods and Services", batchSize, maxWorkers).Return(filter.DimensionOptions{}, testETag(0), nil)
			mockDatasetClient.EXPECT().GetOptionsBatchProcess(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c", "2016", "1", "geography",
				&[]string{"geoUK"}, gomock.Any(), maxDatasetOptions, maxWorkers).Return(nil)
			mockDatasetClient.EXPECT().Get(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, "95c4669b-3ae9-4ba7-b690-87e890a1c67c").Return(dataset.DatasetDetails{}, errors.New("dataset get error"))

			Convey("When a CMD filter page is requested", func() {
				callFilterDimensionPage := func(url2 string) *httptest.ResponseRecorder {
					req, err := http.NewRequest(http.MethodGet, url2, http.NoBody)
					So(err, ShouldBeNil)
					req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)

					u, err := url.Parse("http://localhost")
					So(err, ShouldBeNil)

					filterFlexHandler := helpers.CreateReverseProxy("flex", u)
					router := mux.NewRouter()
					w := httptest.NewRecorder()
					f := NewFilter(mockRend, mockFilterClient, mockDatasetClient, mockHierarchyClient, nil, mockZebedeeClient, "/v1", cfg)
					router.Path("/filters/{filterID}/dimensions/{name}").HandlerFunc(f.FilterType(mockDatasetClient, f.FilterOverview(), filterFlexHandler))
					router.ServeHTTP(w, req)
					return w
				}

				w := callFilterDimensionPage(fmt.Sprintf("/filters/%s/dimensions/%s", filterID, dimensionName))
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
