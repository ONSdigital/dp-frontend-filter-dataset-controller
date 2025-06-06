package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/search"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	core "github.com/ONSdigital/dp-renderer/v2/model"
	gomock "github.com/golang/mock/gomock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitSearch(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := gomock.Any()

	mockUserAuthToken := "testUserAuthToken"
	mockServiceAuthToken := "testServiceAuthToken"
	mockCollectionID := "testCollectionID"

	filterID := "12345"
	datasetID := "abcde"
	name := "aggregate"
	edition := "2017"
	version := "1"
	query := "Newport"
	batchSize := 100
	maxWorkers := 25
	maxDatasetOptions := 10

	cfg := &config.Config{
		SearchAPIAuthToken:   mockServiceAuthToken,
		DownloadServiceURL:   "",
		BatchSizeLimit:       batchSize,
		MaxDatasetOptions:    maxDatasetOptions,
		BatchMaxWorkers:      maxWorkers,
		EnableDatasetPreview: false,
	}

	testSelectedOptions := filter.DimensionOptions{
		Items: []filter.DimensionOption{
			{Option: "op1"},
			{Option: "op2"},
		},
		Count:      2,
		TotalCount: 2,
		Limit:      0,
		Offset:     0,
	}

	Convey("Given a set of mocks for filter, dataset, search and renderer clients", t, func() {
		mfc := NewMockFilterClient(mockCtrl)
		mdc := NewMockDatasetClient(mockCtrl)
		msc := NewMockSearchClient(mockCtrl)
		mrc := NewMockRenderClient(mockCtrl)
		mzc := NewMockZebedeeClient(mockCtrl)

		callSearch := func() *httptest.ResponseRecorder {
			target := fmt.Sprintf("/filters/%s/dimensions/%s/search?q=%s", filterID, name, query)
			req := httptest.NewRequest(http.MethodGet, target, http.NoBody)
			req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)
			req.Header.Add(dprequest.CollectionIDHeaderKey, mockCollectionID)
			w := httptest.NewRecorder()
			router := mux.NewRouter()
			f := NewFilter(mrc, mfc, mdc, nil, msc, mzc, "/v1", cfg)
			router.Path("/filters/{filterID}/dimensions/{name}/search").Methods(http.MethodGet).HandlerFunc(f.Search())
			router.ServeHTTP(w, req)
			return w
		}

		expectedSearchClientConfigs := []search.Config{
			{
				InternalToken: mockServiceAuthToken,
				FlorenceToken: mockUserAuthToken,
			},
		}

		Convey("Then search can successfully load a page", func() {
			mfc.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:23200/v1/datasets/abcde/editions/2017/versions/1",
					},
				},
			}, testETag(0), nil)
			mfc.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name,
				batchSize, maxWorkers).Return(testSelectedOptions, testETag(0), nil)
			mdc.EXPECT().Get(ctx, mockUserAuthToken, "", mockCollectionID, datasetID).Return(dataset.DatasetDetails{}, nil)
			mdc.EXPECT().GetVersionDimensions(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version).Return(dataset.VersionDimensions{}, nil)
			mzc.EXPECT().GetHomepageContent(ctx, mockUserAuthToken, mockCollectionID, "en", "/").Return(zebedee.HomepageContent{}, nil)
			mdc.EXPECT().GetOptionsBatchProcess(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				&[]string{"op1", "op2"}, gomock.Any(), maxDatasetOptions, maxWorkers).Return(nil)
			msc.EXPECT().Dimension(ctx, datasetID, edition, version, name, query, expectedSearchClientConfigs).Return(&search.Model{}, nil)
			mrc.EXPECT().NewBasePageModel().Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mrc.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "hierarchy")

			w := callSearch()
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Then search returns server error if GetJobState errors", func() {
			mfc.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filter.Model{}, "", errors.New("get job state error"))

			w := callSearch()
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Then search returns server error if GetDimensionOptions errors", func() {
			mfc.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:23200/v1/datasets/abcde/editions/2017/versions/1",
					},
				},
			}, testETag(0), nil)
			mfc.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name,
				batchSize, maxWorkers).Return(filter.DimensionOptions{}, "", errors.New("get dimensions options error"))

			w := callSearch()
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Then search returns server error if Dataset Get errors", func() {
			mfc.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:23200/v1/datasets/abcde/editions/2017/versions/1",
					},
				},
			}, testETag(0), nil)
			mfc.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name,
				batchSize, maxWorkers).Return(filter.DimensionOptions{}, testETag(0), nil)
			mdc.EXPECT().Get(ctx, mockUserAuthToken, "", mockCollectionID, datasetID).Return(dataset.DatasetDetails{}, errors.New("dataset get error"))

			w := callSearch()
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Then search returns server error if GetOptionsBatchProcess errors", func() {
			mfc.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:23200/v1/datasets/abcde/editions/2017/versions/1",
					},
				},
			}, testETag(0), nil)
			mfc.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name,
				batchSize, maxWorkers).Return(testSelectedOptions, testETag(0), nil)
			mdc.EXPECT().Get(ctx, mockUserAuthToken, "", mockCollectionID, datasetID).Return(dataset.DatasetDetails{}, nil)
			mdc.EXPECT().GetOptionsBatchProcess(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				&[]string{"op1", "op2"}, gomock.Any(), maxDatasetOptions, maxWorkers).Return(errors.New("get options error"))
			w := callSearch()
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Then search returns server error if search api call errors", func() {
			mfc.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:23200/v1/datasets/abcde/editions/2017/versions/1",
					},
				},
			}, testETag(0), nil)
			mfc.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name,
				batchSize, maxWorkers).Return(testSelectedOptions, testETag(0), nil)
			mdc.EXPECT().Get(ctx, mockUserAuthToken, "", mockCollectionID, datasetID).Return(dataset.DatasetDetails{}, nil)
			mdc.EXPECT().GetOptionsBatchProcess(ctx, mockUserAuthToken, "", mockCollectionID, datasetID, edition, version, name,
				&[]string{"op1", "op2"}, gomock.Any(), maxDatasetOptions, maxWorkers).Return(nil)
			msc.EXPECT().Dimension(ctx, datasetID, edition, version, name, query, expectedSearchClientConfigs).Return(&search.Model{}, errors.New("search api error"))

			w := callSearch()
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Then search returns internal server error if version url cannot be parsed", func() {
			mfc.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filter.Model{
				Links: filter.Links{
					Version: filter.Link{
						HRef: "http://localhost:23200/v1/datasets",
					},
				},
			}, testETag(0), nil)
			mfc.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name,
				batchSize, maxWorkers).Return(filter.DimensionOptions{}, testETag(0), nil)

			w := callSearch()
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestSearchUpdate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := gomock.Any()

	mockUserAuthToken := "testUserAuthToken"
	mockServiceAuthToken := "testServiceAuthToken"
	mockCollectionID := "testCollectionID"

	filterID := "12345"
	name := "aggregate"
	batchSize := 100
	maxWorkers := 25

	cfg := &config.Config{
		SearchAPIAuthToken:   mockServiceAuthToken,
		DownloadServiceURL:   "",
		BatchSizeLimit:       batchSize,
		BatchMaxWorkers:      maxWorkers,
		EnableDatasetPreview: false,
	}

	Convey("Given a set of mocks for filter, dataset, search and renderer clients", t, func() {
		mfc := NewMockFilterClient(mockCtrl)
		mdc := NewMockDatasetClient(mockCtrl)
		msc := NewMockSearchClient(mockCtrl)
		mrc := NewMockRenderClient(mockCtrl)
		mzc := NewMockZebedeeClient(mockCtrl)

		callSearchUpdate := func(formData string) *httptest.ResponseRecorder {
			target := fmt.Sprintf("/filters/%s/dimensions/%s/search/update", filterID, name)
			reader := strings.NewReader(formData)
			req := httptest.NewRequest(http.MethodPost, target, reader)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)
			req.Header.Add(dprequest.CollectionIDHeaderKey, mockCollectionID)
			w := httptest.NewRecorder()
			router := mux.NewRouter()
			f := NewFilter(mrc, mfc, mdc, nil, msc, mzc, "/v1", cfg)
			router.Path("/filters/{filterID}/dimensions/{name}/search/update").HandlerFunc(f.SearchUpdate())
			router.ServeHTTP(w, req)
			return w
		}

		expectedSearchClientConfigs := []search.Config{
			{
				InternalToken: mockServiceAuthToken,
				FlorenceToken: mockUserAuthToken,
			},
		}

		filterModel := filter.Model{
			Links: filter.Links{
				Version: filter.Link{
					HRef: "http://localhost:23200/v1/datasets/cpih01/editions/time-series/versions/1",
				},
			},
		}

		Convey("When the request form includes 'add-all', all dimension values are set and the user is redirected", func() {
			searchModel := &search.Model{
				Items: []search.Item{
					{Code: "clothing-1"},
					{Code: "clothing-2"},
					{Code: "clothing-3"},
				},
			}
			options := []string{"clothing-1", "clothing-2", "clothing-3"}
			mfc.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, testETag(0), nil)
			msc.EXPECT().Dimension(ctx, "cpih01", "time-series", "1", name, "clothing", expectedSearchClientConfigs).Return(searchModel, nil)
			mfc.EXPECT().SetDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name, ItemsEq(options), testETag(0)).Return(testETag(1), nil)
			formData := "q=clothing&cpih1dim1G30100=on&cpih1dim1G30200=on&save-and-return=Save+and+return&add-all=true"
			w := callSearchUpdate(formData)
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("When the request form includes 'remove-all', all dimension values are removed and the user is redirected", func() {
			searchModel := &search.Model{
				Items: []search.Item{
					{Code: "clothing-1"},
					{Code: "clothing-2"},
					{Code: "clothing-3"},
				},
			}
			options := []string{"clothing-1", "clothing-2", "clothing-3"}
			mfc.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, testETag(0), nil)
			msc.EXPECT().Dimension(ctx, "cpih01", "time-series", "1", name, "clothing", expectedSearchClientConfigs).Return(searchModel, nil)
			mfc.EXPECT().PatchDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name, []string{}, ItemsEq(options), batchSize, testETag(0)).Return(testETag(1), nil)
			formData := "q=clothing&cpih1dim1G30100=on&cpih1dim1G30200=on&save-and-return=Save+and+return&remove-all=true"
			w := callSearchUpdate(formData)
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("When the request doesn't contain add-all or remove-all, then the selected options are added and the unselected are removed, in a single PATCH call. The user is redirected", func() {
			searchModel := &search.Model{
				Items: []search.Item{
					{Code: "clothing-1"},
					{Code: "clothing-2"},
					{Code: "clothing-3"},
					{Code: "clothing-4"},
				},
			}
			filterOptions := filter.DimensionOptions{
				Items: []filter.DimensionOption{
					{Option: "clothing-1"},
					{Option: "clothing-2"},
					{Option: "clothing-3"},
					{Option: "clothing-4"},
					{Option: "clothing-5"},
				},
				Count:      5,
				TotalCount: 5,
				Limit:      0,
				Offset:     0,
			}
			expectedAddOptions := []string{"clothing-1", "clothing-2", "clothing-3"}
			expectedRemoveOptions := []string{"clothing-4"}
			mfc.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, testETag(0), nil)
			msc.EXPECT().Dimension(ctx, "cpih01", "time-series", "1", name, "clothing", expectedSearchClientConfigs).Return(searchModel, nil)
			mfc.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name,
				batchSize, maxWorkers).Return(filterOptions, testETag(0), nil)
			mfc.EXPECT().PatchDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name,
				ItemsEq(expectedAddOptions), ItemsEq(expectedRemoveOptions), batchSize, testETag(0)).Return(testETag(1), nil)
			formData := "q=clothing&clothing-1=on&clothing-2=on&clothing-3=on&save-and-return=Save+and+return"
			w := callSearchUpdate(formData)
			So(w.Code, ShouldEqual, http.StatusFound)
		})

		Convey("When GetDimensionOptions fails with a generic error, then the execution is aborted and an Internal Server Error is returned.", func() {
			mfc.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, testETag(0), nil)
			msc.EXPECT().Dimension(ctx, "cpih01", "time-series", "1", name, "clothing", expectedSearchClientConfigs).Return(&search.Model{}, nil)
			mfc.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, "", mockCollectionID, filterID, name,
				batchSize, maxWorkers).Return(filter.DimensionOptions{}, "", errors.New("Error getting dimension options"))
			formData := "q=clothing&clothing-1=on&clothing-2=on&clothing-3=on&save-and-return=Save+and+return"
			w := callSearchUpdate(formData)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
