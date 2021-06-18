package helpers

import (
	"context"
	"net/url"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/filter"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitHelpers(t *testing.T) {
	ctx := context.Background()
	Convey("test ExtractDatasetInfoFromPath", t, func() {
		Convey("extracts datasetID, edition and version from path", func() {
			datasetID, edition, version, err := ExtractDatasetInfoFromPath(ctx, "/datasets/12345/editions/2016/versions/1")
			So(err, ShouldBeNil)
			So(datasetID, ShouldEqual, "12345")
			So(edition, ShouldEqual, "2016")
			So(version, ShouldEqual, "1")
		})

		Convey("returns an error if it is unable to extract the information", func() {
			datasetID, edition, version, err := ExtractDatasetInfoFromPath(ctx, "invalid")
			So(err.Error(), ShouldEqual, "unable to extract datasetID, edition and version from path: invalid")
			So(datasetID, ShouldEqual, "")
			So(edition, ShouldEqual, "")
			So(version, ShouldEqual, "")
		})
	})
}

func TestGetAPIRouterVersion(t *testing.T) {

	Convey("The api router version is correctly extracted from a valid API Router URL", t, func() {
		version, err := GetAPIRouterVersion("http://localhost:23200/v1")
		So(err, ShouldBeNil)
		So(version, ShouldEqual, "/v1")
	})

	Convey("An empty string version is extracted from a valid unversioned API Router URL", t, func() {
		version, err := GetAPIRouterVersion("http://localhost:23200")
		So(err, ShouldBeNil)
		So(version, ShouldEqual, "")
	})

	Convey("Extracting a version from an invalid API Router URL results in the parsing error being returned", t, func() {
		version, err := GetAPIRouterVersion("hello%goodbye")
		So(err, ShouldResemble, &url.Error{
			Op:  "parse",
			URL: "hello%goodbye",
			Err: url.EscapeError("%go"),
		})
		So(version, ShouldEqual, "")
	})
}

// TestStringInSlice tests the helper function TestStringInSlice
func TestStringInSlice(t *testing.T) {
	Convey("given a slice", t, func() {
		subjectSlice := []string{"foo", "Bar", "bAz", "quX", "QUUX", "COrGE", "gRaUlT", "GaRpLy", "WALdo", "frED", "pl ugh", "xyzzy1", "thud-", "", "gRaUlT"}
		Convey("and a string that the slice contains it should find said string", func() {
			testString := "foo"
			index, isFound := StringInSlice(testString, subjectSlice)
			So(isFound, ShouldEqual, true)
			So(index, ShouldEqual, 0)
			Convey("it should be case sensitive and only yield true for exact matches", func() {
				testString = "Bar"
				index, isFound = StringInSlice(testString, subjectSlice)
				So(isFound, ShouldEqual, true)
				So(index, ShouldEqual, 1)
				Convey("even if there is a random case difference in the middle of the string", func() {
					testString = "bAz"
					index, isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, true)
					So(index, ShouldEqual, 2)
				})
				Convey("or a space", func() {
					testString = "pl ugh"
					index, isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, true)
					So(index, ShouldEqual, 10)
				})
				Convey("or a numerical value represented as a char", func() {
					testString = "xyzzy1"
					index, isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, true)
					So(index, ShouldEqual, 11)
				})
				Convey("or a symbolic character represented as a char", func() {
					testString = "thud-"
					index, isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, true)
					So(index, ShouldEqual, 12)
				})
				Convey("or even a completely empty string", func() {
					testString = ""
					index, isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, true)
					So(index, ShouldEqual, 13)
				})
			})
			Convey("it should return the index of the first found position in the array if there are multiple", func() {
				testString = "gRaUlT"
				index, isFound = StringInSlice(testString, subjectSlice)
				So(isFound, ShouldEqual, true)
				So(index, ShouldEqual, 6)
			})
			Convey("it should be case sensitive and yield false for any other value", func() {
				Convey("like if the initial capital letter isn't considered", func() {
					testString = "bar"
					index, isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, false)
					So(index, ShouldEqual, -1)
				})
				Convey("or a random capital in the middle of the string isn't considered", func() {
					testString = "baz"
					index, isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, false)
					So(index, ShouldEqual, -1)
				})
				Convey("or a space in a string is missing", func() {
					testString = "plugh"
					index, isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, false)
					So(index, ShouldEqual, -1)
				})
				Convey("or a numeral represented as a string char is missing", func() {
					testString = "xyzzy"
					index, isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, false)
					So(index, ShouldEqual, -1)

				})
				Convey("or a symbol represented as a char is missing", func() {
					testString = "thud"
					index, isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, false)
					So(index, ShouldEqual, -1)
				})
				Convey("or if a space char slips in and shouldn't be there", func() {
					testString = " "
					index, isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, false)
					So(index, ShouldEqual, -1)
				})
			})
		})
	})
}

// TestCheckAllDimensionHaveAnOption test the helper function CheckAllDimensionsHaveAnOption
func TestCheckAllDimensionHaveAnOption(t *testing.T) {
	Convey("No dimesions provided return error", t, func() {
		testDimensions := []filter.ModelDimension{}

		sut, err := CheckAllDimensionsHaveAnOption(testDimensions)
		So(err.Error(), ShouldEqual, "no dimensions provided: []")
		So(sut, ShouldEqual, false)
	})

	Convey("All dimensions have an option return true", t, func() {
		testDimensions := []filter.ModelDimension{{
			Name:    "test",
			Options: []string{"option"},
		}}

		sut, err := CheckAllDimensionsHaveAnOption(testDimensions)
		So(err, ShouldBeNil)
		So(sut, ShouldEqual, true)
	})

	Convey("Dimension with no option selected return false", t, func() {
		testDimensions := []filter.ModelDimension{{
			Name:    "test",
			Options: []string{},
		}}

		sut, err := CheckAllDimensionsHaveAnOption(testDimensions)
		So(err, ShouldBeNil)
		So(sut, ShouldEqual, false)
	})

	Convey("Two dimensions with options selected return true", t, func() {
		testDimensions := []filter.ModelDimension{
			{
				Name:    "test",
				Options: []string{"option1"},
			},
			{
				Name:    "test2",
				Options: []string{"option2"},
			},
		}

		sut, err := CheckAllDimensionsHaveAnOption(testDimensions)
		So(err, ShouldBeNil)
		So(sut, ShouldEqual, true)
	})

	Convey("Two dimensions with no options selected return false", t, func() {
		testDimensions := []filter.ModelDimension{
			{
				Name:    "test",
				Options: []string{},
			},
			{
				Name:    "test2",
				Options: []string{},
			},
		}

		sut, err := CheckAllDimensionsHaveAnOption(testDimensions)
		So(err, ShouldBeNil)
		So(sut, ShouldEqual, false)
	})

	Convey("Two dimensions with one option selected return false", t, func() {
		testDimensions := []filter.ModelDimension{
			{
				Name:    "test",
				Options: []string{"option1"},
			},
			{
				Name:    "test2",
				Options: []string{},
			},
		}

		sut, err := CheckAllDimensionsHaveAnOption(testDimensions)
		So(err, ShouldBeNil)
		So(sut, ShouldEqual, false)
	})

	Convey("Two dimensions with one option selected return false", t, func() {
		testDimensions := []filter.ModelDimension{
			{
				Name:    "test",
				Options: []string{},
			},
			{
				Name:    "test2",
				Options: []string{"option1"},
			},
		}

		sut, err := CheckAllDimensionsHaveAnOption(testDimensions)
		So(err, ShouldBeNil)
		So(sut, ShouldEqual, false)
	})

	Convey("Three dimensions with only last option selected return false", t, func() {
		testDimensions := []filter.ModelDimension{
			{
				Name:    "test",
				Options: []string{},
			},
			{
				Name:    "test2",
				Options: []string{},
			},
			{
				Name:    "test3",
				Options: []string{"option1"},
			},
		}

		sut, err := CheckAllDimensionsHaveAnOption(testDimensions)
		So(err, ShouldBeNil)
		So(sut, ShouldEqual, false)
	})

	Convey("Three dimensions with only first option selected return false", t, func() {
		testDimensions := []filter.ModelDimension{
			{
				Name:    "test",
				Options: []string{"option1"},
			},
			{
				Name:    "test2",
				Options: []string{},
			},
			{
				Name:    "test3",
				Options: []string{},
			},
		}

		sut, err := CheckAllDimensionsHaveAnOption(testDimensions)
		So(err, ShouldBeNil)
		So(sut, ShouldEqual, false)
	})

	Convey("Five dimensions with only middle option selected return false", t, func() {
		testDimensions := []filter.ModelDimension{
			{
				Name:    "test",
				Options: []string{},
			},
			{
				Name:    "test2",
				Options: []string{},
			},
			{
				Name:    "test3",
				Options: []string{"option1"},
			},
			{
				Name:    "test4",
				Options: []string{},
			},
			{
				Name:    "test5",
				Options: []string{},
			},
		}

		sut, err := CheckAllDimensionsHaveAnOption(testDimensions)
		So(err, ShouldBeNil)
		So(sut, ShouldEqual, false)
	})

	Convey("Five dimensions with option in each return true", t, func() {
		testDimensions := []filter.ModelDimension{
			{
				Name:    "test",
				Options: []string{"option1"},
			},
			{
				Name:    "test2",
				Options: []string{"option1"},
			},
			{
				Name:    "test3",
				Options: []string{"option1"},
			},
			{
				Name:    "test4",
				Options: []string{"option1"},
			},
			{
				Name:    "test5",
				Options: []string{"option1"},
			},
		}

		sut, err := CheckAllDimensionsHaveAnOption(testDimensions)
		So(err, ShouldBeNil)
		So(sut, ShouldEqual, true)
	})
}
