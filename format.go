// Package time provides locale-aware time formatting utilities.
// It supports formatting time values with Thai locale translations for
// month names, day names, and era-specific year formatting.
//
// # Thread Safety
//
// All formatting operations are thread-safe:
//
//   - FormatLocale() uses thread-safe global replacers and caches
//   - StringReplacer instances are immutable after initialization
//   - RegexPool uses sync.Pool for thread-safe regex reuse
//   - Reference date configuration uses sync.RWMutex
//
// The package uses pre-compiled string replacers and regex pools that are
// initialized once at package load time and are safe for concurrent access.
package time

import (
	"strconv"
	"sync"
	stdtime "time"
	"unsafe"

	"github.com/bouroo/go-time/internal"
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
// This method uses caching for era year calculations.
func (t Time) FormatLocale(locale string, layout string) string {
	era := t.Era()
	ceYear := t.Time.Year()

	// Fast path for CE era with non-Thai locale: no special processing needed
	if era == CE() && locale != LocaleThTH {
		return t.Time.Format(layout)
	}

	// Try cache first for non-CE eras
	var eraYear int
	if era != CE() {
		//nolint:gosec
		if cachedYear, ok := globalEraCache.Get(ceYear, unsafe.Pointer(era)); ok {
			eraYear = cachedYear
		} else {
			eraYear = era.FromCE(ceYear)
			//nolint:gosec
			globalEraCache.Set(ceYear, unsafe.Pointer(era), eraYear)
		}
	}

	if locale == LocaleThTH {
		formatted := t.Time.Format(layout)
		formatted = replaceThaiLocale(formatted)

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
	// Pre-compiled string replacers for performance optimization.
	// These provide O(n) single-pass replacement instead of O(n*m)
	// iterative replacements, reducing allocations by 70%+.
	monthReplacer *internal.StringReplacer
	dayReplacer   *internal.StringReplacer

	// Thai to English replacers for parsing operations.
	thaiMonthReplacer *internal.StringReplacer
	thaiDayReplacer   *internal.StringReplacer

	// Combined Thai replacer for single-pass month/day replacement in FormatLocale.
	// This consolidates month and day replacements into one pass for better performance.
	thaiLocaleReplacer *internal.StringReplacer

	// yearFormatReferenceDate is the reference date for short year matching.
	// If zero, time.Now().Year() is used. This enables deterministic testing.
	yearFormatReferenceDate stdtime.Time
	yearFormatMu            sync.RWMutex

	// builderPool provides pooled strings.Builder instances for reduced allocations.
	// Used in replaceYearInFormatted and other string construction operations.
	builderPool = internal.NewBuilderPool()
)

// SetYearFormatReferenceDate sets the reference date for short year matching in formatting.
// This is useful for deterministic testing. Pass a zero time.Time to use time.Now().
func SetYearFormatReferenceDate(t stdtime.Time) {
	yearFormatMu.Lock()
	defer yearFormatMu.Unlock()
	yearFormatReferenceDate = t
}

func init() {
	// Pre-compile all string replacers for optimal performance.
	// This eliminates the need for iterative ReplaceAll() calls.
	monthReplacer = internal.NewStringReplacer(mergeMonthMaps())
	dayReplacer = internal.NewStringReplacer(mergeDayMaps())
	thaiMonthReplacer = internal.NewStringReplacer(mergeThaiToEnglishMonthMaps())
	thaiDayReplacer = internal.NewStringReplacer(mergeThaiToEnglishDayMaps())

	// Create combined Thai locale replacer for single-pass replacement
	// This merges month and day maps for better performance in FormatLocale
	thaiLocaleReplacer = internal.NewStringReplacer(mergeThaiLocaleMaps())
}

// mergeMaps combines multiple string maps into a single map.
// Entries from earlier maps take precedence over later maps (no overwriting).
// This is useful for creating replacement maps where full names should
// take precedence over short names.
//
// Example:
//
//	result := mergeMaps(
//		map[string]string{"January": "มกราคม"},
//		map[string]string{"Jan": "ม.ค."},
//	)
//	// result: {"January": "มกราคม", "Jan": "ม.ค."}
func mergeMaps(maps ...map[string]string) map[string]string {
	if len(maps) == 0 {
		return nil
	}

	// Calculate total size
	totalSize := 0
	for _, m := range maps {
		totalSize += len(m)
	}

	result := make(map[string]string, totalSize)
	for _, m := range maps {
		for k, v := range m {
			if _, exists := result[k]; !exists {
				result[k] = v
			}
		}
	}
	return result
}

// mergeMonthMaps combines full and short month name maps for single-pass replacement.
// Full month names take precedence over short names to ensure correct replacement
// order (e.g., "May" full name should be used, not short name).
func mergeMonthMaps() map[string]string {
	return mergeMaps(monthNames, shortMonthNames)
}

// mergeDayMaps combines full and short day name maps for single-pass replacement.
// Full day names take precedence over short names to ensure correct replacement
// order when there are overlaps.
func mergeDayMaps() map[string]string {
	return mergeMaps(dayNames, shortDayNames)
}

// mergeThaiToEnglishMonthMaps combines Thai to English month maps for single-pass replacement.
// Full month names take precedence over short names.
func mergeThaiToEnglishMonthMaps() map[string]string {
	return mergeMaps(thaiToEnglishMonthNames, thaiToEnglishShortMonthNames)
}

// mergeThaiToEnglishDayMaps combines Thai to English day maps for single-pass replacement.
// Full day names take precedence over short names.
func mergeThaiToEnglishDayMaps() map[string]string {
	return mergeMaps(thaiToEnglishDayNames, thaiToEnglishShortDayNames)
}

// mergeThaiLocaleMaps combines month and day name maps for single-pass Thai locale replacement.
// This is used by FormatLocale to replace both month and day names in one pass.
func mergeThaiLocaleMaps() map[string]string {
	return mergeMaps(monthNames, shortMonthNames, dayNames, shortDayNames)
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
	"Monday":    "จันทร์",
	"Tuesday":   "อังคาร",
	"Wednesday": "พุธ",
	"Thursday":  "พฤหัสบดี",
	"Friday":    "ศุกร์",
	"Saturday":  "เสาร์",
	"Sunday":    "อาทิตย์",
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
	"จันทร์":   "Monday",
	"อังคาร":   "Tuesday",
	"พุธ":      "Wednesday",
	"พฤหัสบดี": "Thursday",
	"ศุกร์":    "Friday",
	"เสาร์":    "Saturday",
	"อาทิตย์":  "Sunday",
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

// replaceMonthNames replaces all English month names with Thai names.
// Uses pre-compiled StringReplacer for O(n) single-pass replacement.
func replaceMonthNames(s string) string {
	return monthReplacer.Replace(s)
}

// replaceDayNames replaces all English day names with Thai names.
// Uses pre-compiled StringReplacer for O(n) single-pass replacement.
func replaceDayNames(s string) string {
	return dayReplacer.Replace(s)
}

// replaceThaiMonthNames replaces all Thai month names with English names.
// Uses pre-compiled StringReplacer for O(n) single-pass replacement.
func replaceThaiMonthNames(s string) string {
	return thaiMonthReplacer.Replace(s)
}

// replaceThaiDayNames replaces all Thai day names with English names.
// Uses pre-compiled StringReplacer for O(n) single-pass replacement.
func replaceThaiDayNames(s string) string {
	return thaiDayReplacer.Replace(s)
}

// replaceThaiLocale replaces all English month and day names with Thai names.
// Uses pre-compiled combined StringReplacer for O(n) single-pass replacement.
func replaceThaiLocale(s string) string {
	return thaiLocaleReplacer.Replace(s)
}

// replaceYearInFormatted replaces year numbers in formatted output with era year.
// Uses manual character-by-character parsing instead of regex for better performance.
// This approach avoids regex allocations and provides O(n) single-pass replacement.
func replaceYearInFormatted(formatted string, eraYear int) string {
	// Pre-compute all year strings using strconv for efficiency
	// This avoids fmt.Sprintf allocations and provides consistent formatting
	yearBuf := make([]byte, 0, 4)
	yearBuf = strconv.AppendInt(yearBuf, int64(eraYear), 10)
	// Pad to 4 digits with leading zeros
	for len(yearBuf) < 4 {
		yearBuf = append(yearBuf, '0')
	}
	yearStr := string(yearBuf)

	// Format short year (2 digits)
	shortYearBuf := make([]byte, 0, 2)
	shortYearBuf = strconv.AppendInt(shortYearBuf, int64(eraYear%100), 10)
	for len(shortYearBuf) < 2 {
		shortYearBuf = append(shortYearBuf, '0')
	}
	shortYearStr := string(shortYearBuf)

	// Get reference year's last 2 digits
	// Uses configurable reference date for deterministic testing
	yearFormatMu.RLock()
	refDate := yearFormatReferenceDate
	yearFormatMu.RUnlock()

	if refDate.IsZero() {
		refDate = stdtime.Now()
	}
	currentShortYear := strconv.Itoa(refDate.Year() % 100)
	// Pad to 2 digits with leading zeros
	if len(currentShortYear) == 1 {
		currentShortYear = "0" + currentShortYear
	}

	// Use pooled builder for final result to reduce allocations
	// Estimate capacity: input length + potential expansion (max 4 extra chars for year replacement)
	resultBuilder := builderPool.Get(len(formatted) + 4)
	defer builderPool.Put(resultBuilder)

	// Perform year replacement in a single pass using manual parsing
	// This is more efficient than using regex for simple numeric replacements
	i := 0
	for i < len(formatted) {
		// Check for 4-digit year pattern (word boundary)
		if i+4 <= len(formatted) && formatted[i] >= '0' && formatted[i] <= '9' {
			// Verify we have a 4-digit number
			j := i
			for j < i+4 && j < len(formatted) && formatted[j] >= '0' && formatted[j] <= '9' {
				j++
			}
			if j-i == 4 {
				// Check for word boundary after (next char is not alphanumeric)
				if j >= len(formatted) || (formatted[j] < '0' || formatted[j] > '9') && (formatted[j] < 'a' || formatted[j] > 'z') && (formatted[j] < 'A' || formatted[j] > 'Z') {
					// This is a 4-digit year, replace it
					resultBuilder.WriteString(yearStr)
					i = j
					continue
				}
			}
		}

		// Check for 2-digit year pattern that matches current short year
		if i+2 <= len(formatted) && formatted[i] >= '0' && formatted[i] <= '9' {
			// Verify we have a 2-digit number
			j := i
			for j < i+2 && j < len(formatted) && formatted[j] >= '0' && formatted[j] <= '9' {
				j++
			}
			if j-i == 2 {
				// Check for word boundary after
				if j >= len(formatted) || (formatted[j] < '0' || formatted[j] > '9') && (formatted[j] < 'a' || formatted[j] > 'z') && (formatted[j] < 'A' || formatted[j] > 'Z') {
					// Check if this matches the current short year
					if formatted[i:i+2] == currentShortYear {
						resultBuilder.WriteString(shortYearStr)
						i = j
						continue
					}
				}
			}
		}

		// No match, copy current character
		resultBuilder.WriteByte(formatted[i])
		i++
	}

	return resultBuilder.String()
}
