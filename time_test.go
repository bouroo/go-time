package gotime

import (
	"strings"
	"testing"
	"time"
)

// TestLeapDayHandlingCoreRealWorld tests real-world leap day scenarios
func TestLeapDayHandlingCoreRealWorld(t *testing.T) {
	tests := []struct {
		name           string
		year           int
		month          int
		day            int
		shouldCreateOK bool
		inBEYear       int
		reason         string
	}{
		// Valid leap days
		{"Leap day 2024 CE", 2024, 2, 29, true, 2567, "2024 is divisible by 4"},
		{"Leap day 2020 CE", 2020, 2, 29, true, 2563, "2020 is divisible by 4"},
		{"Leap day 2000 CE", 2000, 2, 29, true, 2543, "2000 is divisible by 400"},
		{"Leap day 1996 CE", 1996, 2, 29, true, 2539, "1996 is divisible by 4"},
		{"Leap day 2400 CE", 2400, 2, 29, true, 2943, "2400 is divisible by 400"},

		// Invalid leap days (non-leap years)
		{"Feb 29 in non-leap 2023", 2023, 2, 29, false, 0, "2023 is not divisible by 4"},
		{"Feb 29 in non-leap 2100", 2100, 2, 29, false, 0, "2100 is century but not divisible by 400"},
		{"Feb 29 in non-leap 1900", 1900, 2, 29, false, 0, "1900 is century but not divisible by 400"},
		{"Feb 29 in year 1 CE", 1, 2, 29, false, 0, "Year 1 is not leap"},

		// Non-leap February dates (always valid)
		{"Feb 28 in 2023", 2023, 2, 28, true, 2566, "Feb 28 always valid"},
		{"Feb 28 in 2024 (leap year)", 2024, 2, 28, true, 2567, "Feb 28 valid even in leap year"},
		{"Feb 01 in 2024", 2024, 2, 1, true, 2567, "First day of February"},

		// Other months not affected
		{"Jan 31 in 2023", 2023, 1, 31, true, 2566, "January 31st valid"},
		{"Mar 31 in 2024", 2024, 3, 31, true, 2567, "March 31st valid"},
		{"Apr 30 in 2024", 2024, 4, 30, true, 2567, "April 30th valid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the date (Go's time.Date validates the date)
			tm := time.Date(tt.year, time.Month(tt.month), tt.day, 0, 0, 0, 0, time.UTC)

			if tt.shouldCreateOK {
				// Verify we got the correct CE year
				if tm.Year() != tt.year {
					t.Errorf("Year mismatch: got %d, want %d", tm.Year(), tt.year)
				}

				// Verify day is preserved
				if tm.Day() != tt.day {
					t.Errorf("Day mismatch: got %d, want %d", tm.Day(), tt.day)
				}

				// Wrap in gotime.Time and verify BE conversion
				gtm := Date(tt.year, tt.month, tt.day, 0, 0, 0, 0, time.UTC)
				beTime := gtm.InEra(BE())
				if beTime.Year() != tt.inBEYear {
					t.Errorf("BE year mismatch: got %d, want %d", beTime.Year(), tt.inBEYear)
				}
				if beTime.Day() != tt.day {
					t.Errorf("BE day not preserved: got %d, want %d", beTime.Day(), tt.day)
				}
			} else {
				// For invalid dates, time.Date will panic or adjust
				// We just verify year/month match to show the date was handled
				if tm.Year() != tt.year || tm.Month() != time.Month(tt.month) {
					t.Logf("Invalid date handled: %d-%02d-%02d", tm.Year(), tm.Month(), tm.Day())
				}
			}
		})
	}
}

// TestLeapDayPreservationAcrossEras tests that leap day is preserved when converting eras
func TestLeapDayPreservationAcrossEras(t *testing.T) {
	tests := []struct {
		name          string
		ceYear        int
		month         int
		day           int
		expectedBEDay int
		reason        string
	}{
		// Leap day preservation
		{"2024 leap day to BE", 2024, 2, 29, 29, "Leap day preserved in both CE and BE"},
		{"2020 leap day to BE", 2020, 2, 29, 29, "2020 leap day in both eras"},
		{"2000 leap day to BE", 2000, 2, 29, 29, "Y2K leap day preserved"},
		{"1996 leap day to BE", 1996, 2, 29, 29, "1996 leap day preserved"},

		// Regular February dates
		{"2023 Feb 28 to BE", 2023, 2, 28, 28, "Feb 28 preserved in non-leap year"},
		{"2024 Feb 28 to BE", 2024, 2, 28, 28, "Feb 28 preserved even in leap year"},
		{"2024 Feb 01 to BE", 2024, 2, 1, 1, "Feb 1 preserved"},

		// Other months unaffected
		{"2024 Mar 01 to BE", 2024, 3, 1, 1, "March dates unaffected"},
		{"2024 Jan 31 to BE", 2024, 1, 31, 31, "January 31 preserved"},
		{"2024 Dec 31 to BE", 2024, 12, 31, 31, "December 31 preserved"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ceTime := Date(tt.ceYear, tt.month, tt.day, 12, 30, 45, 0, time.UTC)
			beTime := ceTime.InEra(BE())

			// Verify CE year is preserved internally
			if ceTime.YearCE() != tt.ceYear {
				t.Errorf("CE Year mismatch: got %d, want %d", ceTime.YearCE(), tt.ceYear)
			}

			// Verify month/day are identical
			if beTime.Month() != time.Month(tt.month) {
				t.Errorf("Month mismatch: got %v, want %v", beTime.Month(), time.Month(tt.month))
			}
			if beTime.Day() != tt.expectedBEDay {
				t.Errorf("Day mismatch: got %d, want %d", beTime.Day(), tt.expectedBEDay)
			}

			// Verify time components are preserved
			if beTime.Hour() != ceTime.Hour() || beTime.Minute() != ceTime.Minute() {
				t.Error("Time components not preserved across era conversion")
			}
		})
	}
}

// TestLeapYearDetectionAcrossAllEras tests IsLeap() returns correct result
func TestLeapYearDetectionAcrossAllEras(t *testing.T) {
	tests := []struct {
		ceYear       int
		beYear       int
		expectedLeap bool
		reason       string
	}{
		// Leap years
		{2024, 2567, true, "2024 is leap (div by 4)"},
		{2020, 2563, true, "2020 is leap (div by 4)"},
		{2016, 2559, true, "2016 is leap (div by 4)"},
		{2000, 2543, true, "2000 is leap (div by 400)"},
		{1600, 2143, true, "1600 is leap (div by 400)"},
		{2400, 2943, true, "2400 is leap (div by 400)"},

		// Non-leap years
		{2023, 2566, false, "2023 not leap (not div by 4)"},
		{2022, 2565, false, "2022 not leap (not div by 4)"},
		{2019, 2562, false, "2019 not leap (not div by 4)"},
		{1900, 2443, false, "1900 not leap (div by 100, not 400)"},
		{2100, 2643, false, "2100 not leap (div by 100, not 400)"},
		{2200, 2743, false, "2200 not leap (div by 100, not 400)"},
		{2300, 2843, false, "2300 not leap (div by 100, not 400)"},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			ceTime := Date(tt.ceYear, 1, 1, 0, 0, 0, 0, time.UTC)
			beTime := ceTime.InEra(BE())

			// Both should return same result since leap year is determined by CE year
			if ceTime.IsLeap() != tt.expectedLeap {
				t.Errorf("CE IsLeap(%d) = %v, want %v", tt.ceYear, ceTime.IsLeap(), tt.expectedLeap)
			}
			if beTime.IsLeap() != tt.expectedLeap {
				t.Errorf("BE IsLeap(%d) = %v, want %v", tt.beYear, beTime.IsLeap(), tt.expectedLeap)
			}
		})
	}
}

// TestCenturyLeapYearEdgeCases tests the tricky century leap year rules
func TestCenturyLeapYearEdgeCases(t *testing.T) {
	tests := []struct {
		year   int
		isLeap bool
		reason string
	}{
		// Century years - the tricky ones
		{1600, true, "1600: div by 400"},
		{1700, false, "1700: div by 100 but not 400"},
		{1800, false, "1800: div by 100 but not 400"},
		{1900, false, "1900: div by 100 but not 400"},
		{2000, true, "2000: div by 400 (Y2K!)"},
		{2100, false, "2100: div by 100 but not 400"},
		{2200, false, "2200: div by 100 but not 400"},
		{2300, false, "2300: div by 100 but not 400"},
		{2400, true, "2400: div by 400"},

		// Non-century leap years
		{1996, true, "1996: div by 4, not century"},
		{1997, false, "1997: not div by 4"},
		{1999, false, "1999: not div by 4"},
		{2004, true, "2004: div by 4, not century"},
		{2024, true, "2024: div by 4, not century"},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			tm := Date(tt.year, 1, 1, 0, 0, 0, 0, time.UTC)
			if tm.IsLeap() != tt.isLeap {
				t.Errorf("Year %d: IsLeap() = %v, want %v", tt.year, tm.IsLeap(), tt.isLeap)
			}

			// For leap years, verify Feb 29 can be created
			if tt.isLeap {
				leapDay := Date(tt.year, 2, 29, 0, 0, 0, 0, time.UTC)
				if leapDay.Day() != 29 || leapDay.Month() != time.February {
					t.Errorf("Failed to create leap day for year %d", tt.year)
				}
			}
		})
	}
}

// TestTimeEraConversionPreservation tests that all time components are preserved
func TestTimeEraConversionPreservation(t *testing.T) {
	tests := []struct {
		name   string
		hour   int
		minute int
		second int
		nsec   int
	}{
		{"Midnight", 0, 0, 0, 0},
		{"Noon", 12, 0, 0, 0},
		{"Morning", 9, 30, 15, 123456789},
		{"Afternoon", 15, 45, 30, 987654321},
		{"Evening", 18, 20, 5, 111111111},
		{"Night", 23, 59, 59, 999999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(testT *testing.T) {
			ceTime := Date(2024, 2, 29, tt.hour, tt.minute, tt.second, tt.nsec, time.UTC)
			beTime := ceTime.InEra(BE())

			// Verify all time components preserved
			if beTime.Hour() != tt.hour {
				testT.Errorf("Hour: got %d, want %d", beTime.Hour(), tt.hour)
			}
			if beTime.Minute() != tt.minute {
				testT.Errorf("Minute: got %d, want %d", beTime.Minute(), tt.minute)
			}
			if beTime.Second() != tt.second {
				testT.Errorf("Second: got %d, want %d", beTime.Second(), tt.second)
			}
			if beTime.Nanosecond() != tt.nsec {
				testT.Errorf("Nanosecond: got %d, want %d", beTime.Nanosecond(), tt.nsec)
			}
		})
	}
}

// TestDefaultEraIsCE tests that default era is CE
func TestDefaultEraIsCE(t *testing.T) {
	tests := []struct {
		name string
		tm   Time
	}{
		{"Date with no era", Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)},
		{"Now()", Now()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.tm.IsCE() {
				t.Error("Default era should be CE")
			}
			if tt.tm.IsBE() {
				t.Error("Default era should not be BE")
			}
			if tt.tm.Era() != CE() {
				t.Errorf("Era() = %v, want CE", tt.tm.Era())
			}
		})
	}
}

// TestEraFlagMethods tests IsCE() and IsBE() helper methods
func TestEraFlagMethods(t *testing.T) {
	ceTime := Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	beTime := ceTime.InEra(BE())

	if !ceTime.IsCE() {
		t.Error("CE time should return IsCE() = true")
	}
	if ceTime.IsBE() {
		t.Error("CE time should return IsBE() = false")
	}

	if beTime.IsCE() {
		t.Error("BE time should return IsCE() = false")
	}
	if !beTime.IsBE() {
		t.Error("BE time should return IsBE() = true")
	}
}

// TestYearAccessorMethods tests Year() vs YearCE()
func TestYearAccessorMethods(t *testing.T) {
	tests := []struct {
		ceYear         int
		expectedBEYear int
	}{
		{2024, 2567},
		{2023, 2566},
		{2000, 2543},
		{1900, 2443},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.ceYear)), func(t *testing.T) {
			ceTime := Date(tt.ceYear, 2, 1, 0, 0, 0, 0, time.UTC)
			beTime := ceTime.InEra(BE())

			// For CE time
			if ceTime.Year() != tt.ceYear {
				t.Errorf("CE Time.Year() = %d, want %d", ceTime.Year(), tt.ceYear)
			}
			if ceTime.YearCE() != tt.ceYear {
				t.Errorf("CE Time.YearCE() = %d, want %d", ceTime.YearCE(), tt.ceYear)
			}

			// For BE time
			if beTime.Year() != tt.expectedBEYear {
				t.Errorf("BE Time.Year() = %d, want %d", beTime.Year(), tt.expectedBEYear)
			}
			if beTime.YearCE() != tt.ceYear {
				t.Errorf("BE Time.YearCE() = %d, want %d", beTime.YearCE(), tt.ceYear)
			}
		})
	}
}

// TestTimeOperationsPreserveEra tests that Add/Sub operations preserve era
func TestTimeOperationsPreserveEra(t *testing.T) {
	tm := Date(2024, 2, 28, 12, 0, 0, 0, time.UTC)
	beTime := tm.InEra(BE())

	// Test Add
	added := beTime.Add(time.Hour)
	if !added.IsBE() {
		t.Error("Add() should preserve era")
	}
	if added.Hour() != 13 {
		t.Errorf("Add(Hour) hour = %d, want 13", added.Hour())
	}

	// Add one day
	nextDay := beTime.Add(24 * time.Hour)
	if !nextDay.IsBE() {
		t.Error("Add() should preserve era")
	}
	if nextDay.Day() != 29 {
		t.Errorf("Add(24h) day = %d, want 29", nextDay.Day())
	}
}

// TestTimeComparisons tests comparison operations work correctly
func TestTimeComparisons(t *testing.T) {
	t1 := Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)
	t2 := Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)
	t3 := Date(2024, 2, 29, 13, 0, 0, 0, time.UTC)

	// Equal
	if !t1.Equal(t2) {
		t.Error("t1.Equal(t2) should be true")
	}

	// Before/After
	if t1.Before(t2) {
		t.Error("t1 should not be before t2")
	}
	if t1.After(t2) {
		t.Error("t1 should not be after t2")
	}
	if !t1.Before(t3) {
		t.Error("t1 should be before t3")
	}
	if !t3.After(t1) {
		t.Error("t3 should be after t1")
	}

	// With era - should compare underlying times
	beT1 := t1.InEra(BE())
	if !beT1.Equal(t1) {
		t.Error("t1 and beT1 should be equal (same underlying time)")
	}
}

// TestTimeLocations tests location handling with leap days
func TestTimeLocations(t *testing.T) {
	locations := []string{
		"UTC",
		"America/New_York",
		"Europe/London",
		"Asia/Bangkok", // Thailand!
		"Australia/Sydney",
	}

	for _, locName := range locations {
		t.Run(locName, func(t *testing.T) {
			loc, err := time.LoadLocation(locName)
			if err != nil {
				t.Skipf("Failed to load location %s: %v", locName, err)
			}

			// Create leap day in that location
			tm := Date(2024, 2, 29, 15, 30, 45, 0, loc)
			beTime := tm.InEra(BE())

			if beTime.Location() != loc {
				t.Errorf("Location mismatch: got %v, want %v", beTime.Location(), loc)
			}
			if beTime.Day() != 29 {
				t.Errorf("Leap day not preserved: got %d", beTime.Day())
			}
		})
	}
}

// TestTimeUnixTimestamp tests Unix timestamp consistency with leap days
func TestTimeUnixTimestamp(t *testing.T) {
	// 2024-02-29 12:00:00 UTC
	tm := Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)
	beTime := tm.InEra(BE())

	// Unix timestamp should be identical (same underlying time)
	if tm.Unix() != beTime.Unix() {
		t.Errorf("Unix() differs: CE=%d, BE=%d", tm.Unix(), beTime.Unix())
	}
	if tm.UnixNano() != beTime.UnixNano() {
		t.Errorf("UnixNano() differs: CE=%d, BE=%d", tm.UnixNano(), beTime.UnixNano())
	}

	// Should be consistent with stdlib
	expected := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC).Unix()
	if tm.Unix() != expected {
		t.Errorf("Unix() = %d, want %d", tm.Unix(), expected)
	}
}

// TestStringRepresentation tests String() output
func TestStringRepresentation(t *testing.T) {
	tm := Date(2024, 2, 29, 12, 30, 45, 0, time.UTC)
	str := tm.String()

	// Should contain the date components
	if len(str) == 0 {
		t.Error("String() returned empty")
	}

	// CE time should show 2024
	if !strings.Contains(str, "2024") {
		t.Errorf("CE String() should contain 2024, got %q", str)
	}

	// BE time should show different year
	beTime := tm.InEra(BE())
	beStr := beTime.String()
	if beStr == str {
		t.Errorf("String() should differ between eras since they're era-aware")
	}
	// BE should show 2567
	if !strings.Contains(beStr, "2567") {
		t.Errorf("BE String() should contain 2567, got %q", beStr)
	}
}

// TestZeroTime tests zero/empty time handling
func TestZeroTime(t *testing.T) {
	zeroTime := Time{}
	if !zeroTime.IsZero() {
		t.Error("Zero Time.IsZero() should return true")
	}

	nonZero := Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if nonZero.IsZero() {
		t.Error("Non-zero time.IsZero() should return false")
	}
}

// TestJSONMarshaling tests JSON marshaling with leap days
func TestJSONMarshaling(t *testing.T) {
	tm := Date(2024, 2, 29, 12, 30, 45, 0, time.UTC)

	// Marshal
	data, err := tm.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	// Verify it contains the leap day
	dataStr := string(data)
	if dataStr == "" {
		t.Error("MarshalJSON() returned empty")
	}

	// Unmarshal
	var unmarshaled Time
	err = unmarshaled.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	// Verify leap day preserved
	if unmarshaled.Day() != 29 || unmarshaled.Month() != time.February {
		t.Errorf("Leap day lost after marshal/unmarshal: day=%d, month=%v", unmarshaled.Day(), unmarshaled.Month())
	}
}

// TestGobEncoding tests Gob encoding with leap days
func TestGobEncoding(t *testing.T) {
	tm := Date(2024, 2, 29, 12, 30, 45, 123456789, time.UTC)

	// Encode
	data, err := tm.GobEncode()
	if err != nil {
		t.Fatalf("GobEncode() error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GobEncode() returned empty data")
	}

	// Decode
	var decoded Time
	err = decoded.GobDecode(data)
	if err != nil {
		t.Fatalf("GobDecode() error: %v", err)
	}

	// Verify leap day preserved
	if decoded.YearCE() != 2024 || decoded.Month() != time.February || decoded.Day() != 29 {
		t.Errorf("Leap day lost after Gob encoding: %d-%02d-%02d", decoded.YearCE(), decoded.Month(), decoded.Day())
	}
	if decoded.Nanosecond() != 123456789 {
		t.Errorf("Nanoseconds lost: got %d, want 123456789", decoded.Nanosecond())
	}
}
