package dates

import (
	"fmt"
	"sort"
	"time"
)

// TimeSlice allows sorting of a list of time.Time
type TimeSlice []time.Time

func (p TimeSlice) Len() int {
	return len(p)
}

func (p TimeSlice) Less(i, j int) bool {
	return p[i].Before(p[j])
}

func (p TimeSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// ConvertToReadable takes a list of dates in the format Jan-06 and converts to
// a list of time.Time
func ConvertToReadable(dates []string) ([]time.Time, error) {
	var readableDates []time.Time
	for _, val := range dates {
		date, err := time.Parse("Jan-06", val)
		if err != nil {
			return readableDates, err
		}
		readableDates = append(readableDates, date)
	}

	return readableDates, nil
}

// ConvertToMonthYear takes a time.Time object and converts to MM yyyy
func ConvertToMonthYear(d time.Time) string {
	return fmt.Sprintf("%s %d", d.Month().String(), d.Year())
}

// Sort orders a list of times
func Sort(dates []time.Time) TimeSlice {
	d := TimeSlice(dates)
	sort.Sort(d)

	return d
}

// ConvertToCoded takes a list of time.Time and converts to a list of strings in
// the format yyyy.mm
func ConvertToCoded(dates []time.Time) []string {
	var codedDates []string
	for _, date := range dates {
		codedDates = append(codedDates, date.Format("Jan-06"))
	}

	return codedDates
}
