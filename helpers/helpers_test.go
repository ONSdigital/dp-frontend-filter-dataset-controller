package helpers

import (
	"context"
	"net/url"
	"testing"

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
			So(err.Error(), ShouldEqual, "unabled to extract datasetID, edition and version from path: invalid")
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
		subjectSlice := []string{"foo", "Bar", "bAz", "quX", "QUUX", "COrGE", "gRaUlT", "GaRpLy", "WALdo", "frED", "pl ugh", "xyzzy1", "thud-", ""}
		Convey("and a string that the slice contains it should find said string", func() {
			testString := "foo"
			isFound := StringInSlice(testString, subjectSlice)
			So(isFound, ShouldEqual, true)
			Convey("it should be case sensitive and only yield true for exact matches", func() {
				testString = "Bar"
				isFound = StringInSlice(testString, subjectSlice)
				So(isFound, ShouldEqual, true)
				Convey("even if there is a random case difference in the middle of the string", func() {
					testString = "bAz"
					isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, true)
				})
				Convey("or a space", func() {
					testString = "pl ugh"
					isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, true)
				})
				Convey("or a numerical value represented as a char", func() {
					testString = "xyzzy1"
					isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, true)
				})
				Convey("or a symbolic character represented as a char", func() {
					testString = "thud-"
					isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, true)
				})
				Convey("or even a completely empty string", func() {
					testString = ""
					isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, true)
				})
			})
			Convey("it should be case sensitive and yield false for any other value", func() {
				Convey("like if the initial capital letter isn't considered", func() {
					testString = "bar"
					isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, false)
				})
				Convey("or a random capital in the middle of the string isn't considered", func() {
					testString = "baz"
					isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, false)
				})
				Convey("or a space in a string is missing", func() {
					testString = "plugh"
					isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, false)
				})
				Convey("or a numeral represented as a string char is missing", func() {
					testString = "xyzzy"
					isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, false)
				})
				Convey("or a symbol represented as a char is missing", func() {
					testString = "thud"
					isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, false)
				})
				Convey("or if a space char slips in and shouldn't be there", func() {
					testString = " "
					isFound = StringInSlice(testString, subjectSlice)
					So(isFound, ShouldEqual, false)
				})
			})
		})
	})
}
