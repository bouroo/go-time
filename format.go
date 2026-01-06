// Package gotime provides locale-aware time formatting utilities.
// It supports formatting time values with Thai locale translations for
// month names, day names, and era-specific year formatting.
package gotime

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Locale constants for formatting.
const (
	// LocaleThTH represents the Thai (Thailand) locale for formatting.
	LocaleThTH = "th-TH"
	// LocaleEnUS represents the English (United States) locale for formatting.
	LocaleEnUS = "en-US"
	// LocaleDefault represents the default locale (no special formatting).
	LocaleDefault = ""
)

// FormatLocale formats the time value according to the specified locale and layout.
// For Thai locale (th-TH), it translates month and day names to Thai.
// It also adjusts the year to the appropriate era based on the time's era setting.
func (t Time) FormatLocale(locale string, layout string) string {
	era := t.Era()
	ceYear := t.Time.Year()
	eraYear := ceYear + era.Offset()

	if locale == LocaleThTH {
		formatted := t.Time.Format(layout)
		formatted = replaceMonthNames(formatted)
		formatted = replaceDayNames(formatted)

		if era != CE() {
			formatted = replaceYearInFormatted(formatted, eraYear)
		}
		return formatted
	}

	if era != CE() {
		formatted := t.Time.Format(layout)
		return replaceYearInFormatted(formatted, eraYear)
	}

	return t.Time.Format(layout)
}

var (
	yearRegex      *regexp.Regexp
	shortYearRegex *regexp.Regexp
)

func init() {
	// Match 4-digit years from 1000+ (covers past/future eras)
	yearRegex = regexp.MustCompile(`\b\d{4}\b`)
	shortYearRegex = regexp.MustCompile(`\b\d{2}\b`)
}

var monthNames = map[string]string{
	"January":   "มกราคม",
	"February":  "กุมภาพันธ์",
	"March":     "มีนาคม",
	"April":     "เมษายน",
	"May":       "พฤษภาคม",
	"June":      "มิถุนายน",
	"July":      "กรกฎาคม",
	"August":    "สิงหาคม",
	"September": "กันยายน",
	"October":   "ตุลาคม",
	"November":  "พฤศจิกายน",
	"December":  "ธันวาคม",
}

var shortMonthNames = map[string]string{
	"Jan": "ม.ค.",
	"Feb": "ก.พ.",
	"Mar": "มี.ค.",
	"Apr": "เม.ย.",
	"May": "พ.ค.",
	"Jun": "มิ.ย.",
	"Jul": "ก.ค.",
	"Aug": "ส.ค.",
	"Sep": "ก.ย.",
	"Oct": "ต.ค.",
	"Nov": "พ.ย.",
	"Dec": "ธ.ค.",
}

var dayNames = map[string]string{
	"Monday":    "วันจันทร์",
	"Tuesday":   "วันอังคาร",
	"Wednesday": "วันพุธ",
	"Thursday":  "วันพฤหัสบดี",
	"Friday":    "วันศุกร์",
	"Saturday":  "วันเสาร์",
	"Sunday":    "วันอาทิตย์",
}

var shortDayNames = map[string]string{
	"Mon": "จ.",
	"Tue": "อ.",
	"Wed": "พ.",
	"Thu": "พฤ.",
	"Fri": "ศ.",
	"Sat": "ส.",
	"Sun": "อา.",
}

var thaiToEnglishMonthNames = map[string]string{
	"มกราคม":     "January",
	"กุมภาพันธ์": "February",
	"มีนาคม":     "March",
	"เมษายน":     "April",
	"พฤษภาคม":    "May",
	"มิถุนายน":   "June",
	"กรกฎาคม":    "July",
	"สิงหาคม":    "August",
	"กันยายน":    "September",
	"ตุลาคม":     "October",
	"พฤศจิกายน":  "November",
	"ธันวาคม":    "December",
}

var thaiToEnglishShortMonthNames = map[string]string{
	"ม.ค.":  "Jan",
	"ก.พ.":  "Feb",
	"มี.ค.": "Mar",
	"เม.ย.": "Apr",
	"พ.ค.":  "May",
	"มิ.ย.": "Jun",
	"ก.ค.":  "Jul",
	"ส.ค.":  "Aug",
	"ก.ย.":  "Sep",
	"ต.ค.":  "Oct",
	"พ.ย.":  "Nov",
	"ธ.ค.":  "Dec",
}

var thaiToEnglishDayNames = map[string]string{
	"วันจันทร์":   "Monday",
	"วันอังคาร":   "Tuesday",
	"วันพุธ":      "Wednesday",
	"วันพฤหัสบดี": "Thursday",
	"วันศุกร์":    "Friday",
	"วันเสาร์":    "Saturday",
	"วันอาทิตย์":  "Sunday",
}

var thaiToEnglishShortDayNames = map[string]string{
	"จ.":  "Mon",
	"อ.":  "Tue",
	"พ.":  "Wed",
	"พฤ.": "Thu",
	"ศ.":  "Fri",
	"ส.":  "Sat",
	"อา.": "Sun",
}

func replaceMonthNames(s string) string {
	result := s
	for en, th := range monthNames {
		result = strings.ReplaceAll(result, en, th)
	}
	for en, th := range shortMonthNames {
		result = strings.ReplaceAll(result, en, th)
	}
	return result
}

func replaceDayNames(s string) string {
	result := s
	for en, th := range dayNames {
		result = strings.ReplaceAll(result, en, th)
	}
	for en, th := range shortDayNames {
		result = strings.ReplaceAll(result, en, th)
	}
	return result
}

func replaceThaiMonthNames(s string) string {
	result := s
	for th, en := range thaiToEnglishMonthNames {
		result = strings.ReplaceAll(result, th, en)
	}
	for th, en := range thaiToEnglishShortMonthNames {
		result = strings.ReplaceAll(result, th, en)
	}
	return result
}

func replaceThaiDayNames(s string) string {
	result := s
	for th, en := range thaiToEnglishDayNames {
		result = strings.ReplaceAll(result, th, en)
	}
	for th, en := range thaiToEnglishShortDayNames {
		result = strings.ReplaceAll(result, th, en)
	}
	return result
}

func replaceYearInFormatted(formatted string, eraYear int) string {
	yearStr := fmt.Sprintf("%04d", eraYear)
	shortYearStr := fmt.Sprintf("%02d", eraYear%100)

	result := yearRegex.ReplaceAllString(formatted, yearStr)

	// Get current CE year's last 2 digits to match against the formatted output
	currentCEYear := time.Now().Year()
	currentShortYear := fmt.Sprintf("%02d", currentCEYear%100)

	result = shortYearRegex.ReplaceAllStringFunc(result, func(match string) string {
		if match == currentShortYear {
			return shortYearStr
		}
		return match
	})

	return result
}
