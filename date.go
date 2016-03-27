package tm

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

//go:generate msgp

// Date represents a UTC time zone day
type Date struct {
	Year  int
	Month int
	Day   int
}

// ParseDate converts a datestring '2016/02/25' into a Date{} struct.
func ParseDate(datestring string) (*Date, error) {
	parts := strings.Split(datestring, "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("bad datestring '%s': did not have two slashes", datestring)
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse year", datestring)
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse month", datestring)
	}
	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("bad datestring '%s': could not parse day", datestring)
	}

	if year < 1970 || year > 3000 {
		return nil, fmt.Errorf("year out of bounds: %v", year)
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("month out of bounds: %v", month)
	}
	if day < 1 || day > 31 {
		return nil, fmt.Errorf("day out of bounds: %v", day)
	}

	return &Date{Year: year, Month: month, Day: day}, nil
}

var WestCoastUSLocation *time.Location
var EastCoastUSLocation *time.Location
var LondonLocation *time.Location
var UTCLocation = time.UTC

func init() {
	var err error
	WestCoastUSLocation, err = time.LoadLocation("America/Los_Angeles")
	panicOn(err)
	EastCoastUSLocation, err = time.LoadLocation("America/New_York")
	panicOn(err)
	LondonLocation, err = time.LoadLocation("Europe/London")
	panicOn(err)
}

// UTCDateFromTime returns the date after tm is moved to the UTC time zone.
func UTCDateFromTime(tm time.Time) *Date {
	y, m, d := tm.In(time.UTC).Date()
	return &Date{Year: y, Month: int(m), Day: d}
}

// Unix converts the date into an int64 representing the nanoseconds
// since the unix epoch for the ToGoTime() output of Date d.
func (d *Date) Unix() int64 {
	return d.ToGoTime().Unix()
}

// ToGoTime turns the date into UTC time.Time, at the 0 hrs 0 min 0 second start of the day.
func (d *Date) ToGoTime() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
}

// String turns the date into a string.
func (d *Date) String() string {
	return fmt.Sprintf("%04d/%02d/%02d", d.Year, d.Month, d.Day)
}

// return true if a < b
func DateBefore(a *Date, b *Date) bool {
	if a.Year < b.Year {
		return true
	} else if a.Year > b.Year {
		return false
	}

	if a.Month < b.Month {
		return true
	} else if a.Month > b.Month {
		return false
	}

	if a.Day < b.Day {
		return true
	} else if a.Day > b.Day {
		return false
	}

	return false
}

// return true if a > b
func DateAfter(a *Date, b *Date) bool {
	if a.Year > b.Year {
		return true
	} else if a.Year < b.Year {
		return false
	}

	if a.Month > b.Month {
		return true
	} else if a.Month < b.Month {
		return false
	}

	if a.Day > b.Day {
		return true
	} else if a.Day < b.Day {
		return false
	}

	return false
}

// DatesEqual returns true if a and b are the exact same day.
func DatesEqual(a *Date, b *Date) bool {
	if a.Year == b.Year {
		if a.Month == b.Month {
			if a.Day == b.Day {
				return true
			}
		}
	}
	return false
}

// NextDate returns the next calendar day after d.
func NextDate(d *Date) *Date {
	tm := d.ToGoTime()
	next := tm.AddDate(0, 0, 1)
	return UTCDateFromTime(next)
}

// PrevDate returns the first calendar day prior to d.
func PrevDate(d *Date) *Date {
	tm := d.ToGoTime()
	next := tm.AddDate(0, 0, -1)
	return UTCDateFromTime(next)
}

// TimeToDate returns the UTC Date associated with tm.
func TimeToDate(tm time.Time) Date {
	utc := tm.UTC()
	return Date{
		Year:  utc.Year(),
		Month: int(utc.Month()),
		Day:   utc.Day(),
	}
}
