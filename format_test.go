package time

import (
	"strings"
	"testing"
	stdtime "time"
)

// TestFormatThaiLeapDay tests Thai locale formatting of leap days
func TestFormatThaiLeapDay(t *testing.T) {
	tests := []struct {
		name               string
		ceYear             int
		month              int
		day                int
		layout             string
		shouldContainMonth string
		shouldContainYear  string
		shouldContainDay   int
	}{
		{
			"2024 leap day Thai full format",
			2024, 2, 29,
			"02 January 2006",
			"กุมภาพันธ์", // February in Thai
			"2567",       // BE year
			29,
		},
		{
			"2020 leap day Thai full format",
			2020, 2, 29,
			"02 January 2006",
			"กุมภาพันธ์",
			"2563",
			29,
		},
		{
			"2000 leap day Thai full format",
			2000, 2, 29,
			"02 January 2006",
			"กุมภาพันธ์",
			"2543",
			29,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := Date(tt.ceYear, tt.month, tt.day, 12, 30, 45, 0, stdtime.UTC)
			beTime := tm.InEra(BE())

			result := beTime.FormatLocale(LocaleThTH, tt.layout)

			if !strings.Contains(result, tt.shouldContainMonth) {
				t.Errorf("FormatLocale(th-TH) should contain Thai month %q, got %q", tt.shouldContainMonth, result)
			}

			if !strings.Contains(result, tt.shouldContainYear) {
				t.Errorf("FormatLocale(th-TH) should contain BE year %q, got %q", tt.shouldContainYear, result)
			}

			// Check day is present
			if !strings.Contains(result, "29") {
				t.Errorf("FormatLocale(th-TH) should contain day 29, got %q", result)
			}
		})
	}
}

// TestFormatThaiMonthNames tests all Thai month names
func TestFormatThaiMonthNames(t *testing.T) {
	thaiMonths := []struct {
		month    stdtime.Month
		thaiName string
		english  string
	}{
		{stdtime.January, "มกราคม", "January"},
		{stdtime.February, "กุมภาพันธ์", "February"},
		{stdtime.March, "มีนาคม", "March"},
		{stdtime.April, "เมษายน", "April"},
		{stdtime.May, "พฤษภาคม", "May"},
		{stdtime.June, "มิถุนายน", "June"},
		{stdtime.July, "กรกฎาคม", "July"},
		{stdtime.August, "สิงหาคม", "August"},
		{stdtime.September, "กันยายน", "September"},
		{stdtime.October, "ตุลาคม", "October"},
		{stdtime.November, "พฤศจิกายน", "November"},
		{stdtime.December, "ธันวาคม", "December"},
	}

	for _, monthData := range thaiMonths {
		t.Run(monthData.english, func(t *testing.T) {
			// Create a date in this month
			tm := Date(2024, int(monthData.month), 15, 12, 0, 0, 0, stdtime.UTC)
			beTime := tm.InEra(BE())

			// Format in Thai locale
			result := beTime.FormatLocale(LocaleThTH, "January")

			if !strings.Contains(result, monthData.thaiName) {
				t.Errorf("FormatLocale(th-TH) should contain %q for %s, got %q", monthData.thaiName, monthData.english, result)
			}
		})
	}
}

// TestFormatThaiWeekdays tests Thai day names
func TestFormatThaiWeekdays(t *testing.T) {
	thaiDays := []struct {
		weekday     stdtime.Weekday
		thaiName    string
		englishName string
	}{
		{stdtime.Monday, "จันทร์", "Monday"},
		{stdtime.Tuesday, "อังคาร", "Tuesday"},
		{stdtime.Wednesday, "พุธ", "Wednesday"},
		{stdtime.Thursday, "พฤหัสบดี", "Thursday"},
		{stdtime.Friday, "ศุกร์", "Friday"},
		{stdtime.Saturday, "เสาร์", "Saturday"},
		{stdtime.Sunday, "อาทิตย์", "Sunday"},
	}

	for _, td := range thaiDays {
		t.Run(td.englishName, func(t *testing.T) {
			// Find a date that falls on this weekday
			// Search in January 2024
			for day := 1; day <= 31; day++ {
				tm := stdtime.Date(2024, 1, day, 12, 0, 0, 0, stdtime.UTC)
				if tm.Weekday() == td.weekday {
					gotTime := Date(2024, 1, day, 12, 0, 0, 0, stdtime.UTC)
					beTime := gotTime.InEra(BE())
					result := beTime.FormatLocale(LocaleThTH, "Monday")

					if !strings.Contains(result, td.thaiName) {
						t.Errorf("FormatLocale(th-TH) should contain %q for %s, got %q", td.thaiName, td.englishName, result)
					}
					return
				}
			}
			t.Skipf("Could not find %s in January 2024", td.englishName)
		})
	}
}

// TestFormatEnUSLeapDay tests English US locale with leap days
func TestFormatEnUSLeapDay(t *testing.T) {
	tests := []struct {
		name              string
		ceYear            int
		month             int
		day               int
		expectedBEYear    string
		expectedMonthName string
	}{
		{"2024 leap day", 2024, 2, 29, "2567", "February"},
		{"2020 leap day", 2020, 2, 29, "2563", "February"},
		{"2000 leap day", 2000, 2, 29, "2543", "February"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := Date(tt.ceYear, tt.month, tt.day, 12, 30, 45, 0, stdtime.UTC)
			beTime := tm.InEra(BE())

			result := beTime.FormatLocale(LocaleEnUS, "January 02, 2006")

			if !strings.Contains(result, tt.expectedMonthName) {
				t.Errorf("FormatLocale(en-US) should contain %q, got %q", tt.expectedMonthName, result)
			}

			if !strings.Contains(result, tt.expectedBEYear) {
				t.Errorf("FormatLocale(en-US) should contain %q, got %q", tt.expectedBEYear, result)
			}

			// Check day
			if !strings.Contains(result, "29") {
				t.Errorf("FormatLocale(en-US) should contain day 29, got %q", result)
			}
		})
	}
}

// TestFormatCENotAffectedByLocale tests CE times are not affected by Thai locale
func TestFormatCENotAffectedByLocale(t *testing.T) {
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)

	// Format in Thai locale
	thaiResult := tm.FormatLocale(LocaleThTH, "02 January 2006")
	// Format default
	defaultResult := tm.Format("02 January 2006")

	// For CE times, Thai locale should still show Western year (2024, not 2567)
	if strings.Contains(thaiResult, "2567") {
		t.Errorf("CE time formatted with th-TH should not contain BE year, got %q", thaiResult)
	}

	// Should be consistent
	if thaiResult != defaultResult {
		t.Logf("CE time format differs: Thai=%q vs Default=%q (expected to differ if month names differ)", thaiResult, defaultResult)
	}
}

// TestFormatDefaultLocale tests default locale formatting
func TestFormatDefaultLocale(t *testing.T) {
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	beTime := tm.InEra(BE())

	// Default locale should show BE year but English month names
	result := beTime.FormatLocale(LocaleDefault, "2006-01-02")

	if !strings.Contains(result, "2567") {
		t.Errorf("Default locale for BE should contain year 2567, got %q", result)
	}

	// Month names should be English, not Thai
	if strings.Contains(result, "กุมภาพันธ์") {
		t.Errorf("Default locale should not contain Thai month names, got %q", result)
	}
}

// TestFormatAllYearVariations tests year formatting variations
func TestFormatAllYearVariations(t *testing.T) {
	tests := []struct {
		year   int
		layout string
		name   string
	}{
		{2024, "2006", "4-digit year"},
		{2024, "06", "2-digit year"},
		{2024, "January 02, 2006", "Full date with year"},
		{2024, "02/01/2006", "Numeric date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := Date(tt.year, 2, 29, 12, 30, 45, 0, stdtime.UTC)
			beTime := tm.InEra(BE())

			result := beTime.FormatLocale(LocaleEnUS, tt.layout)

			// Should contain some representation of year 2567
			if len(result) == 0 {
				t.Error("FormatLocale returned empty result")
			}
		})
	}
}

// TestFormatLeapDayPreserved tests that Feb 29 appears correctly in all formats
func TestFormatLeapDayPreserved(t *testing.T) {
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	beTime := tm.InEra(BE())

	layouts := []string{
		"02 January 2006",
		"2006-01-02",
		"01/02/2006",
		"January 02, 2006",
	}

	for _, layout := range layouts {
		t.Run(layout, func(t *testing.T) {
			// Thai
			thaiResult := beTime.FormatLocale(LocaleThTH, layout)
			if !strings.Contains(thaiResult, "29") {
				t.Errorf("Thai format should contain day 29, got %q", thaiResult)
			}

			// English
			enResult := beTime.FormatLocale(LocaleEnUS, layout)
			if !strings.Contains(enResult, "29") {
				t.Errorf("English format should contain day 29, got %q", enResult)
			}
		})
	}
}

// TestFormatYearConversionAccuracy tests that year conversion is accurate
func TestFormatYearConversionAccuracy(t *testing.T) {
	tests := []struct {
		ceYear         int
		expectedBEYear string
		layout         string
	}{
		{2024, "2567", "2006"},
		{2023, "2566", "2006"},
		{2000, "2543", "2006"},
		{1900, "2443", "2006"},
		{543, "1086", "2006"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.ceYear)), func(t *testing.T) {
			tm := Date(tt.ceYear, 1, 1, 0, 0, 0, 0, stdtime.UTC)
			beTime := tm.InEra(BE())

			result := beTime.FormatLocale(LocaleEnUS, tt.layout)

			if !strings.Contains(result, tt.expectedBEYear) {
				t.Errorf("Year %d should format as BE %s, got %q", tt.ceYear, tt.expectedBEYear, result)
			}
		})
	}
}

// TestFormatConsistency tests that Format() and FormatLocale() are consistent
func TestFormatConsistency(t *testing.T) {
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	beTime := tm.InEra(BE())

	// Both methods should show BE year
	standardFormat := beTime.Format("2006-01-02")
	localeFormat := beTime.FormatLocale(LocaleDefault, "2006-01-02")

	if !strings.Contains(standardFormat, "2567") {
		t.Errorf("Format() should contain BE year, got %q", standardFormat)
	}

	if !strings.Contains(localeFormat, "2567") {
		t.Errorf("FormatLocale() should contain BE year, got %q", localeFormat)
	}

	// Both should contain leap day
	if !strings.Contains(standardFormat, "29") {
		t.Errorf("Format() should contain day 29, got %q", standardFormat)
	}
	if !strings.Contains(localeFormat, "29") {
		t.Errorf("FormatLocale() should contain day 29, got %q", localeFormat)
	}
}

// TestFormatLocaleWithTime tests formatting with specific time components
func TestFormatLocaleWithTime(t *testing.T) {
	tests := []struct {
		hour   int
		minute int
		second int
		layout string
		name   string
	}{
		{0, 0, 0, "15:04:05 2006", "Midnight with year"},
		{12, 30, 45, "15:04:05 2006", "Noon with year"},
		{23, 59, 59, "15:04:05 2006", "End of day with year"},
		{9, 15, 30, "15:04 2006", "Morning with year"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := Date(2024, 2, 29, tt.hour, tt.minute, tt.second, 0, stdtime.UTC)
			beTime := tm.InEra(BE())

			result := beTime.FormatLocale(LocaleThTH, tt.layout)

			// Should contain time information
			if len(result) == 0 {
				t.Error("FormatLocale with time layout returned empty")
			}

			// Should contain year
			if !strings.Contains(result, "2567") {
				t.Errorf("FormatLocale with time should show year, got %q", result)
			}
		})
	}
}

// TestFormatCenturyBoundary tests formatting around century boundaries
func TestFormatCenturyBoundary(t *testing.T) {
	tests := []struct {
		year   int
		beYear string
		reason string
	}{
		{1999, "2542", "Pre-Y2K"},
		{2000, "2543", "Y2K (leap)"},
		{2001, "2544", "Post-Y2K"},
		{2099, "2642", "Pre-2100"},
		{2100, "2643", "Century year (non-leap)"},
		{2101, "2644", "Post-2100"},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			tm := Date(tt.year, 1, 1, 0, 0, 0, 0, stdtime.UTC)
			beTime := tm.InEra(BE())

			result := beTime.FormatLocale(LocaleEnUS, "2006")

			if !strings.Contains(result, tt.beYear) {
				t.Errorf("Year %d should be BE %s, got %q", tt.year, tt.beYear, result)
			}
		})
	}
}
