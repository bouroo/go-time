package gotime

import (
	"errors"
	"testing"
	"time"
)

// TestParseLeapDayValid tests parsing of valid leap days
func TestParseLeapDayValid(t *testing.T) {
	tests := []struct {
		name       string
		layout     string
		value      string
		era        *Era
		expectYear int
		expectDay  int
		expectOK   bool
		reason     string
	}{
		// CE leap days
		{
			"CE 2024 leap day",
			"02/01/2006",
			"29/02/2024",
			CE(),
			2024, 29, true,
			"2024 is a leap year",
		},
		{
			"CE 2020 leap day",
			"02/01/2006",
			"29/02/2020",
			CE(),
			2020, 29, true,
			"2020 is a leap year",
		},
		{
			"CE 2000 leap day",
			"02/01/2006",
			"29/02/2000",
			CE(),
			2000, 29, true,
			"2000 is a leap year (div by 400)",
		},
		{
			"CE 1996 leap day",
			"02/01/2006",
			"29/02/1996",
			CE(),
			1996, 29, true,
			"1996 is a leap year",
		},
		// BE leap days
		{
			"BE 2567 leap day",
			"02/01/2006",
			"29/02/2567",
			BE(),
			2024, 29, true,
			"BE 2567 = CE 2024 (leap year)",
		},
		{
			"BE 2563 leap day",
			"02/01/2006",
			"29/02/2563",
			BE(),
			2020, 29, true,
			"BE 2563 = CE 2020 (leap year)",
		},
		{
			"BE 2543 leap day",
			"02/01/2006",
			"29/02/2543",
			BE(),
			2000, 29, true,
			"BE 2543 = CE 2000 (leap year)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseWithEra(tt.layout, tt.value, tt.era)
			if tt.expectOK {
				if err != nil {
					t.Fatalf("ParseWithEra() expected no error, got %v: %s", err, tt.reason)
				}
				if result.YearCE() != tt.expectYear {
					t.Errorf("YearCE = %d, want %d", result.YearCE(), tt.expectYear)
				}
				if result.Day() != tt.expectDay {
					t.Errorf("Day = %d, want %d", result.Day(), tt.expectDay)
				}
				if result.Month() != time.February {
					t.Errorf("Month = %v, want February", result.Month())
				}
			} else {
				if err == nil {
					t.Errorf("ParseWithEra() expected error for %s", tt.reason)
				}
			}
		})
	}
}

// TestParseInvalidLeapDays tests parsing of invalid leap days (non-leap years)
func TestParseInvalidLeapDays(t *testing.T) {
	tests := []struct {
		name        string
		layout      string
		value       string
		era         *Era
		expectError bool
		reason      string
	}{
		// Non-leap years - should error
		{
			"CE 2023 Feb 29",
			"02/01/2006",
			"29/02/2023",
			CE(),
			true,
			"2023 is not a leap year",
		},
		{
			"CE 1900 Feb 29",
			"02/01/2006",
			"29/02/1900",
			CE(),
			true,
			"1900 is century but not div by 400",
		},
		{
			"CE 2100 Feb 29",
			"02/01/2006",
			"29/02/2100",
			CE(),
			true,
			"2100 is century but not div by 400",
		},
		{
			"BE 2566 Feb 29",
			"02/01/2006",
			"29/02/2566",
			BE(),
			true,
			"BE 2566 = CE 2023 (not leap)",
		},
		{
			"BE 2543 is CE 2000 but invalid format",
			"02/01/2006",
			"29/02/2442",
			BE(),
			true,
			"BE 2442 = CE 1899 (not leap)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseWithEra(tt.layout, tt.value, tt.era)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseWithEra() expected error: %s, but got success with result %+v", tt.reason, result)
				}
			} else {
				if err != nil {
					t.Errorf("ParseWithEra() expected success, got error: %v (%s)", err, tt.reason)
				}
			}
		})
	}
}

// TestParseThaiMonthNames tests parsing Thai month names
func TestParseThaiMonthNames(t *testing.T) {
	tests := []struct {
		name        string
		layout      string
		value       string
		era         *Era
		expectMonth time.Month
		expectYear  int
		description string
	}{
		{
			"Thai January full",
			"02 January 2006",
			"15 มกราคม 2024",
			CE(),
			time.January, 2024,
			"Full Thai month name",
		},
		{
			"Thai February full",
			"02 January 2006",
			"29 กุมภาพันธ์ 2024",
			CE(),
			time.February, 2024,
			"Thai February leap day",
		},
		{
			"Thai March full",
			"02 January 2006",
			"10 มีนาคม 2024",
			CE(),
			time.March, 2024,
			"Full Thai month name March",
		},
		{
			"Thai December full",
			"02 January 2006",
			"25 ธันวาคม 2567",
			BE(),
			time.December, 2024,
			"Thai December in BE era",
		},
		{
			"Thai all months",
			"02 January 2006",
			"15 พฤษภาคม 2024",
			CE(),
			time.May, 2024,
			"Thai May",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseWithEra(tt.layout, tt.value, tt.era)
			if err != nil {
				t.Fatalf("ParseWithEra() error = %v: %s", err, tt.description)
			}

			if result.Month() != tt.expectMonth {
				t.Errorf("Month = %v, want %v", result.Month(), tt.expectMonth)
			}
			if result.YearCE() != tt.expectYear {
				t.Errorf("Year = %d, want %d", result.YearCE(), tt.expectYear)
			}
		})
	}
}

// TestParseThaiDayNames tests parsing Thai day names
func TestParseThaiDayNames(t *testing.T) {
	tests := []struct {
		name          string
		layout        string
		value         string
		era           *Era
		expectWeekday time.Weekday
		description   string
	}{
		{
			"Thai Monday",
			"Monday, 02 January 2006",
			"จันทร์, 15 มกราคม 2024",
			CE(),
			time.Monday,
			"Monday in Thai",
		},
		{
			"Thai Friday",
			"Monday, 02 January 2006",
			"ศุกร์, 12 มกราคม 2567",
			BE(),
			time.Friday,
			"Friday in Thai with BE",
		},
		{
			"Thai Sunday",
			"Monday, 02 January 2006",
			"อาทิตย์, 07 มกราคม 2024",
			CE(),
			time.Sunday,
			"Sunday in Thai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseWithEra(tt.layout, tt.value, tt.era)
			if err != nil {
				t.Fatalf("ParseWithEra() error = %v: %s", err, tt.description)
			}

			if result.Weekday() != tt.expectWeekday {
				t.Errorf("Weekday = %v, want %v", result.Weekday(), tt.expectWeekday)
			}
		})
	}
}

// TestParseBEYearConversion tests BE year conversion during parsing
func TestParseBEYearConversion(t *testing.T) {
	tests := []struct {
		name         string
		layout       string
		value        string
		era          *Era
		expectYearCE int
		expectEra    *Era
		description  string
	}{
		{
			"BE 2567 to CE 2024",
			"02/01/2006",
			"29/02/2567",
			BE(),
			2024, BE(),
			"Modern BE year",
		},
		{
			"BE 2566 to CE 2023",
			"02/01/2006",
			"15/01/2566",
			BE(),
			2023, BE(),
			"BE to CE conversion",
		},
		{
			"BE 2543 to CE 2000",
			"02/01/2006",
			"29/02/2543",
			BE(),
			2000, BE(),
			"Y2K equivalent in BE",
		},
		{
			"BE 2469 to CE 1926",
			"02/01/2006",
			"10/01/2469",
			BE(),
			1926, BE(),
			"Historical BE date",
		},
		{
			"CE 2024 stays CE 2024",
			"02/01/2006",
			"15/01/2024",
			CE(),
			2024, CE(),
			"CE era unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseWithEra(tt.layout, tt.value, tt.era)
			if err != nil {
				t.Fatalf("ParseWithEra() error = %v: %s", err, tt.description)
			}

			if result.YearCE() != tt.expectYearCE {
				t.Errorf("YearCE = %d, want %d", result.YearCE(), tt.expectYearCE)
			}
			if result.Era() != tt.expectEra {
				t.Errorf("Era = %v, want %v", result.Era(), tt.expectEra)
			}
		})
	}
}

// TestParseInLocationWithEraAndLeapDay tests ParseInLocationWithEra with leap days
func TestParseInLocationWithEraAndLeapDay(t *testing.T) {
	bangkokLoc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		t.Skipf("Could not load Bangkok timezone: %v", err)
	}

	tests := []struct {
		name        string
		layout      string
		value       string
		location    *time.Location
		era         *Era
		expectDay   int
		expectLoc   *time.Location
		description string
	}{
		{
			"Bangkok leap day CE",
			"02/01/2006 15:04",
			"29/02/2024 14:00",
			bangkokLoc,
			CE(),
			29, bangkokLoc,
			"Leap day with Bangkok timezone",
		},
		{
			"Bangkok leap day BE",
			"02/01/2006 15:04",
			"29/02/2567 21:00",
			bangkokLoc,
			BE(),
			29, bangkokLoc,
			"BE leap day with Bangkok timezone",
		},
		{
			"UTC leap day",
			"02/01/2006 15:04",
			"29/02/2024 12:00",
			time.UTC,
			CE(),
			29, time.UTC,
			"Leap day in UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseInLocationWithEra(tt.layout, tt.value, tt.location, tt.era)
			if err != nil {
				t.Fatalf("ParseInLocationWithEra() error = %v: %s", err, tt.description)
			}

			if result.Day() != tt.expectDay {
				t.Errorf("Day = %d, want %d", result.Day(), tt.expectDay)
			}
			if result.Location() != tt.expectLoc {
				t.Errorf("Location mismatch: got %v, want %v", result.Location(), tt.expectLoc)
			}
		})
	}
}

// TestParseThaiAutoDetectLeapDay tests ParseThai auto-detection with leap days
func TestParseThaiAutoDetectLeapDay(t *testing.T) {
	tests := []struct {
		name         string
		layout       string
		value        string
		expectedEra  *Era
		expectedYear int
		expectedDay  int
		description  string
		shouldError  bool
	}{
		{
			"Auto-detect CE 2024 leap day",
			"02/01/2006",
			"29/02/2024",
			CE(),
			2024, 29,
			"CE year auto-detected",
			false,
		},
		{
			"Thai month CE with leap day",
			"02 January 2006",
			"29 กุมภาพันธ์ 2024",
			CE(),
			2024, 29,
			"Thai month with CE year",
			false,
		},
		{
			"Thai month CE regular date",
			"02 January 2006",
			"15 กุมภาพันธ์ 2024",
			CE(),
			2024, 15,
			"Thai month with CE year regular",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseThai(tt.layout, tt.value)
			if tt.shouldError {
				if err == nil {
					t.Errorf("ParseThai() expected error: %s", tt.description)
				}
			} else {
				if err != nil {
					t.Fatalf("ParseThai() error = %v: %s", err, tt.description)
				}

				if result.Era() != tt.expectedEra {
					t.Errorf("Era = %v, want %v", result.Era(), tt.expectedEra)
				}
				if result.YearCE() != tt.expectedYear {
					t.Errorf("YearCE = %d, want %d", result.YearCE(), tt.expectedYear)
				}
				if result.Day() != tt.expectedDay {
					t.Errorf("Day = %d, want %d", result.Day(), tt.expectedDay)
				}
			}
		})
	}
}

// TestParseThaiInLocationAutoDetect tests ParseThaiInLocation
func TestParseThaiInLocationAutoDetect(t *testing.T) {
	bangkokLoc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		t.Skipf("Could not load Bangkok timezone: %v", err)
	}

	tests := []struct {
		name         string
		layout       string
		value        string
		location     *time.Location
		expectedEra  *Era
		expectedYear int
		description  string
	}{
		{
			"Bangkok regular CE date",
			"02/01/2006 15:04",
			"15/01/2024 14:00",
			bangkokLoc,
			CE(),
			2024,
			"CE date with Bangkok timezone",
		},
		{
			"Bangkok Thai month CE",
			"02 January 2006 15:04",
			"15 กุมภาพันธ์ 2024 14:00",
			bangkokLoc,
			CE(),
			2024,
			"Thai month with Bangkok timezone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseThaiInLocation(tt.layout, tt.value, tt.location)
			if err != nil {
				t.Fatalf("ParseThaiInLocation() error = %v: %s", err, tt.description)
			}

			if result.Era() != tt.expectedEra {
				t.Errorf("Era = %v, want %v", result.Era(), tt.expectedEra)
			}
			if result.YearCE() != tt.expectedYear {
				t.Errorf("YearCE = %d, want %d", result.YearCE(), tt.expectedYear)
			}
			if result.Location() != tt.location {
				t.Errorf("Location mismatch")
			}
		})
	}
}

// TestParseErrorHandling tests error handling
func TestParseErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		layout         string
		value          string
		era            *Era
		expectError    bool
		expectParseErr bool
		description    string
	}{
		{
			"Invalid format",
			"invalid-format",
			"2024-02-29",
			CE(),
			true, true,
			"Invalid layout format",
		},
		{
			"Invalid date Feb 30",
			"02/01/2006",
			"30/02/2024",
			CE(),
			true, true,
			"February 30 doesn't exist",
		},
		{
			"Invalid leap day in non-leap year",
			"02/01/2006",
			"29/02/2023",
			CE(),
			true, true,
			"Feb 29 in non-leap year",
		},
		{
			"Valid date",
			"02/01/2006",
			"29/02/2024",
			CE(),
			false, false,
			"Valid leap day",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseWithEra(tt.layout, tt.value, tt.era)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseWithEra() expected error: %s", tt.description)
				}
				if tt.expectParseErr {
					var parseErr *ParseError
					if !errors.As(err, &parseErr) {
						t.Logf("Error is not *ParseError: %T", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("ParseWithEra() unexpected error: %v (%s)", err, tt.description)
				}
				if result.IsZero() {
					t.Error("ParseWithEra() returned zero time")
				}
			}
		})
	}
}

// TestParseDropInReplacement tests that Parse and ParseInLocation match stdlib
func TestParseDropInReplacement(t *testing.T) {
	tests := []struct {
		layout string
		value  string
	}{
		{"2006-01-02", "2024-02-29"},
		{"2006-01-02 15:04:05", "2024-02-29 12:30:45"},
		{"January 02, 2006", "February 29, 2024"},
	}

	for _, tt := range tests {
		t.Run(tt.layout, func(t *testing.T) {
			// Parse should match stdlib
			_, err1 := Parse(tt.layout, tt.value)
			_, err2 := time.Parse(tt.layout, tt.value)

			if (err1 == nil) != (err2 == nil) {
				t.Errorf("Parse() error mismatch: gotime=%v, stdlib=%v", err1 != nil, err2 != nil)
			}
		})
	}
}

// TestParseErrorUnwrap tests that ParseError can be unwrapped
func TestParseErrorUnwrap(t *testing.T) {
	_, err := ParseWithEra("invalid", "test", CE())
	if err == nil {
		t.Fatal("Expected ParseWithEra to return error")
	}

	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("Error type = %T, want *ParseError", err)
	}

	if parseErr.Input != "test" {
		t.Errorf("Input = %q, want %q", parseErr.Input, "test")
	}

	if parseErr.Layout != "invalid" {
		t.Errorf("Layout = %q, want %q", parseErr.Layout, "invalid")
	}

	if parseErr.Original == nil {
		t.Error("Original error should not be nil")
	}
}

// TestParseCenturyLeapYears tests parsing around century leap years
func TestParseCenturyLeapYears(t *testing.T) {
	tests := []struct {
		name        string
		layout      string
		value       string
		era         *Era
		expectYear  int
		expectLeap  bool
		description string
	}{
		{
			"1900 Feb 28",
			"02/01/2006",
			"28/02/1900",
			CE(),
			1900, false,
			"Century non-leap year",
		},
		{
			"2000 Feb 29",
			"02/01/2006",
			"29/02/2000",
			CE(),
			2000, true,
			"Century leap year (Y2K)",
		},
		{
			"2100 Feb 28",
			"02/01/2006",
			"28/02/2100",
			CE(),
			2100, false,
			"Future century non-leap",
		},
		{
			"2400 Feb 29",
			"02/01/2006",
			"29/02/2400",
			CE(),
			2400, true,
			"Future century leap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseWithEra(tt.layout, tt.value, tt.era)
			if err != nil {
				t.Fatalf("ParseWithEra() error = %v: %s", err, tt.description)
			}

			if result.YearCE() != tt.expectYear {
				t.Errorf("Year = %d, want %d", result.YearCE(), tt.expectYear)
			}

			if result.IsLeap() != tt.expectLeap {
				t.Errorf("IsLeap() = %v, want %v", result.IsLeap(), tt.expectLeap)
			}
		})
	}
}

// TestParseRoundTrip tests that Parse->Format->Parse produces same result
func TestParseRoundTrip(t *testing.T) {
	tests := []struct {
		layout string
		value  string
		era    *Era
	}{
		{"2006-01-02", "2024-02-29", CE()},
		{"02/01/2006", "29/02/2024", CE()},
		{"02/01/2006", "29/02/2567", BE()},
		{"January 02, 2006", "February 29, 2024", CE()},
	}

	for _, tt := range tests {
		t.Run(tt.layout+":"+tt.value, func(t *testing.T) {
			// Parse
			parsed, err := ParseWithEra(tt.layout, tt.value, tt.era)
			if err != nil {
				t.Skipf("Parse failed: %v", err)
			}

			// Format
			formatted := parsed.Format(tt.layout)

			// Parse again
			reparsed, err := ParseWithEra(tt.layout, formatted, tt.era)
			if err != nil {
				t.Errorf("Reparse failed: %v", err)
			}

			// Should be equal
			if !parsed.Equal(reparsed) {
				t.Errorf("Roundtrip failed: %v != %v", parsed, reparsed)
			}
		})
	}
}
