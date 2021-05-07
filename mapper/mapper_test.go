package mapper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/hierarchy"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/age"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/filterOverview"
	hierarchyModel "github.com/ONSdigital/dp-frontend-models/model/dataset-filter/hierarchy"
	timeModel "github.com/ONSdigital/dp-frontend-models/model/dataset-filter/time"
	dprequest "github.com/ONSdigital/dp-net/request"
	. "github.com/smartystreets/goconvey/convey"
)

// getExpectedFilterOverviewPage returns the filterOverview.Page model that would be generated from all-empty values
func getExpectedFilterOverviewPage() filterOverview.Page {
	expectedPageModel := filterOverview.Page{
		Data: filterOverview.FilterOverview{
			Dimensions: nil,
			Cancel: filterOverview.Link{
				URL: "/",
			},
			ClearAll: filterOverview.Link{
				URL: "/filters//dimensions/clear-all",
			},
		},
	}
	expectedPageModel.URI = "/"
	expectedPageModel.Breadcrumb = []model.TaxonomyNode{
		{URI: "/datasets//editions"},
		{},
		{Title: "Filter options"},
	}
	expectedPageModel.Data.PreviewAndDownloadDisabled = true
	expectedPageModel.SearchDisabled = true
	expectedPageModel.IsInFilterBreadcrumb = true
	expectedPageModel.BetaBannerEnabled = true
	expectedPageModel.Metadata = model.Metadata{Title: "Filter Options"}
	expectedPageModel.CookiesPolicy = model.CookiesPolicy{Essential: true}
	return expectedPageModel
}

func TestCreateFilterOverview(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	filterID := "12349876"
	datasetID := "12345"
	releaseDate := "18asdfoh1-11-1992"
	apiRouterVersion := "/v1"
	lang := dprequest.DefaultLang

	Convey("Calling CreateFilterOverview with empty values returns the expected filter overview page without error", t, func() {
		expectedFop := getExpectedFilterOverviewPage()
		fop := CreateFilterOverview(req, []filter.ModelDimension{}, dataset.VersionDimensionItems{}, filter.Model{}, dataset.DatasetDetails{}, "", "", "", "", "")
		So(fop, ShouldResemble, expectedFop)
	})

	Convey("Calling CreateFilterOverview with non-empty values for filterID, datasetID, release date, apiRouterVersion and language returns the expected filter overview page without error", t, func() {
		expectedFop := getExpectedFilterOverviewPage()
		expectedFop.DatasetId = datasetID
		expectedFop.FilterID = filterID
		expectedFop.Language = lang
		expectedFop.Data.ClearAll.URL = fmt.Sprintf("/filters/%s/dimensions/clear-all", filterID)
		fop := CreateFilterOverview(req, []filter.ModelDimension{}, dataset.VersionDimensionItems{}, filter.Model{}, dataset.DatasetDetails{}, filterID, datasetID, releaseDate, apiRouterVersion, lang)
		So(fop, ShouldResemble, expectedFop)
	})

	Convey("Given mocked datasetDimensions, filter and dataset", t, func() {
		datasetDimension := getTestDatasetDimensions()
		f := getTestFilter()
		dst := getTestDataset()

		expectedFop := getExpectedFilterOverviewPage()
		expectedFop.DatasetId = datasetID
		expectedFop.FilterID = filterID
		expectedFop.Language = lang
		expectedFop.Data.ClearAll.URL = fmt.Sprintf("/filters/%s/dimensions/clear-all", filterID)
		expectedFop.DatasetTitle = "Small Area Population Estimates"
		expectedFop.Data = filterOverview.FilterOverview{
			PreviewAndDownload: filterOverview.Link{
				URL: "/filters/" + f.FilterID,
			},
			Cancel: filterOverview.Link{
				URL: "/",
			},
			ClearAll: filterOverview.Link{
				URL: "/filters/12349876/dimensions/clear-all",
			},
		}
		expectedFop.Breadcrumb = []model.TaxonomyNode{
			{
				Title: dst.Title,
				URI:   "/datasets//editions",
			},
			{
				Title: "5678",
				URI:   "/datasets/1234/editions/5678/versions/1",
			},
			{
				Title: "Filter options",
			},
		}

		Convey("And a dimensions without values, then calling CreateFilterOverview returns the expected filter overview page, with 'Add' label for that dimension, without error", func() {
			dimensions := []filter.ModelDimension{
				{
					Name:   "emptyDimension",
					Values: []string{},
				},
			}
			expectedFop.Data.Dimensions = []filterOverview.Dimension{
				{
					Filter:          "",
					AddedCategories: nil,
					Link: filterOverview.Link{
						Label: "Add",
						URL:   "/filters/" + f.FilterID + "/dimensions/emptyDimension",
					},
				},
			}
			expectedFop.Data.PreviewAndDownloadDisabled = true
			expectedFop.Data.PreviewAndDownload = filterOverview.Link{}
			fop := CreateFilterOverview(req, dimensions, datasetDimension, f, dst, filterID, datasetID, releaseDate, apiRouterVersion, lang)
			So(fop, ShouldResemble, expectedFop)
		})

		Convey("And a set of generic dimensions, then calling CreateFilterOverview returns the expected filter overview page without error", func() {
			dimensions := getTestDimensions()
			expectedFop.Data.Dimensions = []filterOverview.Dimension{
				{
					Filter:          "Year",
					AddedCategories: []string{"2014"},
					Link: filterOverview.Link{
						Label: "Edit",
						URL:   "/filters/" + f.FilterID + "/dimensions/year",
					},
				},
				{
					Filter:          "Geographic Areas",
					AddedCategories: []string{"England and Wales", "Bristol"},
					Link: filterOverview.Link{
						Label: "Edit",
						URL:   "/filters/" + f.FilterID + "/dimensions/geography",
					},
				},
				{
					Filter:          "Sex",
					AddedCategories: []string{"All persons"},
					Link: filterOverview.Link{
						Label: "Edit",
						URL:   "/filters/" + f.FilterID + "/dimensions/sex",
					},
				},
				{
					Filter:          "Age",
					AddedCategories: []string{"0 - 92", "2 - 18", "18 - 65"},
					Link: filterOverview.Link{
						Label: "Edit",
						URL:   "/filters/" + f.FilterID + "/dimensions/age-range",
					},
				},
				{
					Filter:          "Time",
					AddedCategories: []string{"2002.10", "2009.08", "1996.08"},
					Link: filterOverview.Link{
						Label: "Edit",
						URL:   "/filters/12349876/dimensions/time",
					},
				},
			}
			fop := CreateFilterOverview(req, dimensions, datasetDimension, f, dst, filterID, datasetID, releaseDate, apiRouterVersion, lang)
			So(fop, ShouldResemble, expectedFop)
		})

		Convey("And a time dimension with value with format 'Jan-06", func() {
			dimensions := []filter.ModelDimension{
				{
					Name:   "time",
					Values: []string{"Jan-01", "Sep-08", "Apr-85"},
				},
			}
			expectedFop.Data.Dimensions = []filterOverview.Dimension{
				{
					Filter:          "Time",
					AddedCategories: []string{"January 2001", "September 2008", "April 1985"},
					Link: filterOverview.Link{
						Label: "Edit",
						URL:   "/filters/12349876/dimensions/time",
					},
				},
			}

			Convey("Then CreateFilterOverview returns the exected filter overview page, formatted as 'January 2006', sorted in the same order as provided", func() {
				fop := CreateFilterOverview(req, dimensions, datasetDimension, f, dst, filterID, datasetID, releaseDate, apiRouterVersion, lang)
				So(fop, ShouldResemble, expectedFop)
			})

			Convey("And a corresponding dataset dimension with a non-empty label for the time dimension, then CreateFilterOverview returns the exected filter overview page, formatted as 'January 2006', sorted in the same order as provided and with the Filter value overwritten by the label", func() {
				datasetDimension[5] = dataset.VersionDimension{
					Name:  "time",
					Label: "TimeOverwrite",
				}
				expectedFop.Data.Dimensions[0].Filter = "TimeOverwrite"
				fop := CreateFilterOverview(req, dimensions, datasetDimension, f, dst, filterID, datasetID, releaseDate, apiRouterVersion, lang)
				So(fop, ShouldResemble, expectedFop)
			})
		})

		Convey("And an age dimension with some sorting that is not from young to old, then CreateFilterOverview returns the items sorted in the same order as provided", func() {
			dimensions := []filter.ModelDimension{
				{
					Name:   "age",
					Values: []string{"20", "40", "60", "80", "100+", "nonnumerical", "10", "30", "50", "70", "90"},
				},
			}
			expectedFop.Data.Dimensions = []filterOverview.Dimension{
				{
					Filter:          "Age",
					AddedCategories: []string{"20", "40", "60", "80", "100+", "nonnumerical", "10", "30", "50", "70", "90"},
					Link: filterOverview.Link{
						Label: "Edit",
						URL:   "/filters/12349876/dimensions/age",
					},
				},
			}
			fop := CreateFilterOverview(req, dimensions, datasetDimension, f, dst, filterID, datasetID, releaseDate, apiRouterVersion, lang)
			So(fop, ShouldResemble, expectedFop)
		})
	})
}

func TestUnitMapper(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	Convey("test CreatePreviewPage correctly maps to previewPage frontend model", t, func() {
		dimensions := getTestDimensions()
		filter := getTestFilter()
		dataset := getTestDataset()

		pp := CreatePreviewPage(req, dimensions, filter, dataset, filter.FilterID, "12345", "11-11-1992", "/v1", false, "en")
		So(pp.SearchDisabled, ShouldBeFalse)
		So(pp.Breadcrumb, ShouldHaveLength, 4)
		So(pp.Breadcrumb[0].Title, ShouldEqual, dataset.Title)
		So(pp.Breadcrumb[0].URI, ShouldEqual, "/datasets//editions")
		So(pp.Breadcrumb[1].Title, ShouldEqual, "5678")
		So(pp.Breadcrumb[1].URI, ShouldEqual, "/datasets/1234/editions/5678/versions/1")
		So(pp.Breadcrumb[2].Title, ShouldEqual, "Filter options")
		So(pp.Breadcrumb[2].URI, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions")
		So(pp.Breadcrumb[3].Title, ShouldEqual, "Preview")
		So(pp.Breadcrumb[3].URI, ShouldEqual, "")
		So(pp.Data.FilterID, ShouldEqual, filter.Links.FilterBlueprint.ID)
		if pp.Data.Downloads[0].Extension == "csv" {
			So(pp.Data.Downloads[0].Extension, ShouldEqual, "csv")
			So(pp.Data.Downloads[0].Size, ShouldEqual, "362783")
			So(pp.Data.Downloads[0].URI, ShouldEqual, "/")
		} else {
			So(pp.Data.Downloads[0].Extension, ShouldEqual, "xls")
			So(pp.Data.Downloads[0].Size, ShouldEqual, "373929")
			So(pp.Data.Downloads[0].URI, ShouldEqual, "/")
		}

		for i, dim := range pp.Data.Dimensions {
			So(dim.Values, ShouldResemble, dimensions[i].Values)
			So(dim.Name, ShouldEqual, dimensions[i].Name)
		}
	})

	Convey("test CreateListSelector page... ", t, func() {
		Convey("correctly maps to listSelector frontend model", func() {
			allValues := dataset.Options{
				Items: []dataset.Option{
					{
						Label:  "Feb-10",
						Option: "abcdefg",
					},
					{
						Label:  "Mar-10",
						Option: "38jd83ik",
					},
					{
						Label:  "Apr-10",
						Option: "13984094",
					},
				},
			}
			d := getTestDataset()
			selectedValues := []filter.DimensionOption{
				{
					Option: "38jd83ik",
				},
				{
					Option: "bcdefg",
				},
			}

			filter := getTestFilter()

			p := CreateListSelectorPage(req, "time", selectedValues, allValues, filter, d, dataset.VersionDimensions{}, "12345", "11-11-1992", "/v1", "en")
			So(p.Data.Title, ShouldEqual, "Time")
			So(p.SearchDisabled, ShouldBeTrue)
			So(p.FilterID, ShouldEqual, filter.FilterID)

			So(p.Breadcrumb, ShouldHaveLength, 4)
			So(p.Breadcrumb[0].Title, ShouldEqual, d.Title)
			So(p.Breadcrumb[0].URI, ShouldEqual, "/datasets//editions")
			So(p.Breadcrumb[1].Title, ShouldEqual, "5678")
			So(p.Breadcrumb[1].URI, ShouldEqual, "/datasets/1234/editions/5678/versions/1")
			So(p.Breadcrumb[2].Title, ShouldEqual, "Filter options")
			So(p.Breadcrumb[2].URI, ShouldEqual, "/filters/"+filter.Links.FilterBlueprint.ID+"/dimensions")
			So(p.Breadcrumb[3].Title, ShouldEqual, "Time")
			So(p.Breadcrumb[3].URI, ShouldEqual, "")
			So(p.Data.AddFromRange.Label, ShouldEqual, "add time range")
			So(p.Data.AddFromRange.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/time")
			So(p.Data.SaveAndReturn.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions")
			So(p.Data.Cancel.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions")
			So(p.Data.AddAllInRange.Label, ShouldEqual, "All times")
			So(p.Data.RangeData.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/time/list")
			So(p.Data.RemoveAll.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/time/remove-all")
			So(p.Data.RangeData.Values, ShouldHaveLength, 3)
			So(p.Data.RangeData.Values[0].Label, ShouldEqual, "Feb-10")
			So(p.Data.RangeData.Values[0].IsSelected, ShouldBeFalse)
			So(p.Data.RangeData.Values[1].Label, ShouldEqual, "Mar-10")
			So(p.Data.RangeData.Values[1].IsSelected, ShouldBeTrue)
			So(p.Data.RangeData.Values[2].Label, ShouldEqual, "Apr-10")
			So(p.Data.RangeData.Values[2].IsSelected, ShouldBeFalse)
			So(p.Data.FiltersAmount, ShouldEqual, 2)
		})

		Convey("keeps the same order for the time values as provided by dataset API", func() {
			p := CreateListSelectorPage(req, "time", []filter.DimensionOption{}, dataset.Options{
				Items: []dataset.Option{
					{
						Label: "2013",
					},
					{
						Label: "2010",
					},
					{
						Label: "2009",
					},
					{
						Label: "2017",
					},
				},
			}, filter.Model{}, dataset.DatasetDetails{}, dataset.VersionDimensions{}, "1234", "today", "/v1", "en")

			So(len(p.Data.RangeData.Values), ShouldEqual, 4)

			So(p.Data.RangeData.Values[0].Label, ShouldEqual, "2013")
			So(p.Data.RangeData.Values[1].Label, ShouldEqual, "2010")
			So(p.Data.RangeData.Values[2].Label, ShouldEqual, "2009")
			So(p.Data.RangeData.Values[3].Label, ShouldEqual, "2017")
		})

		Convey("keeps the same order for the non time/age values as provided by dataset API", func() {
			p := CreateListSelectorPage(req, "geography", []filter.DimensionOption{}, dataset.Options{
				Items: []dataset.Option{
					{
						Label: "Wales",
					},
					{
						Label: "Scotland",
					},
					{
						Label: "England",
					},
					{
						Label: "Ireland",
					},
				},
			}, filter.Model{}, dataset.DatasetDetails{}, dataset.VersionDimensions{}, "1234", "today", "/v1", "en")

			So(len(p.Data.RangeData.Values), ShouldEqual, 4)

			So(p.Data.RangeData.Values[0].Label, ShouldEqual, "Wales")
			So(p.Data.RangeData.Values[1].Label, ShouldEqual, "Scotland")
			So(p.Data.RangeData.Values[2].Label, ShouldEqual, "England")
			So(p.Data.RangeData.Values[3].Label, ShouldEqual, "Ireland")
		})

	})
}

func getTestDimensions() []filter.ModelDimension {
	return []filter.ModelDimension{
		{
			Name:   "year",
			Values: []string{"2014"},
		},
		{
			Name:   "geography",
			Values: []string{"England and Wales", "Bristol"},
		},
		{
			Name:   "sex",
			Values: []string{"All persons"},
		},
		{
			Name:   "age-range",
			Values: []string{"0 - 92", "2 - 18", "18 - 65"},
		},
		{
			Name:   "time",
			Values: []string{"2002.10", "2009.08", "1996.08"},
		},
	}
}

// getTestDatasetTimeOptions returns 3 dataset Options for the time dimension, sorted in a non-chronological order
func getTestDatasetTimeOptions() dataset.Options {
	return dataset.Options{Items: []dataset.Option{
		{
			DimensionID: "time",
			Label:       "Apr-07",
			Links: dataset.Links{
				CodeList: dataset.Link{
					URL: "http://api.localhost:23200/v1/code-lists/mmm-yy",
					ID:  "mmm-yy",
				},
				Version: dataset.Link{
					URL: "http://api.localhost:23200/v1/datasets/cpih01/editions/time-series/versions/7",
					ID:  "7",
				},
				Code: dataset.Link{
					URL: "http://api.localhost:23200/v1/code-lists/mmm-yy/codes/Month",
					ID:  "Month",
				},
			},
			Option: "Apr-07",
		},
		{
			DimensionID: "time",
			Label:       "Apr-05",
			Links: dataset.Links{
				CodeList: dataset.Link{
					URL: "http://api.localhost:23200/v1/code-lists/mmm-yy",
					ID:  "mmm-yy",
				},
				Version: dataset.Link{
					URL: "http://api.localhost:23200/v1/datasets/cpih01/editions/time-series/versions/7",
					ID:  "7",
				},
				Code: dataset.Link{
					URL: "http://api.localhost:23200/v1/code-lists/mmm-yy/codes/Month",
					ID:  "Month",
				},
			},
			Option: "Apr-05",
		},
		{
			DimensionID: "time",
			Label:       "Apr-06",
			Links: dataset.Links{
				CodeList: dataset.Link{
					URL: "http://api.localhost:23200/v1/code-lists/mmm-yy",
					ID:  "mmm-yy",
				},
				Version: dataset.Link{
					URL: "http://api.localhost:23200/v1/datasets/cpih01/editions/time-series/versions/7",
					ID:  "7",
				},
				Code: dataset.Link{
					URL: "http://api.localhost:23200/v1/code-lists/mmm-yy/codes/Month",
					ID:  "Month",
				},
			},
			Option: "Apr-06",
		},
		{
			DimensionID: "time",
			Label:       "Jun-05",
			Links: dataset.Links{
				CodeList: dataset.Link{
					URL: "http://api.localhost:23200/v1/code-lists/mmm-yy",
					ID:  "mmm-yy",
				},
				Version: dataset.Link{
					URL: "http://api.localhost:23200/v1/datasets/cpih01/editions/time-series/versions/7",
					ID:  "7",
				},
				Code: dataset.Link{
					URL: "http://api.localhost:23200/v1/code-lists/mmm-yy/codes/Month",
					ID:  "Month",
				},
			},
			Option: "Jun-05",
		},
		{
			DimensionID: "time",
			Label:       "May-05",
			Links: dataset.Links{
				CodeList: dataset.Link{
					URL: "http://api.localhost:23200/v1/code-lists/mmm-yy",
					ID:  "mmm-yy",
				},
				Version: dataset.Link{
					URL: "http://api.localhost:23200/v1/datasets/cpih01/editions/time-series/versions/7",
					ID:  "7",
				},
				Code: dataset.Link{
					URL: "http://api.localhost:23200/v1/code-lists/mmm-yy/codes/Month",
					ID:  "Month",
				},
			},
			Option: "May-05",
		},
	}}
}

func getTestDatasetDimensions() []dataset.VersionDimension {
	return []dataset.VersionDimension{
		{
			Name: "year",
		},
		{
			Name:  "geography",
			Label: "Geographic Areas",
		},
		{
			Name: "sex",
		},
		{
			Name:  "age-range",
			Label: "Age",
		},
		{
			Name:        "age",
			Description: "Description of the Age Dimension",
		},
		{
			Name: "time",
		},
	}
}

func getTestFilter() filter.Model {
	return filter.Model{
		FilterID:  "12349876",
		Edition:   "12345",
		DatasetID: "849209",
		Version:   "2017",
		Links: filter.Links{
			Version: filter.Link{
				HRef: "/v1/datasets/1234/editions/5678/versions/1",
			},
			FilterBlueprint: filter.Link{
				ID: "12349876",
			},
		},
		Downloads: map[string]filter.Download{
			"csv": {
				Size: "362783",
				URL:  "/",
			},
			"xls": {
				Size: "373929",
				URL:  "/",
			},
		},
	}
}

func getTestDataset() dataset.DatasetDetails {
	return dataset.DatasetDetails{
		NextRelease: "17 January 2018",
		Contacts: &[]dataset.Contact{
			{
				Name:      "Matt Rout",
				Telephone: "07984593234",
				Email:     "matt@gmail.com",
			},
		},
		Title: "Small Area Population Estimates",
	}
}

func TestUnitMapCookiesPreferences(t *testing.T) {
	req := httptest.NewRequest("", "/", nil)
	pageModel := model.Page{
		CookiesPreferencesSet: false,
		CookiesPolicy: model.CookiesPolicy{
			Essential: false,
			Usage:     false,
		},
	}

	Convey("maps cookies preferences cookie data to page model correctly", t, func() {
		So(pageModel.CookiesPreferencesSet, ShouldEqual, false)
		So(pageModel.CookiesPolicy.Essential, ShouldEqual, false)
		So(pageModel.CookiesPolicy.Usage, ShouldEqual, false)
		req.AddCookie(&http.Cookie{Name: "cookies_preferences_set", Value: "true"})
		req.AddCookie(&http.Cookie{Name: "cookies_policy", Value: "%7B%22essential%22%3Atrue%2C%22usage%22%3Atrue%7D"})
		mapCookiePreferences(req, &pageModel.CookiesPreferencesSet, &pageModel.CookiesPolicy)
		So(pageModel.CookiesPreferencesSet, ShouldEqual, true)
		So(pageModel.CookiesPolicy.Essential, ShouldEqual, true)
		So(pageModel.CookiesPolicy.Usage, ShouldEqual, true)
	})
}

func TestCreateHierarchyPage(t *testing.T) {
	Convey("CreateHierarchyPage maps the hierarchy data to the page model correctly", t, func() {
		var testHierarchyPage hierarchyModel.Page
		testHierarchyPage.Page = model.Page{
			Type:      "type",
			DatasetId: "datasetID",
			HasJSONLD: false,
			FeatureFlags: model.FeatureFlags{
				HideCookieBanner: false,
			},
			CookiesPolicy: model.CookiesPolicy{
				Essential: true,
				Usage:     false,
			},
			CookiesPreferencesSet:            false,
			BetaBannerEnabled:                true,
			SiteDomain:                       "",
			SearchDisabled:                   true,
			URI:                              "/",
			Taxonomy:                         nil,
			ReleaseDate:                      "",
			IsInFilterBreadcrumb:             true,
			Language:                         "en",
			IncludeAssetsIntegrityAttributes: false,
			DatasetTitle:                     "datasetTitle",
			Metadata: model.Metadata{
				Title:       "Filter Options - DatasetTitle",
				Description: "",
				ServiceName: "",
				Keywords:    nil,
			},
			Breadcrumb: []model.TaxonomyNode{
				{Title: "datasetTitle", URI: "/datasets/datasetID/editions"},
				{Title: "5678", URI: "/v1/datasets/1234/editions/5678/versions/1"},
				{Title: "Filter options", URI: "/filters/12349876/dimensions"},
				{Title: "DatasetTitle", URI: ""},
			},
			PatternLibraryAssetsPath: "",
		}

		testHierarchyPage.Data = hierarchyModel.Hierarchy{
			Title: "DatasetTitle",
			SaveAndReturn: hierarchyModel.Link{
				URL:   "//update",
				Label: "",
			},
			Cancel: hierarchyModel.Link{
				URL:   "/filters/12349876/dimensions",
				Label: "",
			},
			FiltersAmount: "",
			FilterList:    nil,
			AddAllFilters: hierarchyModel.AddAll{
				Amount: "0",
				URL:    "//add-all",
			},
			FiltersAdded: []hierarchyModel.Filter{
				{
					Label:     "This is option 1",
					RemoveURL: "//remove/op1",
					ID:        "op1",
				},
			},
			RemoveAll: hierarchyModel.Link{
				URL: "//remove-all",
			},
			DimensionName: "DatasetTitle",
			SearchURL:     "/filters/12349876/dimensions/datasetTitle/search",
		}
		testHierarchyPage.FilterID = "12349876"

		testSelectedOptions := map[string]string{"op1": "This is option 1"}

		testVersion := dataset.Version{
			ReleaseDate: "testRelease",
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

		req := httptest.NewRequest("", "/", nil)
		filterModel := getTestFilter()
		apiRouterVersion := "v1"
		lang := dprequest.DefaultLang
		hierarchyPageModel := CreateHierarchyPage(req, hierarchy.Model{}, testDatasetDetails, filterModel, testSelectedOptions, testVersionDimensions, testDatasetDetails.Title, req.URL.Path, testDatasetDetails.ID, testVersion.ReleaseDate, apiRouterVersion, lang)
		So(hierarchyPageModel, ShouldResemble, testHierarchyPage)
	})
}

// getExpectedTimePage returns the timeModel.Page model that would be generated
// from the options returned by getTestDatasetTimeOptions and no selected options, keeping the original values order
func getExpectedTimePage(datasetID, filterID, lang string) timeModel.Page {
	p := timeModel.Page{
		Page: model.Page{},
		Data: timeModel.Data{
			LatestTime: timeModel.Value{
				Month:      "April",
				Year:       "2007",
				Option:     "Apr-07",
				IsSelected: false,
			},
			FirstTime: timeModel.Value{
				Month:      "April",
				Year:       "2005",
				Option:     "Apr-05",
				IsSelected: false,
			},
			Values: []timeModel.Value{
				{Month: "April", Year: "2007", Option: "Apr-07", IsSelected: false},
				{Month: "April", Year: "2005", Option: "Apr-05", IsSelected: false},
				{Month: "April", Year: "2006", Option: "Apr-06", IsSelected: false},
				{Month: "June", Year: "2005", Option: "Jun-05", IsSelected: false},
				{Month: "May", Year: "2005", Option: "May-05", IsSelected: false},
			},
			Months:     []string{"Select", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
			Years:      []string{"Select", "2005", "2006", "2007"},
			FormAction: timeModel.Link{Label: "", URL: "/filters/12349876/dimensions/time/update"},
			Type:       "month",
			GroupedSelection: timeModel.GroupedSelection{
				Months: []timeModel.Month{
					{
						Name:       "January",
						IsSelected: false,
					},
					{
						Name:       "February",
						IsSelected: false,
					},
					{
						Name:       "March",
						IsSelected: false,
					},
					{
						Name:       "April",
						IsSelected: false,
					},
					{
						Name:       "May",
						IsSelected: false,
					},
					{
						Name:       "June",
						IsSelected: false,
					},
					{
						Name:       "July",
						IsSelected: false,
					},
					{
						Name:       "August",
						IsSelected: false,
					},
					{
						Name:       "September",
						IsSelected: false,
					},
					{
						Name:       "October",
						IsSelected: false,
					},
					{
						Name:       "November",
						IsSelected: false,
					},
					{
						Name:       "December",
						IsSelected: false,
					},
				},
			},
		},
	}
	p.FilterID = filterID
	p.DatasetId = datasetID
	p.URI = "/"
	p.DatasetTitle = "Small Area Population Estimates"
	p.IsInFilterBreadcrumb = true
	p.Metadata = model.Metadata{Title: "Time"}
	p.SearchDisabled = true
	p.BetaBannerEnabled = true
	p.Language = lang
	p.CookiesPolicy = model.CookiesPolicy{Essential: true}
	p.Breadcrumb = []model.TaxonomyNode{
		{
			Title: "Small Area Population Estimates",
			URI:   "/datasets//editions",
		},
		{
			Title: "5678",
			URI:   "/v1/datasets/1234/editions/5678/versions/1",
		},
		{
			Title: "Filter options",
			URI:   "/filters/12349876/dimensions",
		},
		{
			Title: "Time",
		},
	}
	return p
}

func TestIsTimeRange(t *testing.T) {
	t0, _ := time.Parse("Jan-06", "Jan-21")
	t1, _ := time.Parse("Jan-06", "Feb-21")
	t2, _ := time.Parse("Jan-06", "Mar-21")
	t3, _ := time.Parse("Jan-06", "Apr-21")
	t4, _ := time.Parse("Jan-06", "May-21")
	t5, _ := time.Parse("Jan-06", "Jun-21")
	t6, _ := time.Parse("Jan-06", "Jul-21")
	t7, _ := time.Parse("Jan-06", "Aug-21")
	t8, _ := time.Parse("Jan-06", "Sep-21")
	t9, _ := time.Parse("Jan-06", "Oct-21")
	sortedTimes := []time.Time{t0, t1, t2, t3, t4, t5, t6, t7, t8, t9}

	Convey("Given an empty array of selected values", t, func() {
		selVals := []filter.DimensionOption{}
		Convey("Then isTimeRange returns false", func() {
			So(isTimeRange(sortedTimes, selVals), ShouldBeFalse)
		})
	})

	Convey("Given a single selected value", t, func() {
		selVals := []filter.DimensionOption{
			{Option: "Apr-21"},
		}
		Convey("Then isTimeRange returns false", func() {
			So(isTimeRange(sortedTimes, selVals), ShouldBeFalse)
		})
	})

	Convey("Given two chronologically consecutive selected values", t, func() {
		selVals := []filter.DimensionOption{
			{Option: "May-21"},
			{Option: "Apr-21"},
		}
		Convey("Then isTimeRange returns true", func() {
			So(isTimeRange(sortedTimes, selVals), ShouldBeTrue)
		})
	})

	Convey("Given two chronologically discontinuous selected values", t, func() {
		selVals := []filter.DimensionOption{
			{Option: "Sep-21"},
			{Option: "Apr-21"},
		}
		Convey("Then isTimeRange returns false", func() {
			So(isTimeRange(sortedTimes, selVals), ShouldBeFalse)
		})
	})

	Convey("Given six chronologically consecutive selected values", t, func() {
		selVals := []filter.DimensionOption{
			{Option: "May-21"},
			{Option: "Apr-21"},
			{Option: "Jun-21"},
			{Option: "Feb-21"},
			{Option: "Jul-21"},
			{Option: "Mar-21"},
		}
		Convey("Then isTimeRange returns true", func() {
			So(isTimeRange(sortedTimes, selVals), ShouldBeTrue)
		})
	})

	Convey("Given 2 groups of 3 chronologically consecutive selected values", t, func() {
		selVals := []filter.DimensionOption{
			{Option: "May-21"},
			{Option: "Jan-21"},
			{Option: "Jun-21"},
			{Option: "Feb-21"},
			{Option: "Jul-21"},
			{Option: "Mar-21"},
		}
		Convey("Then isTimeRange returns false", func() {
			So(isTimeRange(sortedTimes, selVals), ShouldBeFalse)
		})
	})

	Convey("Given 2 selected values with the wrong format", t, func() {
		selVals := []filter.DimensionOption{
			{Option: "wrong1"},
			{Option: "wrong2"},
		}
		Convey("Then isTimeRange returns false", func() {
			So(isTimeRange(sortedTimes, selVals), ShouldBeFalse)
		})
	})
}

func TestCreateTimePage(t *testing.T) {
	req := httptest.NewRequest("", "/", nil)
	datasetID := "cpih01"
	apiRouterVersion := "v1"
	lang := dprequest.DefaultLang

	Convey("Given a valid request and all empty values, then CreateTimePage generates the expected timeModel page", t, func() {
		expected := timeModel.Page{}
		expected.BetaBannerEnabled = true
		expected.CookiesPolicy = model.CookiesPolicy{Essential: true}

		timeModelPage, err := CreateTimePage(req, filter.Model{}, dataset.DatasetDetails{}, dataset.Version{}, dataset.Options{}, []filter.DimensionOption{}, dataset.VersionDimensions{}, "", "", "")
		So(err, ShouldBeNil)
		So(timeModelPage, ShouldResemble, expected)
	})

	Convey("Given a valid request with no selected options, then CreateTimePage generates the expected timeModel page", t, func() {
		filterModel := getTestFilter()
		datasetDetails := getTestDataset()
		datasetVersion := dataset.Version{}
		options := getTestDatasetTimeOptions()
		selectedOptions := []filter.DimensionOption{}
		versionDimensions := dataset.VersionDimensions{Items: getTestDatasetDimensions()}

		expected := getExpectedTimePage(datasetID, filterModel.FilterID, lang)
		timeModelPage, err := CreateTimePage(req, filterModel, datasetDetails, datasetVersion, options, selectedOptions, versionDimensions, datasetID, apiRouterVersion, lang)
		So(err, ShouldBeNil)
		So(timeModelPage, ShouldResemble, expected)
	})

	Convey("Given a valid request with a selected option, then CreateTimePage generates the expected single timeModel page", t, func() {
		filterModel := getTestFilter()
		datasetDetails := getTestDataset()
		datasetVersion := dataset.Version{}
		options := getTestDatasetTimeOptions()
		selectedOptions := []filter.DimensionOption{
			{Option: "Apr-05"},
		}
		versionDimensions := dataset.VersionDimensions{Items: getTestDatasetDimensions()}

		expected := getExpectedTimePage(datasetID, filterModel.FilterID, lang)
		expected.Data.Values[1] = timeModel.Value{Month: "April", Year: "2005", Option: "Apr-05", IsSelected: true}
		expected.Data.CheckedRadio = "single"
		expected.Data.SelectedStartMonth = "April"
		expected.Data.SelectedStartYear = "2005"
		expected.Data.GroupedSelection.Months[3] = timeModel.Month{
			Name:       "April",
			IsSelected: true,
		}
		expected.Data.GroupedSelection.YearStart = "2005"
		expected.Data.GroupedSelection.YearEnd = "2005"

		timeModelPage, err := CreateTimePage(req, filterModel, datasetDetails, datasetVersion, options, selectedOptions, versionDimensions, datasetID, apiRouterVersion, lang)
		So(err, ShouldBeNil)
		So(timeModelPage, ShouldResemble, expected)
	})

	Convey("Given a valid request with the latest time option selected, even if it's not the last item in the options list, then CreateTimePage generates the expected latest timeModel page", t, func() {
		filterModel := getTestFilter()
		datasetDetails := getTestDataset()
		datasetVersion := dataset.Version{}
		options := getTestDatasetTimeOptions()
		selectedOptions := []filter.DimensionOption{
			{Option: "Apr-07"},
		}
		versionDimensions := dataset.VersionDimensions{Items: getTestDatasetDimensions()}

		expected := getExpectedTimePage(datasetID, filterModel.FilterID, lang)
		expected.Data.Values[0] = timeModel.Value{Month: "April", Year: "2007", Option: "Apr-07", IsSelected: true}
		expected.Data.CheckedRadio = "latest"
		expected.Data.GroupedSelection.Months[3] = timeModel.Month{
			Name:       "April",
			IsSelected: true,
		}
		expected.Data.GroupedSelection.YearStart = "2007"
		expected.Data.GroupedSelection.YearEnd = "2007"

		timeModelPage, err := CreateTimePage(req, filterModel, datasetDetails, datasetVersion, options, selectedOptions, versionDimensions, datasetID, apiRouterVersion, lang)
		So(err, ShouldBeNil)
		So(timeModelPage, ShouldResemble, expected)
	})

	Convey("Given a valid request with two non-chronologically-consecutive selected options, then CreateTimePage generates the expected list timeModel page", t, func() {
		filterModel := getTestFilter()
		datasetDetails := getTestDataset()
		datasetVersion := dataset.Version{}
		options := getTestDatasetTimeOptions()
		selectedOptions := []filter.DimensionOption{
			{Option: "Apr-05"},
			{Option: "Apr-07"},
		}
		versionDimensions := dataset.VersionDimensions{Items: getTestDatasetDimensions()}

		expected := getExpectedTimePage(datasetID, filterModel.FilterID, lang)
		expected.Data.Values[0] = timeModel.Value{Month: "April", Year: "2007", Option: "Apr-07", IsSelected: true}
		expected.Data.Values[1] = timeModel.Value{Month: "April", Year: "2005", Option: "Apr-05", IsSelected: true}
		expected.Data.CheckedRadio = "list"
		expected.Data.GroupedSelection.Months[3] = timeModel.Month{
			Name:       "April",
			IsSelected: true,
		}
		expected.Data.GroupedSelection.YearStart = "2005"
		expected.Data.GroupedSelection.YearEnd = "2007"

		timeModelPage, err := CreateTimePage(req, filterModel, datasetDetails, datasetVersion, options, selectedOptions, versionDimensions, datasetID, apiRouterVersion, lang)
		So(err, ShouldBeNil)
		So(timeModelPage, ShouldResemble, expected)
	})

	Convey("Given a valid request with three chronologically consecutive selected options, then CreateTimePage generates the expected range timeModel page", t, func() {
		filterModel := getTestFilter()
		datasetDetails := getTestDataset()
		datasetVersion := dataset.Version{}
		options := getTestDatasetTimeOptions()
		selectedOptions := []filter.DimensionOption{
			{Option: "Jun-05"},
			{Option: "Apr-05"},
			{Option: "May-05"},
		}
		versionDimensions := dataset.VersionDimensions{Items: getTestDatasetDimensions()}

		expected := getExpectedTimePage(datasetID, filterModel.FilterID, lang)
		expected.Data.Values[1] = timeModel.Value{Month: "April", Year: "2005", Option: "Apr-05", IsSelected: true}
		expected.Data.Values[4] = timeModel.Value{Month: "May", Year: "2005", Option: "May-05", IsSelected: true}
		expected.Data.Values[3] = timeModel.Value{Month: "June", Year: "2005", Option: "Jun-05", IsSelected: true}
		expected.Data.CheckedRadio = "range"
		expected.Data.GroupedSelection.Months[3] = timeModel.Month{
			Name:       "April",
			IsSelected: true,
		}
		expected.Data.GroupedSelection.Months[4] = timeModel.Month{
			Name:       "May",
			IsSelected: true,
		}
		expected.Data.GroupedSelection.Months[5] = timeModel.Month{
			Name:       "June",
			IsSelected: true,
		}
		expected.Data.SelectedStartMonth = "April"
		expected.Data.SelectedStartYear = "2005"
		expected.Data.SelectedEndMonth = "June"
		expected.Data.SelectedEndYear = "2005"
		expected.Data.GroupedSelection.YearStart = "2005"
		expected.Data.GroupedSelection.YearEnd = "2005"

		timeModelPage, err := CreateTimePage(req, filterModel, datasetDetails, datasetVersion, options, selectedOptions, versionDimensions, datasetID, apiRouterVersion, lang)
		So(err, ShouldBeNil)
		So(timeModelPage, ShouldResemble, expected)
	})

	Convey("Given an invalid URL link, then CreateTimePage returns the expected error", t, func() {
		filterModel := getTestFilter()
		filterModel.Links.Version.HRef = "invalid%url"
		datasetDetails := getTestDataset()
		datasetVersion := dataset.Version{}
		options := getTestDatasetTimeOptions()
		selectedOptions := []filter.DimensionOption{}
		versionDimensions := dataset.VersionDimensions{Items: getTestDatasetDimensions()}

		_, err := CreateTimePage(req, filterModel, datasetDetails, datasetVersion, options, selectedOptions, versionDimensions, datasetID, apiRouterVersion, lang)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "parse \"invalid%url\": invalid URL escape \"%ur\"")
	})

	Convey("Given that one of the dimension options has the wrong format, then CreateTimePage returns the expected error", t, func() {
		filterModel := getTestFilter()
		datasetDetails := getTestDataset()
		datasetVersion := dataset.Version{}
		options := getTestDatasetTimeOptions()
		options.Items[3].Label = "wrongFormat"
		selectedOptions := []filter.DimensionOption{}
		versionDimensions := dataset.VersionDimensions{Items: getTestDatasetDimensions()}

		_, err := CreateTimePage(req, filterModel, datasetDetails, datasetVersion, options, selectedOptions, versionDimensions, datasetID, apiRouterVersion, lang)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "parsing time \"wrongFormat\" as \"Jan-06\": cannot parse \"wrongFormat\" as \"Jan\"")
	})
}

// getTestDatasetAgeOptions returns an age dataset.Options for testing, with items sorted in a an order that is not from youngest to oldest
func getTestDatasetAgeOptions(hasPlusSign bool) dataset.Options {
	opts := dataset.Options{
		Count:      10,
		TotalCount: 10,
		Items: []dataset.Option{
			{DimensionID: "age", Label: "20", Option: "20"},
			{DimensionID: "age", Label: "40", Option: "40"},
			{DimensionID: "age", Label: "60", Option: "60"},
			{DimensionID: "age", Label: "80", Option: "80"},
			{DimensionID: "age", Label: "100", Option: "100"},
			{DimensionID: "age", Label: "10", Option: "10"},
			{DimensionID: "age", Label: "30", Option: "30"},
			{DimensionID: "age", Label: "50", Option: "50"},
			{DimensionID: "age", Label: "70", Option: "70"},
			{DimensionID: "age", Label: "90", Option: "90"},
		},
	}
	if hasPlusSign {
		opts.Items[4] = dataset.Option{DimensionID: "age", Label: "100+", Option: "100+"}
	}
	return opts
}

// getExpectedEmptyPageModel returns the age.Page model that would be generated from all-empty values
func getExpectedEmptyPageModel() age.Page {
	p := age.Page{
		Page: model.Page{
			Breadcrumb: []model.TaxonomyNode{
				{
					URI:      "/datasets//editions",
					Type:     "",
					Children: []model.TaxonomyNode(nil),
				},
				{},
				{
					Title:    "Filter options",
					URI:      "/filters//dimensions",
					Type:     "",
					Children: []model.TaxonomyNode(nil),
				},
				{
					Title: "Age",
					URI:   "", Type: "",
					Children: []model.TaxonomyNode(nil),
				},
			},
			IsInFilterBreadcrumb: true,
			Metadata: model.Metadata{
				Title: "Age",
			},
			SearchDisabled:    true,
			BetaBannerEnabled: true,
			CookiesPolicy: model.CookiesPolicy{
				Essential: true,
			},
		},
		Data: age.Data{
			CheckedRadio: "range",
			FormAction: age.Link{
				Label: "",
				URL:   "/filters//dimensions/age/update",
			},
		},
	}
	p.URI = "/"
	return p
}

func TestCreateAgePage(t *testing.T) {

	apiRouterVersion := "v1"
	datasetID := "cpih01"
	lang := dprequest.DefaultLang
	filterID := "12349876"

	Convey("Given a valid request, empty values for filter, dataset, options and selected options", t, func() {
		req := httptest.NewRequest("", "/", nil)
		filterModel := filter.Model{}
		datasetDetails := dataset.DatasetDetails{}
		datasetVersion := dataset.Version{}
		options := dataset.Options{}
		dimensionOptions := filter.DimensionOptions{}
		versionDimensions := dataset.VersionDimensions{}

		Convey("Then, CreateAgePage returns the expected age page without error", func() {
			ageModelPage, err := CreateAgePage(req, filterModel, datasetDetails, datasetVersion, options, dimensionOptions, versionDimensions, "", "", "")
			So(err, ShouldBeNil)
			expectedPageModel := getExpectedEmptyPageModel()
			So(ageModelPage, ShouldResemble, expectedPageModel)
		})

		Convey("Then, CreateAgePage with datasetID and language values returns the expected age page without error", func() {
			ageModelPage, err := CreateAgePage(req, filterModel, datasetDetails, datasetVersion, options, dimensionOptions, versionDimensions, datasetID, apiRouterVersion, lang)
			So(err, ShouldBeNil)

			expectedPageModel := getExpectedEmptyPageModel()
			expectedPageModel.DatasetId = datasetID
			expectedPageModel.Language = lang
			So(ageModelPage, ShouldResemble, expectedPageModel)
		})
	})

	Convey("Given a valid request, valid non-empty values for filter, dataset, datasetID and language", t, func() {
		req := httptest.NewRequest("", "/", nil)
		filterModel := getTestFilter()
		datasetDetails := getTestDataset()
		datasetVersion := dataset.Version{}
		versionDimensions := dataset.VersionDimensions{Items: getTestDatasetDimensions()}

		expectedPageModel := getExpectedEmptyPageModel()
		expectedPageModel.DatasetId = datasetID
		expectedPageModel.Language = lang
		expectedPageModel.DatasetTitle = datasetDetails.Title
		expectedPageModel.Metadata.Description = "Description of the Age Dimension"
		expectedPageModel.Breadcrumb[0].Title = datasetDetails.Title
		expectedPageModel.Breadcrumb[1] = model.TaxonomyNode{
			Title:    "5678",
			URI:      "/v1/datasets/1234/editions/5678/versions/1",
			Type:     "",
			Children: []model.TaxonomyNode(nil),
		}
		expectedPageModel.Breadcrumb[2] = model.TaxonomyNode{
			Title:    "Filter options",
			URI:      fmt.Sprintf("/filters/%s/dimensions", filterID),
			Type:     "",
			Children: []model.TaxonomyNode(nil),
		}
		expectedPageModel.Data.FormAction.URL = fmt.Sprintf("/filters/%s/dimensions/age/update", filterID)
		expectedPageModel.FilterID = filterID
		expectedPageModel.Data.Youngest = "10"
		expectedPageModel.Data.Oldest = "100+"

		Convey("Where one dataset option contains '+' and selected options is empty", func() {
			allOptions := getTestDatasetAgeOptions(true)
			selectedOptions := filter.DimensionOptions{}

			Convey("Then, the expected age page is generated without error, with options in the same order as provided by dataset API, and all of them marked as not selected", func() {
				expectedPageModel.Data.Ages = []age.Value{
					{Label: "20", Option: "20", IsSelected: false},
					{Label: "40", Option: "40", IsSelected: false},
					{Label: "60", Option: "60", IsSelected: false},
					{Label: "80", Option: "80", IsSelected: false},
					{Label: "100+", Option: "100+", IsSelected: false},
					{Label: "10", Option: "10", IsSelected: false},
					{Label: "30", Option: "30", IsSelected: false},
					{Label: "50", Option: "50", IsSelected: false},
					{Label: "70", Option: "70", IsSelected: false},
					{Label: "90", Option: "90", IsSelected: false},
				}
				ageModelPage, err := CreateAgePage(req, filterModel, datasetDetails, datasetVersion, allOptions, selectedOptions, versionDimensions, datasetID, apiRouterVersion, lang)
				So(err, ShouldBeNil)
				So(ageModelPage, ShouldResemble, expectedPageModel)
			})
		})

		Convey("Where one dataset option contains '+' and selected options is a subset of the dataset options", func() {
			allOptions := getTestDatasetAgeOptions(true)
			selectedOptions := filter.DimensionOptions{
				Count:      3,
				TotalCount: 3,
				Items:      []filter.DimensionOption{{Option: "60"}, {Option: "70"}, {Option: "100+"}},
			}

			Convey("Then, the expected age page is generated without error, with options in the same order as provided by dataset API, and only the selected ones marked as selected", func() {
				expectedPageModel.Data.Ages = []age.Value{
					{Label: "20", Option: "20", IsSelected: false},
					{Label: "40", Option: "40", IsSelected: false},
					{Label: "60", Option: "60", IsSelected: true},
					{Label: "80", Option: "80", IsSelected: false},
					{Label: "100+", Option: "100+", IsSelected: true},
					{Label: "10", Option: "10", IsSelected: false},
					{Label: "30", Option: "30", IsSelected: false},
					{Label: "50", Option: "50", IsSelected: false},
					{Label: "70", Option: "70", IsSelected: true},
					{Label: "90", Option: "90", IsSelected: false},
				}
				expectedPageModel.Data.CheckedRadio = "list"
				ageModelPage, err := CreateAgePage(req, filterModel, datasetDetails, datasetVersion, allOptions, selectedOptions, versionDimensions, datasetID, apiRouterVersion, lang)
				So(err, ShouldBeNil)
				So(ageModelPage, ShouldResemble, expectedPageModel)
			})
		})

		Convey("Where one dataset option contains '+' and and all dataset options are selected", func() {
			allOptions := getTestDatasetAgeOptions(true)
			selectedOptions := filter.DimensionOptions{
				Count:      3,
				TotalCount: 3,
				Items: []filter.DimensionOption{
					{Option: "10"}, {Option: "20"}, {Option: "30"}, {Option: "40"}, {Option: "50"},
					{Option: "60"}, {Option: "70"}, {Option: "80"}, {Option: "90"}, {Option: "100+"},
				},
			}
			Convey("Then, the expected age page is generated without error, with options in the same order as provided by dataset API, and all of them marked as selected", func() {
				expectedPageModel.Data.Ages = []age.Value{
					{Label: "20", Option: "20", IsSelected: true},
					{Label: "40", Option: "40", IsSelected: true},
					{Label: "60", Option: "60", IsSelected: true},
					{Label: "80", Option: "80", IsSelected: true},
					{Label: "100+", Option: "100+", IsSelected: true},
					{Label: "10", Option: "10", IsSelected: true},
					{Label: "30", Option: "30", IsSelected: true},
					{Label: "50", Option: "50", IsSelected: true},
					{Label: "70", Option: "70", IsSelected: true},
					{Label: "90", Option: "90", IsSelected: true},
				}
				expectedPageModel.Data.CheckedRadio = "range"
				expectedPageModel.Data.FirstSelected = "20"
				expectedPageModel.Data.LastSelected = "90"
				ageModelPage, err := CreateAgePage(req, filterModel, datasetDetails, datasetVersion, allOptions, selectedOptions, versionDimensions, datasetID, apiRouterVersion, lang)
				So(err, ShouldBeNil)
				So(ageModelPage, ShouldResemble, expectedPageModel)
			})
		})

		Convey("Where dataset options doesn't contain any value with '+'", func() {
			allOptions := getTestDatasetAgeOptions(false)
			selectedOptions := filter.DimensionOptions{}

			Convey("Then, the expected age page is generated without error, with options in the same order as provided by dataset API, and all of them marked as not selected", func() {
				expectedPageModel.Data.Ages = []age.Value{
					{Label: "20", Option: "20", IsSelected: false},
					{Label: "40", Option: "40", IsSelected: false},
					{Label: "60", Option: "60", IsSelected: false},
					{Label: "80", Option: "80", IsSelected: false},
					{Label: "100", Option: "100", IsSelected: false},
					{Label: "10", Option: "10", IsSelected: false},
					{Label: "30", Option: "30", IsSelected: false},
					{Label: "50", Option: "50", IsSelected: false},
					{Label: "70", Option: "70", IsSelected: false},
					{Label: "90", Option: "90", IsSelected: false},
				}
				expectedPageModel.Data.Oldest = "100"
				ageModelPage, err := CreateAgePage(req, filterModel, datasetDetails, datasetVersion, allOptions, selectedOptions, versionDimensions, datasetID, apiRouterVersion, lang)
				So(err, ShouldBeNil)
				So(ageModelPage, ShouldResemble, expectedPageModel)
			})
		})

		Convey("Where dataset options contains a nonnumerical value", func() {
			allOptions := dataset.Options{
				Count:      3,
				TotalCount: 3,
				Items: []dataset.Option{
					{DimensionID: "age", Label: "10", Option: "10"},
					{DimensionID: "age", Label: "nonnumerical", Option: "nonnumerical"},
					{DimensionID: "age", Label: "100+", Option: "100+"},
				},
			}
			selectedOptions := filter.DimensionOptions{}

			Convey("Then, the expected age page is generated without error, with options in the same order as provided by dataset API, and all of them marked as not selected and AllAgesOption set to the nonnumerical value", func() {
				expectedPageModel.Data.Ages = []age.Value{
					{Label: "10", Option: "10", IsSelected: false},
					{Label: "100+", Option: "100+", IsSelected: false},
				}
				expectedPageModel.Data.HasAllAges = true
				expectedPageModel.Data.AllAgesOption = "nonnumerical"
				ageModelPage, err := CreateAgePage(req, filterModel, datasetDetails, datasetVersion, allOptions, selectedOptions, versionDimensions, datasetID, apiRouterVersion, lang)
				So(err, ShouldBeNil)
				So(ageModelPage, ShouldResemble, expectedPageModel)
			})
		})
	})

	Convey("calling CreateAgePage with nil request results in an empty age page being generated and the expected error being returned", t, func() {
		ageModelPage, err := CreateAgePage(nil, filter.Model{}, dataset.DatasetDetails{}, dataset.Version{}, dataset.Options{}, filter.DimensionOptions{}, dataset.VersionDimensions{}, datasetID, apiRouterVersion, lang)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid request provided to CreateAgePage")
		So(ageModelPage, ShouldResemble, age.Page{})
	})

	Convey("calling CreateAGePage with an invalid filter version link results in an empty age page being generated and the expected error being returned", t, func() {
		req := httptest.NewRequest("", "/", nil)
		filterModel := filter.Model{
			Links: filter.Links{
				Version: filter.Link{
					HRef: "invalid%url",
				},
			},
		}

		ageModelPage, err := CreateAgePage(req, filterModel, dataset.DatasetDetails{}, dataset.Version{}, dataset.Options{}, filter.DimensionOptions{}, dataset.VersionDimensions{}, datasetID, apiRouterVersion, lang)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "parse \"invalid%url\": invalid URL escape \"%ur\"")
		So(ageModelPage, ShouldResemble, age.Page{})
	})
}
