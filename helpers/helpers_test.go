package helpers

import (
	"context"
	"net/url"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
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
	cases := []struct {
		Description string
		Dimensions  []filter.ModelDimension
		Expected    bool
	}{
		{"Dimension with no option selected return false", []filter.ModelDimension{
			{
				Name:    "test",
				Options: []string{},
			}}, false},
		{"Three dimensions with options in each return true", []filter.ModelDimension{
			{
				Name:    "test",
				Options: []string{"option1"},
			},
			{
				Name:    "test2",
				Options: []string{"option2"},
			},
			{
				Name:    "test3",
				Options: []string{"option3"},
			}}, true},
		{"Two dimensions with no options selected return false", []filter.ModelDimension{
			{
				Name:    "test",
				Options: []string{},
			},
			{
				Name:    "test2",
				Options: []string{},
			}}, false},
		{"Three dimensions with only last option selected return false", []filter.ModelDimension{
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
			}}, false},
		{"Three dimensions with only first option selected return false", []filter.ModelDimension{
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
			}}, false},
		{"Five dimensions with only middle option selected return false", []filter.ModelDimension{
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
			}}, false},
	}

	for _, test := range cases {
		Convey(test.Description, t, func() {
			got, err := CheckAllDimensionsHaveAnOption(test.Dimensions)
			So(err, ShouldBeNil)
			So(got, ShouldEqual, test.Expected)
		})
	}

	Convey("No dimensions provided return error", t, func() {
		testDimensions := []filter.ModelDimension{}

		sut, err := CheckAllDimensionsHaveAnOption(testDimensions)
		So(err, ShouldBeError, "no dimensions provided: []")
		So(sut, ShouldEqual, false)
	})
}

// TestTitleCaseStr test the helper function TitleCaseStr
func TestTitleCaseStr(t *testing.T) {
	cases := []struct {
		Description string
		Input       string
		Expected    string
	}{
		{"Given string to title case", "title", "Title"},
		{"Given string in title case", "Title", "Title"},
		{"Given string in mixed case", "tItle", "TItle"},        // tests exist that expect this behaviour
		{"Given strings to title case", "title me", "Title Me"}, // tests exist that expect this behaviour
		{"Empty string given", "", ""},
	}

	for _, test := range cases {
		Convey(test.Description, t, func() {
			got := TitleCaseStr(test.Input)
			So(got, ShouldEqual, test.Expected)
		})
	}
}
