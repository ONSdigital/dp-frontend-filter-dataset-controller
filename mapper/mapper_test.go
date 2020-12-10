package mapper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-frontend-models/model"
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

		Convey("correctly orders the time values into ascending numeric order", func() {
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

			So(p.Data.RangeData.Values[0].Label, ShouldEqual, "2017")
			So(p.Data.RangeData.Values[1].Label, ShouldEqual, "2013")
			So(p.Data.RangeData.Values[2].Label, ShouldEqual, "2010")
			So(p.Data.RangeData.Values[3].Label, ShouldEqual, "2009")
		})

		Convey("correctly orders non time/age values alphabetically", func() {
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

			So(p.Data.RangeData.Values[0].Label, ShouldEqual, "England")
			So(p.Data.RangeData.Values[1].Label, ShouldEqual, "Ireland")
			So(p.Data.RangeData.Values[2].Label, ShouldEqual, "Scotland")
			So(p.Data.RangeData.Values[3].Label, ShouldEqual, "Wales")
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

func getTestDatasetOptions() dataset.Options {
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
	var testHierarchyPage hierarchyModel.Page
	testHierarchyPage.Page = model.Page{
		Type:      "type",
		DatasetId: "testDatasetID",
		HasJSONLD: false,
		FeatureFlags: model.FeatureFlags{
			HideCookieBanner: false,
		},
		CookiesPolicy: model.CookiesPolicy{
			Essential: true,
			Usage:     false,
		},
		CookiesPreferencesSet:            true,
		BetaBannerEnabled:                false,
		SiteDomain:                       "",
		SearchDisabled:                   false,
		URI:                              "",
		Taxonomy:                         nil,
		ReleaseDate:                      "",
		IsInFilterBreadcrumb:             false,
		Language:                         "en",
		IncludeAssetsIntegrityAttributes: false,
		DatasetTitle:                     "datasetTitle",
		Metadata: model.Metadata{
			Title:       "Filter Options - MyDimension",
			Description: "",
			ServiceName: "",
			Keywords:    nil,
		},
		Breadcrumb: []model.TaxonomyNode{
			{Title: ""},
			{Title: ""},
			{Title: ""},
			{Title: ""},
		},
		PatternLibraryAssetsPath: "",
	}

	testHierarchyPage.Data = hierarchyModel.Hierarchy{
		Title: "MyDimension",
		SaveAndReturn: hierarchyModel.Link{
			URL:   "filters/12345/dimensions/myDimension/update",
			Label: "",
		},
		Cancel: hierarchyModel.Link{
			URL:   "/filters/12345/dimensions",
			Label: "",
		},
		FiltersAmount: "",
		FilterList:    nil,
		AddAllFilters: hierarchyModel.AddAll{
			Amount: "0",
			URL:    "/filters/12345/dimensions/myDimension/add-all",
		},
		FiltersAdded: []hierarchyModel.Filter{
			{
				Label:     "",
				RemoveURL: "/filters/12345/dimensions/myDimension/remove/op1",
				ID:        "op1",
			},
			{
				Label:     "",
				RemoveURL: "/filters/12345/dimensions/myDimension/remove/op2",
				ID:        "op2",
			},
		},
		RemoveAll:     hierarchyModel.Link{},
		DimensionName: "MyDimension",
		SearchURL:     "/filters/12345/dimensions/myDimension/search",
	}
	testHierarchyPage.FilterID = ""
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
		options := getTestDatasetOptions()
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
