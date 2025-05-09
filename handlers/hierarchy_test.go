package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/hierarchy"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	core "github.com/ONSdigital/dp-renderer/v2/model"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

// Header keys constants
const (
	CollectionIDHeaderKey  = "Collection-Id"
	FlorenceTokenHeaderKey = "X-Florence-Token"
)

func TestHierarchy(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := gomock.Any()

	mockSearchAPIAuthToken := "testServiceAuthToken"
	mockUserAuthToken := "testUserAuthToken"
	mockCollectionID := "testCollectionID"
	mockDatasetID := "testDatasetID"
	mockEdition := "testEdition"
	mockVersion := "testVersion"
	mockCode := "testCode"

	filterID := "12345"
	dimensionName := "myDimension"
	testInstanceID := "testInstanceID"
	batchSize := 100
	maxWorkers := 25
	maxDatasetOptions := 10
	filterModel := filter.Model{
		InstanceID: testInstanceID,
		FilterID:   filterID,
		Links: filter.Links{
			Version: filter.Link{
				HRef: fmt.Sprintf("http://localhost:1234/v1/datasets/%s/editions/%s/versions/%s", mockDatasetID, mockEdition, mockVersion),
			},
		},
	}

	cfg := &config.Config{
		SearchAPIAuthToken:   mockSearchAPIAuthToken,
		DownloadServiceURL:   "",
		BatchSizeLimit:       batchSize,
		BatchMaxWorkers:      maxWorkers,
		MaxDatasetOptions:    maxDatasetOptions,
		EnableDatasetPreview: false,
	}

	Convey("Given a set of mocked clients and models", t, func() {
		mockFilterClient := NewMockFilterClient(mockCtrl)
		mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
		mockDatasetClient := NewMockDatasetClient(mockCtrl)
		mockRend := NewMockRenderClient(mockCtrl)
		mockZebedeeClient := NewMockZebedeeClient(mockCtrl)

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

		testVersionDimensions := dataset.VersionDimensions{
			Items: dataset.VersionDimensionItems{
				dataset.VersionDimension{
					ID:          "testDimension",
					Name:        "DimensionName",
					Label:       "DimensionLabel",
					Description: "This is mocked Dimension for testing",
				},
			},
		}

		testDatasetDetails := dataset.DatasetDetails{
			ID:    "datasetID",
			Title: "datasetTitle",
		}

		// prepare request for provided url and form, then perform the call with a response writer, which is returned
		callHierarchy := func(url string) *httptest.ResponseRecorder {
			req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
			So(err, ShouldBeNil)
			cookie := http.Cookie{Name: dprequest.CollectionIDCookieKey, Value: mockCollectionID}
			req.AddCookie(&cookie)
			req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)

			router := mux.NewRouter()
			w := httptest.NewRecorder()
			f := NewFilter(mockRend, mockFilterClient, mockDatasetClient, mockHierarchyClient, nil, mockZebedeeClient, "/v1", cfg)
			router.Path("/filters/{filterID}/dimensions/{name}").HandlerFunc(f.Hierarchy())
			router.Path("/filters/{filterID}/dimensions/{name}/{code}").HandlerFunc(f.Hierarchy())
			router.ServeHTTP(w, req)
			return w
		}

		Convey("Hierarchy called for the root node calls the expected methods and returns the expected marshlled hierarchy page", func() {
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, testETag(0), nil)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, dimensionName).Return(hierarchy.Model{}, nil)
			mockFilterClient.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				batchSize, maxWorkers).Return(testSelectedOptions, testETag(0), nil)
			mockDatasetClient.EXPECT().Get(ctx, mockUserAuthToken, "", mockCollectionID, mockDatasetID).Return(testDatasetDetails, nil)
			mockDatasetClient.EXPECT().GetVersionDimensions(ctx, mockUserAuthToken, "", mockCollectionID, mockDatasetID, mockEdition, mockVersion).Return(testVersionDimensions, nil)
			mockDatasetClient.EXPECT().GetOptionsBatchProcess(ctx, mockUserAuthToken, "", mockCollectionID, mockDatasetID, mockEdition, mockVersion, dimensionName,
				&[]string{"op1", "op2"}, gomock.Any(), maxDatasetOptions, maxWorkers).Return(nil)
			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, mockUserAuthToken, mockCollectionID, "en", "/").Return(zebedee.HomepageContent{}, nil)
			mockRend.EXPECT().NewBasePageModel().Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "hierarchy")

			w := callHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s", filterID, dimensionName))

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Hierarchy called for the child node calls the expected methods and returns the expected marshlled hierarchy page", func() {
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, testETag(0), nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, dimensionName, mockCode).Return(hierarchy.Model{}, nil)
			mockFilterClient.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				batchSize, maxWorkers).Return(testSelectedOptions, testETag(0), nil)
			mockDatasetClient.EXPECT().Get(ctx, mockUserAuthToken, "", mockCollectionID, mockDatasetID).Return(testDatasetDetails, nil)
			mockDatasetClient.EXPECT().GetVersionDimensions(ctx, mockUserAuthToken, "", mockCollectionID, mockDatasetID, mockEdition, mockVersion).Return(testVersionDimensions, nil)
			mockDatasetClient.EXPECT().GetOptionsBatchProcess(ctx, mockUserAuthToken, "", mockCollectionID, mockDatasetID, mockEdition, mockVersion, dimensionName,
				&[]string{"op1", "op2"}, gomock.Any(), maxDatasetOptions, maxWorkers).Return(nil)
			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, mockUserAuthToken, mockCollectionID, "en", "/").Return(zebedee.HomepageContent{}, nil)
			mockRend.EXPECT().NewBasePageModel().Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "hierarchy")

			w := callHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/%s", filterID, dimensionName, mockCode))

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Hierarchy called for the root node calls the expected methods. If dataset GetOption fails, an InternalServerError status code is returned", func() {
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, testETag(0), nil)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, dimensionName).Return(hierarchy.Model{}, nil)
			mockFilterClient.EXPECT().GetDimensionOptionsInBatches(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				batchSize, maxWorkers).Return(testSelectedOptions, testETag(0), nil)
			mockDatasetClient.EXPECT().Get(ctx, mockUserAuthToken, "", mockCollectionID, mockDatasetID).Return(testDatasetDetails, nil)
			mockDatasetClient.EXPECT().GetVersionDimensions(ctx, mockUserAuthToken, "", mockCollectionID, mockDatasetID, mockEdition, mockVersion).Return(testVersionDimensions, nil)
			mockDatasetClient.EXPECT().GetOptionsBatchProcess(ctx, mockUserAuthToken, "", mockCollectionID, mockDatasetID, mockEdition, mockVersion, dimensionName,
				&[]string{"op1", "op2"}, gomock.Any(), maxDatasetOptions, maxWorkers).Return(errors.New("error in DatasetAPI"))

			w := callHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s", filterID, dimensionName))

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			So(w.Body.Bytes(), ShouldBeNil)
		})
	})
}

func TestHierarchyUpdate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := gomock.Any()

	mockSearchAPIAuthToken := "testServiceAuthToken"
	mockUserAuthToken := "testUserAuthToken"
	mockCollectionID := "testCollectionID"

	filterID := "12345"
	dimensionName := "myDimension"
	mockCode := "testCode"
	testInstanceID := "testInstanceID"
	batchSize := 100
	filterModel := filter.Model{
		InstanceID: testInstanceID,
		FilterID:   filterID,
	}

	cfg := &config.Config{
		SearchAPIAuthToken:   mockSearchAPIAuthToken,
		DownloadServiceURL:   "",
		BatchSizeLimit:       batchSize,
		EnableDatasetPreview: false,
	}

	Convey("Given a set of mocked clients", t, func() {
		mockFilterClient := NewMockFilterClient(mockCtrl)
		mockHierarchyClient := NewMockHierarchyClient(mockCtrl)

		mockHierarchyModel := hierarchy.Model{
			Children: []hierarchy.Child{
				{
					Links: hierarchy.Links{
						Code: hierarchy.Link{
							ID: "opt1",
						},
					},
				}, {
					Links: hierarchy.Links{
						Code: hierarchy.Link{
							ID: "opt2",
						},
					},
				},
			},
		}

		// prepare request for provided url and form, then perform the call with a response writer, which is returned
		callUpdateHierarchy := func(url string, form url.Values) *httptest.ResponseRecorder {
			req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
			So(err, ShouldBeNil)
			cookie := http.Cookie{Name: dprequest.CollectionIDCookieKey, Value: mockCollectionID}
			req.AddCookie(&cookie)
			req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)
			req.Form = form

			router := mux.NewRouter()
			w := httptest.NewRecorder()
			f := NewFilter(nil, mockFilterClient, nil, mockHierarchyClient, nil, nil, "/v1", cfg)
			router.Path("/filters/{filterID}/dimensions/{name}/update").HandlerFunc(f.HierarchyUpdate())
			router.Path("/filters/{filterID}/dimensions/{name}/{code}/update").HandlerFunc(f.HierarchyUpdate())
			router.ServeHTTP(w, req)
			return w
		}

		Convey("HierarchyUpdate called with a form containing new and overlapping options results in a patch with all options to add and nothing to remove", func() {
			testForm := url.Values{
				"opt3": []string{"v31", "v32", "v33"},
				"opt4": []string{"v41", "v42", "v43"},
				"opt5": []string{"v51", "v52", "v53"},
			}

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, testETag(0), nil)
			mockFilterClient.EXPECT().PatchDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{"opt3", "opt4", "opt5"}), []string{""}, batchSize, testETag(0)).Return(testETag(1), nil)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, dimensionName).Return(hierarchy.Model{}, nil)

			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/update", filterID, dimensionName), testForm)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension\">Found</a>.\n\n")
		})

		Convey("HierarchyUpdate with code and called against a model with children, "+
			"results in options present in children and not in the request being removed via a patch operation", func() {
			testForm := url.Values{
				"opt2": []string{"v31", "v32", "v33"},
			}

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, testETag(0), nil)
			mockFilterClient.EXPECT().PatchDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{"opt2"}), []string{"opt1"}, batchSize, testETag(0)).Return(testETag(1), nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, dimensionName, mockCode).Return(mockHierarchyModel, nil)

			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/%s/update", filterID, dimensionName, mockCode), testForm)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension/testCode\">Found</a>.\n\n")
		})

		Convey("Dimension HierarchyUpdate with a form containing 'add-all' results in a single call to set the options returned by hierarchy GetRoot", func() {
			testForm := url.Values{
				"add-all": []string{"true"},
			}

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, testETag(0), nil)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, dimensionName).Return(mockHierarchyModel, nil)
			mockFilterClient.EXPECT().SetDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{"opt1", "opt2"}), testETag(0)).Return(testETag(1), nil)

			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/update", filterID, dimensionName), testForm)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension\">Found</a>.\n\n")
		})

		Convey("Dimension code HierarchyUpdated with a form containing 'add-all' results in a single call to set the options returned by hierarchy GetChild for the provided code", func() {
			testForm := url.Values{
				"add-all": []string{"true"},
			}

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, testETag(0), nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, dimensionName, mockCode).Return(mockHierarchyModel, nil)
			mockFilterClient.EXPECT().SetDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{"opt1", "opt2"}), testETag(0)).Return(testETag(1), nil)

			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/%s/update", filterID, dimensionName, mockCode), testForm)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension/testCode\">Found</a>.\n\n")
		})

		Convey("Dimension HierarchyUpdate with a form containing 'remove-all' results in a single call to patch-remove the options returned by hierarchy GetRoot", func() {
			testForm := url.Values{
				"remove-all": []string{"true"},
			}

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, testETag(0), nil)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, dimensionName).Return(mockHierarchyModel, nil)
			mockFilterClient.EXPECT().PatchDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{}), ItemsEq([]string{"opt1", "opt2"}), batchSize, testETag(0)).Return(testETag(1), nil)

			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/update", filterID, dimensionName), testForm)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension\">Found</a>.\n\n")
		})

		Convey("Dimension code HierarchyUpdated with a form containing 'remove-all' results in a single call to patch-remove the options returned by hierarchy GetChild for the provided code", func() {
			testForm := url.Values{
				"remove-all": []string{"true"},
			}

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, testETag(0), nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, dimensionName, mockCode).Return(mockHierarchyModel, nil)
			mockFilterClient.EXPECT().PatchDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{}), ItemsEq([]string{"opt1", "opt2"}), batchSize, testETag(0)).Return(testETag(1), nil)

			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/%s/update", filterID, dimensionName, mockCode), testForm)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension/testCode\">Found</a>.\n\n")
		})

		Convey("Then if GetJobState fails, the hierarchy update is aborted and a 500 status code is returned", func() {
			errGetJobState := errors.New("error getting job state")
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filter.Model{}, "", errGetJobState)
			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/update", filterID, dimensionName), nil)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestFlattenGeographyTopLevel(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := context.Background()

	mockSearchAPIAuthToken := "testServiceAuthToken"
	batchSize := 100
	testInstanceID := "testInstanceID"

	expectedDimensionName := "geography"

	cfg := &config.Config{
		SearchAPIAuthToken:   mockSearchAPIAuthToken,
		DownloadServiceURL:   "",
		BatchSizeLimit:       batchSize,
		EnableDatasetPreview: false,
	}

	Convey("Given a successful hierarchy client mock that returns an empty root hierarchy model", t, func() {
		mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
		mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, expectedDimensionName).Return(hierarchy.Model{}, nil)
		f := NewFilter(nil, nil, nil, mockHierarchyClient, nil, nil, "/v1", cfg)

		Convey("then flattenGeographyTopLevel returns an empty hierarchy without error", func() {
			h, err := f.flattenGeographyTopLevel(ctx, testInstanceID)
			So(err, ShouldBeNil)
			So(h, ShouldResemble, hierarchy.Model{})
		})
	})

	Convey("Given a successful hierarchy client mock that returns a valid root hierarchy model with empty children", t, func() {
		testUK := hierarchy.Model{
			Label: "United Kingdom",
			Links: hierarchy.Links{
				Code: hierarchy.Link{ID: Uk},
			},
			HasData:  true,
			Children: []hierarchy.Child{},
		}

		mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
		mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, expectedDimensionName).Return(testUK, nil)
		f := NewFilter(nil, nil, nil, mockHierarchyClient, nil, nil, "/v1", cfg)

		Convey("then flattenGeographyTopLevel returns the root item without error", func() {
			h, err := f.flattenGeographyTopLevel(ctx, testInstanceID)
			So(err, ShouldBeNil)
			So(h, ShouldResemble, testUK)
		})
	})

	Convey("Given mocked hierarchy model and child items", t, func() {
		order0 := 0
		order1 := 10
		order2 := 25
		order3 := 38
		order4 := 41
		order5 := 52

		chWales := hierarchy.Child{
			Label: "Wales",
			Links: hierarchy.Links{
				Code: hierarchy.Link{ID: Wales},
			},
			HasData: true,
			Order:   &order0,
		}

		chEngland := hierarchy.Child{
			Label: "England",
			Links: hierarchy.Links{
				Code: hierarchy.Link{ID: England},
			},
			HasData: true,
			Order:   &order1,
		}

		chNorthernIreland := hierarchy.Child{
			Label: "Northern Ireland",
			Links: hierarchy.Links{
				Code: hierarchy.Link{ID: NorthernIreland},
			},
			HasData: true,
			Order:   &order2,
		}

		chScotland := hierarchy.Child{
			Label: "Scotland",
			Links: hierarchy.Links{
				Code: hierarchy.Link{ID: Scotland},
			},
			HasData: true,
			Order:   &order3,
		}

		chGreatBritain := hierarchy.Child{
			Label: "Great Britain",
			Links: hierarchy.Links{
				Code: hierarchy.Link{ID: GreatBritain},
			},
			HasData: true,
			Order:   &order4,
		}

		chEnglandAndWales := hierarchy.Child{
			Label: "England and Wales",
			Links: hierarchy.Links{
				Code: hierarchy.Link{ID: EnglandAndWales},
			},
			HasData: true,
			Order:   &order5,
		}

		testUK := hierarchy.Model{
			Label: "United Kingdom",
			Links: hierarchy.Links{
				Code: hierarchy.Link{ID: Uk},
			},
			HasData:          true,
			NumberofChildren: 2,
			Children:         []hierarchy.Child{chNorthernIreland, chGreatBritain},
		}

		testGB := hierarchy.Model{
			Label: "Great Britain",
			Links: hierarchy.Links{
				Code: hierarchy.Link{ID: GreatBritain},
			},
			HasData:          true,
			Order:            &order4,
			NumberofChildren: 2,
			Children:         []hierarchy.Child{chScotland, chEnglandAndWales},
		}

		testEnglandAndWales := hierarchy.Model{
			Label: "England and Wales",
			Links: hierarchy.Links{
				Code: hierarchy.Link{ID: EnglandAndWales},
			},
			HasData:          true,
			Order:            &order5,
			NumberofChildren: 2,
			Children:         []hierarchy.Child{chWales, chEngland},
		}

		Convey("And a successful hierarchy client mock where all models contain order", func() {
			mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, expectedDimensionName).Return(testUK, nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, expectedDimensionName, GreatBritain).Return(testGB, nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, expectedDimensionName, EnglandAndWales).Return(testEnglandAndWales, nil)
			f := NewFilter(nil, nil, nil, mockHierarchyClient, nil, nil, "/v1", cfg)

			Convey("then flattenGeographyTopLevel returns a flat list of geography nodes sorted in the order defined by the children order property", func() {
				expectedFlatGeography := hierarchy.Model{
					Label:   "United Kingdom",
					HasData: true,
					Links: hierarchy.Links{
						Code: hierarchy.Link{ID: Uk},
					},
					Children: []hierarchy.Child{chWales, chEngland, chNorthernIreland, chScotland, chGreatBritain, chEnglandAndWales},
				}

				h, err := f.flattenGeographyTopLevel(ctx, testInstanceID)
				So(err, ShouldBeNil)
				So(h, ShouldResemble, expectedFlatGeography)
			})
		})

		Convey("And a successful hierarchy client mock where some models don't contain order", func() {
			// expected without order
			chScotland.Order = nil

			// mock hierarchy response without order
			testGB.Children[0].Order = nil

			mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, expectedDimensionName).Return(testUK, nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, expectedDimensionName, GreatBritain).Return(testGB, nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, expectedDimensionName, EnglandAndWales).Return(testEnglandAndWales, nil)
			f := NewFilter(nil, nil, nil, mockHierarchyClient, nil, nil, "/v1", cfg)

			Convey("then flattenGeographyTopLevel returns a flat list of geography nodes sorted according to the default hardcoded order", func() {
				expectedFlatGeography := hierarchy.Model{
					Label:   "United Kingdom",
					HasData: true,
					Links: hierarchy.Links{
						Code: hierarchy.Link{ID: Uk},
					},
					Children: []hierarchy.Child{chGreatBritain, chEnglandAndWales, chEngland, chNorthernIreland, chScotland, chWales},
				}

				h, err := f.flattenGeographyTopLevel(ctx, testInstanceID)
				So(err, ShouldBeNil)
				So(h, ShouldResemble, expectedFlatGeography)
			})
		})

		Convey("And a successful hierarchy client mock where none of the models contain order", func() {
			// expected without order
			chWales.Order = nil
			chEngland.Order = nil
			chNorthernIreland.Order = nil
			chScotland.Order = nil
			chGreatBritain.Order = nil
			chEnglandAndWales.Order = nil

			// mock hierarchy response without order
			testUK.Order = nil
			testUK.Children[0].Order = nil
			testUK.Children[1].Order = nil
			testGB.Order = nil
			testGB.Children[0].Order = nil
			testGB.Children[1].Order = nil
			testEnglandAndWales.Order = nil
			testEnglandAndWales.Children[0].Order = nil
			testEnglandAndWales.Children[1].Order = nil

			mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, expectedDimensionName).Return(testUK, nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, expectedDimensionName, GreatBritain).Return(testGB, nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, expectedDimensionName, EnglandAndWales).Return(testEnglandAndWales, nil)
			f := NewFilter(nil, nil, nil, mockHierarchyClient, nil, nil, "/v1", cfg)

			Convey("then flattenGeographyTopLevel returns a flat list of geography nodes sorted according to the default hardcoded order", func() {
				expectedFlatGeography := hierarchy.Model{
					Label:   "United Kingdom",
					HasData: true,
					Links: hierarchy.Links{
						Code: hierarchy.Link{ID: Uk},
					},
					Children: []hierarchy.Child{chGreatBritain, chEnglandAndWales, chEngland, chNorthernIreland, chScotland, chWales},
				}

				h, err := f.flattenGeographyTopLevel(ctx, testInstanceID)
				So(err, ShouldBeNil)
				So(h, ShouldResemble, expectedFlatGeography)
			})
		})

		Convey("And a hierarchy client that fails on getRoot", func() {
			testErr := errors.New("testError")
			mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, expectedDimensionName).Return(hierarchy.Model{}, testErr)
			f := NewFilter(nil, nil, nil, mockHierarchyClient, nil, nil, "/v1", cfg)

			Convey("then flattenGeographyTopLevel fails with the same error and no other call is performed", func() {
				_, err := f.flattenGeographyTopLevel(ctx, testInstanceID)
				So(err, ShouldResemble, testErr)
			})
		})

		Convey("And a hierarchy client that fails on getChildren for greatBritain code", func() {
			testErr := errors.New("testError")
			mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, expectedDimensionName).Return(testUK, nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, expectedDimensionName, GreatBritain).Return(hierarchy.Model{}, testErr)
			f := NewFilter(nil, nil, nil, mockHierarchyClient, nil, nil, "/v1", cfg)

			Convey("then flattenGeographyTopLevel fails with the same error and no other call is performed", func() {
				_, err := f.flattenGeographyTopLevel(ctx, testInstanceID)
				So(err, ShouldResemble, testErr)
			})
		})

		Convey("And a hierarchy client that fails on getChildren for englandAndWales code", func() {
			testErr := errors.New("testError")
			mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, expectedDimensionName).Return(testUK, nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, expectedDimensionName, GreatBritain).Return(testGB, nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, expectedDimensionName, EnglandAndWales).Return(hierarchy.Model{}, testErr)
			f := NewFilter(nil, nil, nil, mockHierarchyClient, nil, nil, "/v1", cfg)

			Convey("then flattenGeographyTopLevel fails with the same error and no other call is performed", func() {
				_, err := f.flattenGeographyTopLevel(ctx, testInstanceID)
				So(err, ShouldResemble, testErr)
			})
		})
	})
}

func TestFlatNodes_AddWithChildren(t *testing.T) {
	testChild := hierarchy.Child{
		Label: "testLabel",
		Links: hierarchy.Links{
			Self: hierarchy.Link{ID: "selfID", URL: "selfURL"},
			Code: hierarchy.Link{ID: "codeID", URL: "codeURL"},
		},
		HasData:          true,
		Order:            nil,
		NumberofChildren: 2,
	}

	Convey("given an empty flatNodes", t, func() {
		var n = flatNodes{}

		Convey("Then addWithChildren will add the child value to the list", func() {
			n.addWithChildren(testChild)
			So(n, ShouldResemble, flatNodes{
				list: []hierarchy.Child{testChild},
			})
		})
	})
}

func TestFlatNodes_AddWithoutChildren(t *testing.T) {
	testChildWithData := hierarchy.Child{
		Label: "testLabel",
		Links: hierarchy.Links{
			Self: hierarchy.Link{ID: "selfID", URL: "selfURL"},
			Code: hierarchy.Link{ID: "codeID", URL: "codeURL"},
		},
		HasData:          true,
		Order:            nil,
		NumberofChildren: 0,
	}

	testChildWithoutData := hierarchy.Child{
		Label: "testLabel",
		Links: hierarchy.Links{
			Self: hierarchy.Link{ID: "selfID", URL: "selfURL"},
			Code: hierarchy.Link{ID: "codeID", URL: "codeURL"},
		},
		HasData:          false,
		Order:            nil,
		NumberofChildren: 0,
	}

	Convey("given a nil flatNodes variable, then addWithoutChildren does not panic", t, func() {
		var n *flatNodes = nil
		n.addWithoutChildren(hierarchy.Child{})
		So(n, ShouldBeNil)
	})

	Convey("given an empty flatNodes", t, func() {
		var n = flatNodes{}

		Convey("Then addWithoutChildren with a child that has data will add the child value to the list", func() {
			n.addWithoutChildren(testChildWithData)
			So(n, ShouldResemble, flatNodes{
				list: []hierarchy.Child{testChildWithData},
			})
		})

		Convey("Then addWithoutChildren with a child that does not have data will not add the child value to the list", func() {
			n.addWithoutChildren(testChildWithoutData)
			So(n, ShouldResemble, flatNodes{})
		})
	})
}

func TestFlatNodes_HasOrder(t *testing.T) {
	order1 := 1
	order2 := 2

	Convey("given an empty flatNodes, then hasOrder returns false", t, func() {
		var n = flatNodes{}
		So(n.hasOrder(), ShouldBeFalse)
	})

	Convey("given a flatNodes with all items in the list having order, then hasOrder returns true", t, func() {
		var n = flatNodes{
			list: []hierarchy.Child{
				{Label: "child1", Order: &order1},
				{Label: "child2", Order: &order2},
			},
		}
		So(n.hasOrder(), ShouldBeTrue)
	})

	Convey("given a flatNodes with a mixture of items with order and items without order in the list, then hasOrder returns false", t, func() {
		var n = flatNodes{
			list: []hierarchy.Child{
				{Label: "child1", Order: &order1},
				{Label: "child2", Order: nil},
			},
		}
		So(n.hasOrder(), ShouldBeFalse)
	})

	Convey("given a flatNodes with all items in the list not having order, then hasOrder returns false", t, func() {
		var n = flatNodes{
			list: []hierarchy.Child{
				{Label: "child1", Order: nil},
				{Label: "child2", Order: nil},
			},
		}
		So(n.hasOrder(), ShouldBeFalse)
	})
}

func TestFlatNodes_GetOrder(t *testing.T) {
	order1 := 1
	order2 := 2

	Convey("given an empty flatNodes, then getOrder returns -1", t, func() {
		var n = flatNodes{}
		So(n.getOrder(0), ShouldEqual, -1)
	})

	Convey("given a flatNodes with items with order and without order", t, func() {
		var n = flatNodes{
			list: []hierarchy.Child{
				{Label: "child1", Order: &order1},
				{Label: "child2", Order: &order2},
				{Label: "child3", Order: nil},
			},
		}

		Convey("Then getOrder returns the order value for nodes with order, and -1 for nodes without order, nil or out of range", func() {
			So(n.getOrder(0), ShouldEqual, order1)
			So(n.getOrder(1), ShouldEqual, order2)
			So(n.getOrder(2), ShouldEqual, -1)
			So(n.getOrder(3), ShouldEqual, -1)
		})
	})
}

func TestFlatNodes_GetDefaultOrder(t *testing.T) {
	order1 := 1

	Convey("given an empty flatNodes, then getDefaultOrder returns -1", t, func() {
		var n = flatNodes{}
		So(n.getDefaultOrder(0), ShouldEqual, -1)
	})

	Convey("given a flatNodes with a list of unordered items but without a default order list", t, func() {
		var n = flatNodes{
			list: []hierarchy.Child{
				{Label: "child1", Order: nil, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code1"}}},
			},
		}

		Convey("Then getDefaultOrder returns -1 for indices within range and out of range", func() {
			So(n.getDefaultOrder(0), ShouldEqual, -1)
			So(n.getDefaultOrder(1), ShouldEqual, -1)
		})
	})

	Convey("given a flatNodes with a list of unordered items and a default order list", t, func() {
		var n = flatNodes{
			list: []hierarchy.Child{
				{Label: "child1", Order: nil, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code1"}}},
				{Label: "child2", Order: nil, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code2"}}},
				{Label: "child3", Order: nil, Links: hierarchy.Links{Code: hierarchy.Link{}}},
				{Label: "child4", Order: nil, Links: hierarchy.Links{}},
				{Label: "child5", Order: nil},
			},
			defaultOrder: map[string]int{
				"code1": order1,
			},
		}

		Convey("Then getDefaultOrder returns the expected values for items with order and -1 in any other case", func() {
			So(n.getDefaultOrder(0), ShouldEqual, order1)
			So(n.getDefaultOrder(1), ShouldEqual, -1)
			So(n.getDefaultOrder(2), ShouldEqual, -1)
			So(n.getDefaultOrder(3), ShouldEqual, -1)
			So(n.getDefaultOrder(4), ShouldEqual, -1)
			So(n.getDefaultOrder(5), ShouldEqual, -1)
		})
	})
}

func TestFlatNodes_Sort(t *testing.T) {
	order1 := 1
	order2 := 2
	order3 := 3
	order4 := 4

	Convey("given an empty flatNodes, then sort does not panic", t, func() {
		var n = flatNodes{}
		n.sort()
		So(n, ShouldResemble, flatNodes{})
	})

	Convey("given a flatNodes with a list of children with order", t, func() {
		var n = flatNodes{
			list: []hierarchy.Child{
				{Label: "child1", Order: &order3},
				{Label: "child2", Order: &order4},
				{Label: "child3", Order: &order1},
				{Label: "child4", Order: &order2},
			},
		}

		Convey("Then sort sorts the list of children according to their order values, incrementally", func() {
			n.sort()
			So(n, ShouldResemble, flatNodes{
				list: []hierarchy.Child{
					{Label: "child3", Order: &order1},
					{Label: "child4", Order: &order2},
					{Label: "child1", Order: &order3},
					{Label: "child2", Order: &order4},
				},
			})
		})
	})

	Convey("given a flatNodes with a list of children where at least one of them does not have order, and a list of default orders that contains default order for all the items in list", t, func() {
		var n = flatNodes{
			list: []hierarchy.Child{
				{Label: "child1", Order: &order1, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code1"}}},
				{Label: "child2", Order: nil, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code2"}}},
				{Label: "child3", Order: &order3, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code3"}}},
				{Label: "child4", Order: &order4, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code4"}}},
			},
			defaultOrder: map[string]int{
				"code1": order4,
				"code2": order3,
				"code3": order1,
				"code4": order2,
			},
		}

		Convey("Then sort sorts the list of children according to the default order values, incrementally", func() {
			n.sort()
			So(n.list, ShouldResemble, []hierarchy.Child{
				{Label: "child3", Order: &order3, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code3"}}},
				{Label: "child4", Order: &order4, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code4"}}},
				{Label: "child2", Order: nil, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code2"}}},
				{Label: "child1", Order: &order1, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code1"}}},
			})
		})
	})

	Convey("given a flatNodes with a list of children where at least one of them does not have order, and a list of default orders that does not contain some of the items in list", t, func() {
		var n = flatNodes{
			list: []hierarchy.Child{
				{Label: "child1", Order: &order1, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code1"}}},
				{Label: "child2", Order: nil, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code2"}}},
				{Label: "child3", Order: &order3, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code3"}}},
				{Label: "child4", Order: &order4, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code4"}}},
			},
			defaultOrder: map[string]int{
				"code1": order4,
				"code2": order3,
				"code3": order1,
			},
		}

		Convey("Then sort sorts the list of children according to the default order values that exist, incrementally, after the ones that do not have a default order, which appear first in the list", func() {
			n.sort()
			So(n.list, ShouldResemble, []hierarchy.Child{
				{Label: "child4", Order: &order4, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code4"}}},
				{Label: "child3", Order: &order3, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code3"}}},
				{Label: "child2", Order: nil, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code2"}}},
				{Label: "child1", Order: &order1, Links: hierarchy.Links{Code: hierarchy.Link{ID: "code1"}}},
			})
		})
	})
}
