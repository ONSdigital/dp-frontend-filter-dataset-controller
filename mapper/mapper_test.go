package mapper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/hierarchy"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/dataset-filter/age"
	hierarchyModel "github.com/ONSdigital/dp-frontend-models/model/dataset-filter/hierarchy"
	timeModel "github.com/ONSdigital/dp-frontend-models/model/dataset-filter/time"
	dprequest "github.com/ONSdigital/dp-net/request"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitMapper(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	Convey("test CreateFilterOverview correctly maps item to filterOverview page model", t, func() {
		dimensions := getTestDimensions()
		datasetDimension := getTestDatasetDimensions()
		filter := getTestFilter()
		dst := getTestDataset()

		fop := CreateFilterOverview(req, dimensions, datasetDimension, filter, dst, filter.FilterID, "12345", "11-11-1992", "/v1", "en")
		So(fop.FilterID, ShouldEqual, filter.FilterID)
		So(fop.SearchDisabled, ShouldBeTrue)
		So(fop.Data.Dimensions, ShouldHaveLength, 5)
		So(fop.Data.Dimensions[0].Filter, ShouldEqual, "Year")
		So(fop.Data.Dimensions[0].AddedCategories[0], ShouldEqual, "2014")
		So(fop.Data.Dimensions[0].Link.Label, ShouldEqual, "Edit")
		So(fop.Data.Dimensions[0].Link.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/year")
		So(fop.Data.Dimensions[1].Filter, ShouldEqual, "Geographic Areas")
		So(fop.Data.Dimensions[1].AddedCategories[0], ShouldEqual, "England and Wales")
		So(fop.Data.Dimensions[1].AddedCategories[1], ShouldEqual, "Bristol")
		So(fop.Data.Dimensions[1].Link.Label, ShouldEqual, "Edit")
		So(fop.Data.Dimensions[1].Link.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/geography")
		So(fop.Data.Dimensions[2].Filter, ShouldEqual, "Sex")
		So(fop.Data.Dimensions[2].AddedCategories[0], ShouldEqual, "All persons")
		So(fop.Data.Dimensions[2].Link.Label, ShouldEqual, "Edit")
		So(fop.Data.Dimensions[2].Link.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/sex")
		So(fop.Data.Dimensions[3].Filter, ShouldEqual, "Age")
		So(fop.Data.Dimensions[3].AddedCategories[0], ShouldEqual, "0 - 92")
		So(fop.Data.Dimensions[3].AddedCategories[1], ShouldEqual, "2 - 18")
		So(fop.Data.Dimensions[3].AddedCategories[2], ShouldEqual, "18 - 65")
		So(fop.Data.Dimensions[3].Link.Label, ShouldEqual, "Edit")
		So(fop.Data.Dimensions[3].Link.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/age-range")
		So(fop.Data.PreviewAndDownload.URL, ShouldEqual, "/filters/"+filter.FilterID)
		So(fop.Data.Cancel.URL, ShouldEqual, "/")
		So(fop.Breadcrumb, ShouldHaveLength, 3)
		So(fop.Breadcrumb[0].Title, ShouldEqual, dst.Title)
		So(fop.Breadcrumb[0].URI, ShouldEqual, "/datasets//editions")
		So(fop.Breadcrumb[1].Title, ShouldEqual, "5678")
		So(fop.Breadcrumb[1].URI, ShouldEqual, "/datasets/1234/editions/5678/versions/1")
		So(fop.Breadcrumb[2].Title, ShouldEqual, "Filter options")
		So(fop.Breadcrumb[2].URI, ShouldEqual, "")
	})

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

func getTestDatasetTimeOptions() dataset.Options {
	return dataset.Options{Items: []dataset.Option{
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
		}}}
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
			URI:                              "",
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

func TestCreateTimePage(t *testing.T) {
	Convey("maps filter to page model correctly", t, func() {
		desiredPageModel := timeModel.Page{
			Page: model.Page{},
			Data: timeModel.Data{
				LatestTime:         timeModel.Value{},
				FirstTime:          timeModel.Value{},
				Values:             nil,
				Months:             []string{"Select", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
				Years:              []string{"Select", "2005", "2006", "2007"},
				CheckedRadio:       "",
				FormAction:         timeModel.Link{},
				SelectedStartMonth: "",
				SelectedStartYear:  "",
				SelectedEndMonth:   "",
				SelectedEndYear:    "",
				Type:               "",
				DatasetTitle:       "",
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
					YearStart: "",
					YearEnd:   "",
				},
			},
			FilterID: "",
		}
		req := httptest.NewRequest("", "/", nil)
		filterModel := getTestFilter()
		datasetDetails := getTestDataset()
		// Never actually used in the mapper but func requires it so leaving blank until needed in a test
		datasetVersion := dataset.Version{}
		options := getTestDatasetTimeOptions()
		dimensionOptions := []filter.DimensionOption{{}}
		versionDimensions := dataset.VersionDimensions{Items: getTestDatasetDimensions()}
		datasetID := "cpih01"
		apiRouterVersion := "v1"
		lang := dprequest.DefaultLang
		timeModelPage, err := CreateTimePage(req, filterModel, datasetDetails, datasetVersion, options, dimensionOptions, versionDimensions, datasetID, apiRouterVersion, lang)
		So(err, ShouldBeNil)
		So(timeModelPage.Data.GroupedSelection, ShouldResemble, desiredPageModel.Data.GroupedSelection)
		So(timeModelPage.Data.Months, ShouldResemble, desiredPageModel.Data.Months)
		So(timeModelPage.Data.Years, ShouldResemble, desiredPageModel.Data.Years)
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
	return age.Page{
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
