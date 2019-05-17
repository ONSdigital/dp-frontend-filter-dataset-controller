package dates

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitDates(t *testing.T) {
	Convey("test ConvertToReadable", t, func() {
		Convey("converts a string slice with values in the format `yyyy.mm` to time.Time slice", func() {
			times, err := ConvertToReadable([]string{"May-06"})
			So(err, ShouldBeNil)
			So(times, ShouldHaveLength, 1)
			So(times[0].Month().String(), ShouldEqual, "May")
			So(times[0].Year(), ShouldEqual, 2006)
		})

		Convey("test error thrown if unable to parse date", func() {
			times, err := ConvertToReadable([]string{"Nop-13"})
			So(err, ShouldNotBeNil)
			So(times, ShouldBeEmpty)
		})
	})

	Convey("test ConvertToMonthYear", t, func() {
		time, err := time.Parse("01-02-2006", "05-01-2006")
		So(err, ShouldBeNil)
		formattedData := ConvertToMonthYear(time)
		So(formattedData, ShouldEqual, "May 2006")
	})

	Convey("test Sort sorts a list of times", t, func() {
		time1, _ := time.Parse("01-02-2006", "07-01-2006")
		time2, _ := time.Parse("01-02-2006", "02-01-2006")
		time3, _ := time.Parse("01-02-2006", "10-01-2006")
		times := Sort([]time.Time{time1, time2, time3})
		So(times[0].Equal(time2), ShouldBeTrue)
		So(times[1].Equal(time1), ShouldBeTrue)
		So(times[2].Equal(time3), ShouldBeTrue)
	})

	Convey("test ConvertToCoded converts a time list to a coded string list", t, func() {
		time1, _ := time.Parse("01-02-2006", "07-01-2006")
		time2, _ := time.Parse("01-02-2006", "02-01-2006")
		time3, _ := time.Parse("01-02-2006", "10-01-2006")

		codedDates := ConvertToCoded([]time.Time{time1, time2, time3})
		So(codedDates[0], ShouldEqual, "Jul-06")
		So(codedDates[1], ShouldEqual, "Feb-06")
		So(codedDates[2], ShouldEqual, "Oct-06")
	})
}
