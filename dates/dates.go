package dates

import (
	"fmt"
	"regexp"
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

// ConvertToReadable takes a list of dates in the format yyyy.mm and converts to
// a list of time.Time
func ConvertToReadable(dates []string) ([]time.Time, error) {
	var readableDates []time.Time
	for _, val := range dates {
		myrReg := regexp.MustCompile(`^(\d{4})\.(\d{1}|\d{2})$`)
		myrSubs := myrReg.FindStringSubmatch(val)
		if len(myrSubs) == 3 {
			date, err := time.Parse("01-02-2006", fmt.Sprintf("%02s-01-%s", myrSubs[2], myrSubs[1]))
			if err != nil {
				return readableDates, err
			}
			readableDates = append(readableDates, date)
		}
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
		codedDates = append(codedDates, fmt.Sprintf("%d.%02d", date.Year(), date.Month()))
	}

	return codedDates
}
