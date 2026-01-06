// Package gotime provides an enhanced Time type that wraps the standard
// library's time.Time with era-specific functionality. It supports Buddhist Era
// (BE) and Common Era (CE) calendars, Thai language text processing, and
// locale-aware formatting.
package gotime

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var (
	beYearRegex = regexp.MustCompile(`\b(\d{4,10})\b`)
)

// Time wraps time.Time with era-specific functionality.
// It embeds the standard library's Time type and adds an optional Era field
// to support Buddhist Era and other calendar systems.
type Time struct {
	time.Time
	era *Era
}

// Now returns the current local time with no era set (defaults to CE).
func Now() Time {
	return Time{Time: time.Now(), era: nil}
}

// Date constructs a Time with the given components and no era set (defaults to CE).
// It follows the same signature as time.Date from the standard library.
func Date(year, month, day, hour, min, sec, nsec int, loc *time.Location) Time {
	return Time{Time: time.Date(year, time.Month(month), day, hour, min, sec, nsec, loc), era: nil}
}

// Era returns the era associated with this time, or CE if no era is set.
func (t Time) Era() *Era {
	if t.era == nil {
		return CE()
	}
	return t.era
}

// InEra returns a new Time with the specified era. If the given era is nil,
// it defaults to CE.
func (t Time) InEra(e *Era) Time {
	if e == nil {
		e = CE()
	}
	return Time{Time: t.Time, era: e}
}

// Year returns the year in the associated era. For BE era, this returns
// the Buddhist Era year (e.g., 2567 for CE 2024).
func (t Time) Year() int {
	return t.Era().FromCE(t.Time.Year())
}

// YearCE returns the year in Common Era, regardless of the associated era.
func (t Time) YearCE() int {
	return t.Time.Year()
}

// Month returns the month of the year (January=1, December=12).
func (t Time) Month() time.Month {
	return t.Time.Month()
}

// Day returns the day of the month (1-31).
func (t Time) Day() int {
	return t.Time.Day()
}

// Hour returns the hour within the day (0-23).
func (t Time) Hour() int {
	return t.Time.Hour()
}

// Minute returns the minute within the hour (0-59).
func (t Time) Minute() int {
	return t.Time.Minute()
}

// Second returns the second within the minute (0-59).
func (t Time) Second() int {
	return t.Time.Second()
}

// Nanosecond returns the nanosecond within the second (0-999999999).
func (t Time) Nanosecond() int {
	return t.Time.Nanosecond()
}

// Location returns the time's location.
func (t Time) Location() *time.Location {
	return t.Time.Location()
}

// Zone returns the time zone name and offset from UTC.
func (t Time) Zone() (name string, offset int) {
	return t.Time.Zone()
}

// Unix returns t as a Unix time, the number of seconds elapsed
// since January 1, 1970 UTC.
func (t Time) Unix() int64 {
	return t.Time.Unix()
}

// UnixNano returns t as a Unix time, the number of nanoseconds elapsed
// since January 1, 1970 UTC.
func (t Time) UnixNano() int64 {
	return t.Time.UnixNano()
}

// IsZero reports whether t represents the zero time instant.
func (t Time) IsZero() bool {
	return t.Time.IsZero()
}

// IsLeap reports whether the year in Common Era is a leap year.
// A leap year is divisible by 4, except for century years which must be
// divisible by 400.
func (t Time) IsLeap() bool {
	year := t.YearCE()
	return (year%4 == 0 && year%100 != 0) || year%400 == 0
}

// IsCE reports whether this time is in Common Era (or has no era set).
func (t Time) IsCE() bool {
	return t.era == nil || t.era == CE()
}

// IsBE reports whether this time is in Buddhist Era.
func (t Time) IsBE() bool {
	return t.era != nil && t.era == BE()
}

// Format returns the time formatted according to layout.
// If the time's era is not CE, the year in the formatted output
// is adjusted to the appropriate era year.
func (t Time) Format(layout string) string {
	era := t.Era()
	ceYear := t.Time.Year()
	eraYear := ceYear + era.Offset()

	if era != CE() {
		formatted := t.Time.Format(layout)
		return replaceYearInFormatted(formatted, eraYear)
	}

	return t.Time.Format(layout)
}

// String returns the time formatted as "2006-01-02 15:04:05 -0700 MST".
func (t Time) String() string {
	return t.Format("2006-01-02 15:04:05 -0700 MST")
}

// Add returns the time t+d.
func (t Time) Add(d time.Duration) Time {
	return Time{Time: t.Time.Add(d), era: t.era}
}

// Sub returns the duration t-u.
func (t Time) Sub(u Time) time.Duration {
	return t.Time.Sub(u.Time)
}

// Before reports whether the time t is before u.
func (t Time) Before(u Time) bool {
	return t.Time.Before(u.Time)
}

// After reports whether the time t is after u.
func (t Time) After(u Time) bool {
	return t.Time.After(u.Time)
}

// Equal reports whether t and u represent the same time instant.
func (t Time) Equal(u Time) bool {
	return t.Time.Equal(u.Time)
}

// MarshalJSON implements json.Marshaler. The time is marshaled
// in the same format as time.Time.MarshalJSON.
func (t Time) MarshalJSON() ([]byte, error) {
	return t.Time.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler. The time is unmarshaled
// in the same format as time.Time.UnmarshalJSON.
func (t *Time) UnmarshalJSON(data []byte) error {
	return t.Time.UnmarshalJSON(data)
}

// GobEncode implements gob.GobEncoder.
func (t Time) GobEncode() ([]byte, error) {
	return t.Time.GobEncode()
}

// GobDecode implements gob.GobDecoder.
func (t *Time) GobDecode(data []byte) error {
	return t.Time.GobDecode(data)
}

// Parse is a wrapper around time.Parse from the standard library.
// It parses a formatted time string and returns the result as time.Time.
func Parse(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// ParseInLocation is a wrapper around time.ParseInLocation from the standard library.
// It parses a formatted time string in the given location.
func ParseInLocation(layout, value string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(layout, value, loc)
}

// ParseWithEra parses a time string with era-specific processing.
// It converts Thai month and day names to English before parsing.
// If the era is BE, it also converts Buddhist Era years to Common Era.
// Returns a ParseError if parsing fails.
func ParseWithEra(layout, value string, era *Era) (Time, error) {
	if era == nil {
		era = CE()
	}

	converted := replaceThaiMonthNames(value)
	converted = replaceThaiDayNames(converted)

	if era == BE() {
		converted = convertBEYearToCE(converted)
	}

	t, err := time.Parse(layout, converted)
	if err != nil {
		return Time{}, &ParseError{
			Input:    value,
			Layout:   layout,
			Era:      era,
			Original: err,
		}
	}

	return Time{Time: t, era: era}, nil
}

// ParseInLocationWithEra parses a time string in a specific location with
// era-specific processing. It converts Thai month and day names to English
// before parsing. If the era is BE, it also converts Buddhist Era years
// to Common Era. Returns a ParseError if parsing fails.
func ParseInLocationWithEra(layout, value string, loc *time.Location, era *Era) (Time, error) {
	if era == nil {
		era = CE()
	}

	converted := replaceThaiMonthNames(value)
	converted = replaceThaiDayNames(converted)

	if era == BE() {
		converted = convertBEYearToCE(converted)
	}

	t, err := time.ParseInLocation(layout, converted, loc)
	if err != nil {
		return Time{}, &ParseError{
			Input:    value,
			Layout:   layout,
			Era:      era,
			Original: err,
		}
	}

	return Time{Time: t, era: era}, nil
}

// ParseThai parses a time string that may contain Thai month and day names.
// It automatically detects whether the year is in BE or CE format based on
// proximity to the current year, and returns a Time with the detected era.
func ParseThai(layout, value string) (Time, error) {
	converted := replaceThaiMonthNames(value)
	converted = replaceThaiDayNames(converted)

	t, err := time.Parse(layout, converted)
	if err != nil {
		return Time{}, err
	}

	detectedEra := DetectEraFromYear(t.Year())

	if detectedEra == BE() {
		ceYear := BE().ToCE(t.Year())
		t = time.Date(ceYear, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
		return Time{Time: t, era: BE()}, nil
	}

	return Time{Time: t, era: CE()}, nil
}

// ParseThaiInLocation parses a time string with Thai month and day names
// in a specific location. It automatically detects whether the year is in
// BE or CE format based on proximity to the current year.
func ParseThaiInLocation(layout, value string, loc *time.Location) (Time, error) {
	converted := replaceThaiMonthNames(value)
	converted = replaceThaiDayNames(converted)

	t, err := time.ParseInLocation(layout, converted, loc)
	if err != nil {
		return Time{}, err
	}

	detectedEra := DetectEraFromYear(t.Year())

	if detectedEra == BE() {
		ceYear := BE().ToCE(t.Year())
		t = time.Date(ceYear, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
		return Time{Time: t, era: BE()}, nil
	}

	return Time{Time: t, era: CE()}, nil
}

func convertBEYearToCE(value string) string {
	ceValue := beYearRegex.ReplaceAllStringFunc(value, func(match string) string {
		year, err := strconv.Atoi(match)
		if err != nil {
			return match
		}
		if DetectEraFromYear(year) == BE() {
			ceYear := BE().ToCE(year)
			return fmt.Sprintf("%d", ceYear)
		}
		return match
	})
	return ceValue
}
