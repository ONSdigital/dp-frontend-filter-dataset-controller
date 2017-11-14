package mapper

import (
	"testing"

	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitMapper(t *testing.T) {
	Convey("test CreateFilterOverview correctly maps item to filterOverview page model", t, func() {
		dimensions := getTestDimensions()
		filter := getTestFilter()
		dataset := getTestDataset()

		fop := CreateFilterOverview(dimensions, filter, dataset, filter.FilterID, "12345", "11-11-1992")
		So(fop.FilterID, ShouldEqual, filter.FilterID)
		So(fop.SearchDisabled, ShouldBeTrue)
		So(fop.Data.Dimensions, ShouldHaveLength, 5)
		So(fop.Data.Dimensions[0].Filter, ShouldEqual, "Year")
		So(fop.Data.Dimensions[0].AddedCategories[0], ShouldEqual, "2014")
		So(fop.Data.Dimensions[0].Link.Label, ShouldEqual, "Filter")
		So(fop.Data.Dimensions[0].Link.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/year")
		So(fop.Data.Dimensions[1].Filter, ShouldEqual, "Geographic Areas")
		So(fop.Data.Dimensions[1].AddedCategories[0], ShouldEqual, "England and Wales")
		So(fop.Data.Dimensions[1].AddedCategories[1], ShouldEqual, "Bristol")
		So(fop.Data.Dimensions[1].Link.Label, ShouldEqual, "Filter")
		So(fop.Data.Dimensions[1].Link.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/geography")
		So(fop.Data.Dimensions[2].Filter, ShouldEqual, "Sex")
		So(fop.Data.Dimensions[2].AddedCategories[0], ShouldEqual, "All persons")
		So(fop.Data.Dimensions[2].Link.Label, ShouldEqual, "Filter")
		So(fop.Data.Dimensions[2].Link.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/sex")
		So(fop.Data.Dimensions[3].Filter, ShouldEqual, "Age")
		So(fop.Data.Dimensions[3].AddedCategories[0], ShouldEqual, "0 - 92")
		So(fop.Data.Dimensions[3].AddedCategories[1], ShouldEqual, "2 - 18")
		So(fop.Data.Dimensions[3].AddedCategories[2], ShouldEqual, "18 - 65")
		So(fop.Data.Dimensions[3].Link.Label, ShouldEqual, "Filter")
		So(fop.Data.Dimensions[3].Link.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/age-range")
		So(fop.Data.PreviewAndDownload.URL, ShouldEqual, "/filters/"+filter.FilterID)
		So(fop.Data.Cancel.URL, ShouldEqual, "/")
		So(fop.Breadcrumb, ShouldHaveLength, 2)
		So(fop.Breadcrumb[0].Title, ShouldEqual, dataset.Title)
		So(fop.Breadcrumb[1].Title, ShouldEqual, "Filter options")
		So(fop.Metadata.Footer.Enabled, ShouldBeTrue)
		So(fop.Metadata.Footer.Contact, ShouldEqual, dataset.Contacts[0].Name)
		So(fop.Metadata.Footer.ReleaseDate, ShouldEqual, "11-11-1992")
		So(fop.Metadata.Footer.DatasetID, ShouldEqual, "12345")
	})

	Convey("test CreatePreviewPage correctly maps to previewPage frontend model", t, func() {
		dimensions := getTestDimensions()
		filter := getTestFilter()
		dataset := getTestDataset()

		pp := CreatePreviewPage(dimensions, filter, dataset, filter.FilterID, "12345", "11-11-1992")
		So(pp.SearchDisabled, ShouldBeFalse)
		So(pp.Breadcrumb, ShouldHaveLength, 3)
		So(pp.Breadcrumb[0].Title, ShouldEqual, dataset.Title)
		So(pp.Breadcrumb[1].Title, ShouldEqual, "Filter options")
		So(pp.Breadcrumb[1].URI, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions")
		So(pp.Breadcrumb[2].Title, ShouldEqual, "Preview")
		So(pp.Data.FilterID, ShouldEqual, filter.Links.FilterBlueprint.ID)
		So(pp.Metadata.Footer.Enabled, ShouldBeTrue)
		So(pp.Metadata.Footer.Contact, ShouldEqual, dataset.Contacts[0].Name)
		So(pp.Metadata.Footer.ReleaseDate, ShouldEqual, "11-11-1992")
		So(pp.Metadata.Footer.DatasetID, ShouldEqual, "12345")
		So(pp.Data.Downloads[0].Extension, ShouldEqual, "csv")
		So(pp.Data.Downloads[0].Size, ShouldEqual, "362783")
		So(pp.Data.Downloads[0].URI, ShouldEqual, "/")

		for i, dim := range pp.Data.Dimensions {
			So(dim.Values, ShouldResemble, dimensions[i].Values)
			So(dim.Name, ShouldEqual, dimensions[i].Name)
		}
	})

	Convey("test CreateListSelector page correctly maps to listSelector frontend model", t, func() {
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
		dataset := getTestDataset()
		selectedValues := []filter.DimensionOption{
			{
				Option: "38jd83ik",
			},
			{
				Option: "bcdefg",
			},
		}

		filter := getTestFilter()

		p := CreateListSelectorPage("time", selectedValues, allValues, filter, dataset, "12345", "11-11-1992")
		So(p.Data.Title, ShouldEqual, "Time")
		So(p.SearchDisabled, ShouldBeTrue)
		So(p.FilterID, ShouldEqual, filter.FilterID)

		So(p.Breadcrumb, ShouldHaveLength, 3)
		So(p.Breadcrumb[0].Title, ShouldEqual, dataset.Title)
		So(p.Breadcrumb[1].Title, ShouldEqual, "Filter options")
		So(p.Breadcrumb[1].URI, ShouldEqual, "/filters/"+filter.Links.FilterBlueprint.ID+"/dimensions")
		So(p.Breadcrumb[2].Title, ShouldEqual, "Time")
		So(p.Data.AddFromRange.Label, ShouldEqual, "add time range")
		So(p.Data.AddFromRange.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/time")
		So(p.Data.SaveAndReturn.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions")
		So(p.Data.Cancel.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions")
		So(p.Data.AddAllInRange.Label, ShouldEqual, "All times")
		So(p.Data.RangeData.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/time/list")
		So(p.Data.RemoveAll.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/time/remove-all?selectorType=list")
		So(p.Data.RangeData.Values, ShouldHaveLength, 3)
		So(p.Data.RangeData.Values[0].Label, ShouldEqual, "Feb-10")
		So(p.Data.RangeData.Values[0].IsSelected, ShouldBeFalse)
		So(p.Data.RangeData.Values[1].Label, ShouldEqual, "Mar-10")
		So(p.Data.RangeData.Values[1].IsSelected, ShouldBeTrue)
		So(p.Data.RangeData.Values[2].Label, ShouldEqual, "Apr-10")
		So(p.Data.RangeData.Values[2].IsSelected, ShouldBeFalse)
		So(p.Data.FiltersAmount, ShouldEqual, 2)
		So(p.Metadata.Footer.Enabled, ShouldBeTrue)
		So(p.Metadata.Footer.Contact, ShouldEqual, dataset.Contacts[0].Name)
		So(p.Metadata.Footer.ReleaseDate, ShouldEqual, "11-11-1992")
		So(p.Metadata.Footer.DatasetID, ShouldEqual, "12345")
	})

	Convey("test CreateRangeSelectorPage successfully maps to a rangeSelector page model", t, func() {

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
		dataset := getTestDataset()
		selectedValues := []filter.DimensionOption{
			{
				Option: "38jd83ik",
			},
			{
				Option: "bcdefg",
			},
		}

		filter := getTestFilter()

		p := CreateRangeSelectorPage("time", selectedValues, allValues, filter, dataset, "12345", "11-11-1992")
		So(p.Data.Title, ShouldEqual, "Time")
		So(p.SearchDisabled, ShouldBeTrue)
		So(p.FilterID, ShouldEqual, filter.FilterID)

		So(p.Breadcrumb, ShouldHaveLength, 3)
		So(p.Breadcrumb[0].Title, ShouldEqual, dataset.Title)
		So(p.Breadcrumb[1].Title, ShouldEqual, "Filter options")
		So(p.Breadcrumb[1].URI, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions")
		So(p.Breadcrumb[2].Title, ShouldEqual, "Time")
		So(p.Data.AddFromList.Label, ShouldEqual, "Add time range")
		So(p.Data.AddFromList.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/time?selectorType=list")
		So(p.Data.SaveAndReturn.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions")
		So(p.Data.Cancel.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions")
		So(p.Data.AddAllInRange.Label, ShouldEqual, "All times")
		So(p.Data.RangeData.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/time/range")
		So(p.Data.RemoveAll.URL, ShouldEqual, "/filters/"+filter.FilterID+"/dimensions/time/remove-all")
		So(p.Metadata.Footer.Enabled, ShouldBeTrue)
		So(p.Metadata.Footer.Contact, ShouldEqual, dataset.Contacts[0].Name)
		So(p.Metadata.Footer.ReleaseDate, ShouldEqual, "11-11-1992")
		So(p.Metadata.Footer.DatasetID, ShouldEqual, "12345")
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

func getTestFilter() filter.Model {
	return filter.Model{
		FilterID: "12349876",
		Edition:  "12345",
		Dataset:  "849209",
		Version:  "2017",
		Links: filter.Links{
			Version: filter.Link{
				HRef: "/datasets/1234/editions/5678/versions/1",
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

func getTestDataset() dataset.Model {
	return dataset.Model{
		NextRelease: "17 January 2018",
		Contacts: []dataset.Contact{
			{
				Name:      "Matt Rout",
				Telephone: "07984593234",
				Email:     "matt@gmail.com",
			},
		},
		Title: "Small Area Population Estimates",
	}
}
