package mapper

import (
	"testing"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/data"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitMapper(t *testing.T) {
	Convey("test CreateFilterOverview correctly maps item to filterOverview page model", t, func() {
		dimensions := getTestDimensions()
		filter := getTestFilter()
		dataset := getTestDataset()

		fop := CreateFilterOverview(dimensions, filter, dataset, filter.FilterID)
		So(fop.FilterID, ShouldEqual, filter.FilterID)
		So(fop.SearchDisabled, ShouldBeTrue)
		So(fop.Data.Dimensions, ShouldHaveLength, 4)
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
		So(fop.Breadcrumb[0].URI, ShouldEqual, "/datasets/"+filter.Dataset+"/editions/"+filter.Edition+"/versions/"+filter.Version)
		So(fop.Breadcrumb[1].Title, ShouldEqual, "Filter this dataset")
		So(fop.Metadata.Footer.Enabled, ShouldBeTrue)
		So(fop.Metadata.Footer.Contact, ShouldEqual, dataset.Contact)
		So(fop.Metadata.Footer.ReleaseDate, ShouldEqual, dataset.ReleaseDate)
		So(fop.Metadata.Footer.DatasetID, ShouldEqual, dataset.ID)
	})
}

func getTestDimensions() []data.Dimension {
	return []data.Dimension{
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
	}
}

func getTestFilter() data.Filter {
	return data.Filter{
		FilterID: "12349876",
		Edition:  "12345",
		Dataset:  "849209",
		Version:  "2017",
	}
}

func getTestDataset() data.Dataset {
	return data.Dataset{
		ID:          "849209",
		ReleaseDate: "17 January 2017",
		Contact: data.Contact{
			Name:      "Matt Rout",
			Telephone: "07984593234",
			Email:     "matt@gmail.com",
		},
		Title: "Small Area Population Estimates",
	}
}
