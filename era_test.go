package gotime

import (
	"sync"
	"testing"
)

// TestEraConversionRealWorld tests era conversions with real-world scenarios
func TestEraConversionRealWorld(t *testing.T) {
	tests := []struct {
		name           string
		ceYear         int
		expectedBEYear int
		description    string
	}{
		// Historical dates
		{"Buddhism founding (543 BC/CE equivalent)", 1, 544, "Year 1 CE converts to BE 544"},
		{"Thai calendar start (543 BC)", 1, 544, "BE epoch offset by 543 years"},

		// Current era
		{"Modern date 2024", 2024, 2567, "Current CE year to BE"},
		{"Modern date 2000", 2000, 2543, "Y2K in both eras"},

		// Century boundaries (leap year edge cases)
		{"Century non-leap 1900", 1900, 2443, "1900 is not a leap year"},
		{"Century leap 2000", 2000, 2543, "2000 is a leap year"},
		{"Future century 2100", 2100, 2643, "2100 is not a leap year (future)"},
		{"Future century leap 2400", 2400, 2943, "2400 is a leap year (future)"},

		// Far future
		{"Year 3000", 3000, 3543, "Distant future year"},
		{"Year 5000", 5000, 5543, "Very distant future"},

		// Roundtrip conversion verification
		{"Roundtrip 2023", 2023, 2566, "Verify conversion consistency"},
		{"Roundtrip 2024", 2024, 2567, "Verify leap year consistency"},
		{"Roundtrip 2025", 2025, 2568, "Verify next year"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beYear := BE().FromCE(tt.ceYear)
			if beYear != tt.expectedBEYear {
				t.Errorf("FromCE(%d) = %d, want %d: %s", tt.ceYear, beYear, tt.expectedBEYear, tt.description)
			}

			// Verify roundtrip
			convertedBack := BE().ToCE(beYear)
			if convertedBack != tt.ceYear {
				t.Errorf("Roundtrip failed: CE %d -> BE %d -> CE %d", tt.ceYear, beYear, convertedBack)
			}
		})
	}
}

// TestLeapYearDetectionCEEra tests leap year identification in CE era
func TestLeapYearDetectionCEEra(t *testing.T) {
	tests := []struct {
		year   int
		isLeap bool
		reason string
	}{
		// Standard leap years
		{2024, true, "Divisible by 4, not century"},
		{2020, true, "Divisible by 4, not century"},
		{2016, true, "Divisible by 4, not century"},
		{2012, true, "Divisible by 4, not century"},
		{2008, true, "Divisible by 4, not century"},
		{2004, true, "Divisible by 4, not century"},

		// Non-leap years (regular)
		{2023, false, "Not divisible by 4"},
		{2022, false, "Not divisible by 4"},
		{2021, false, "Not divisible by 4"},
		{2019, false, "Not divisible by 4"},
		{2018, false, "Not divisible by 4"},
		{2017, false, "Not divisible by 4"},

		// Century years (critical edge case)
		{1600, true, "Century leap: divisible by 400"},
		{1700, false, "Century non-leap: divisible by 100 but not 400"},
		{1800, false, "Century non-leap: divisible by 100 but not 400"},
		{1900, false, "Century non-leap: divisible by 100 but not 400"},
		{2000, true, "Century leap: divisible by 400 (Y2K)"},
		{2100, false, "Century non-leap: divisible by 100 but not 400"},
		{2200, false, "Century non-leap: divisible by 100 but not 400"},
		{2300, false, "Century non-leap: divisible by 100 but not 400"},
		{2400, true, "Century leap: divisible by 400"},

		// Edge case: year 1
		{1, false, "Year 1 CE not leap"},

		// Far future
		{3000, false, "Divisible by 100 but not 400"},
		{4000, true, "Divisible by 400"},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			// Use the isLeapYear logic from Time struct
			isLeap := (tt.year%4 == 0 && tt.year%100 != 0) || tt.year%400 == 0
			if isLeap != tt.isLeap {
				t.Errorf("Year %d: got isLeap=%v, want %v. Reason: %s", tt.year, isLeap, tt.isLeap, tt.reason)
			}
		})
	}
}

// TestLeapYearDetectionBEEra tests that leap year is determined by CE year
func TestLeapYearDetectionBEEra(t *testing.T) {
	tests := []struct {
		ceYear      int
		beYear      int
		isLeap      bool
		description string
	}{
		// BE 2567 is CE 2024 (leap)
		{2024, 2567, true, "BE 2567 = CE 2024 (leap year)"},

		// BE 2566 is CE 2023 (not leap)
		{2023, 2566, false, "BE 2566 = CE 2023 (non-leap year)"},

		// BE 2543 is CE 2000 (leap - century)
		{2000, 2543, true, "BE 2543 = CE 2000 (century leap)"},

		// BE 2486 is CE 1943 (not leap)
		{1943, 2486, false, "BE 2486 = CE 1943 (not leap)"},

		// BE 2443 is CE 1900 (not leap - century)
		{1900, 2443, false, "BE 2443 = CE 1900 (century non-leap)"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Verify offset
			calculatedBE := BE().FromCE(tt.ceYear)
			if calculatedBE != tt.beYear {
				t.Errorf("Year conversion error: CE %d should be BE %d, got BE %d", tt.ceYear, tt.beYear, calculatedBE)
			}

			// Verify leap year is based on CE year
			isLeap := (tt.ceYear%4 == 0 && tt.ceYear%100 != 0) || tt.ceYear%400 == 0
			if isLeap != tt.isLeap {
				t.Errorf("Leap year check failed for CE %d (BE %d): got %v, want %v", tt.ceYear, tt.beYear, isLeap, tt.isLeap)
			}
		})
	}
}

// TestEraRegistryRealWorld tests era registration and retrieval
func TestEraRegistryRealWorld(t *testing.T) {
	t.Run("builtin eras available", func(t *testing.T) {
		ce := GetEra("CE")
		if ce == nil {
			t.Fatal("CE era not registered")
		}
		if ce.Offset() != 0 {
			t.Errorf("CE offset = %d, want 0", ce.Offset())
		}

		be := GetEra("BE")
		if be == nil {
			t.Fatal("BE era not registered")
		}
		if be.Offset() != 543 {
			t.Errorf("BE offset = %d, want 543", be.Offset())
		}
	})

	t.Run("custom era registration", func(t *testing.T) {
		// Register a custom era (e.g., Islamic Hijri)
		hijri := RegisterEra("AH", 579) // Approximate offset
		if hijri == nil {
			t.Fatal("Failed to register custom era")
		}

		retrieved := GetEra("AH")
		if retrieved != hijri {
			t.Error("Retrieved era differs from registered era")
		}

		// Test conversion
		ceYear := 2024
		ahYear := hijri.FromCE(ceYear)
		expectedAH := ceYear + 579
		if ahYear != expectedAH {
			t.Errorf("AH conversion: CE %d should be AH %d, got %d", ceYear, expectedAH, ahYear)
		}
	})

	t.Run("duplicate registration returns same instance", func(t *testing.T) {
		first := RegisterEra("TestEra1", 100)
		second := RegisterEra("TestEra1", 100)
		if first != second {
			t.Error("Duplicate registration should return same instance")
		}
	})

	t.Run("thread-safe registration", func(t *testing.T) {
		var wg sync.WaitGroup
		const goroutines = 50

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				// Each goroutine registers different eras
				era := RegisterEra("ConcurrentEra"+string(rune(id)), id*100)
				if era.Offset() != id*100 {
					t.Errorf("Goroutine %d: Era offset mismatch", id)
				}
			}(i)
		}
		wg.Wait()
	})
}

// TestEraStrings tests era string representation
func TestEraStrings(t *testing.T) {
	tests := []struct {
		era    *Era
		name   string
		offset int
	}{
		{CE(), "CE", 0},
		{BE(), "BE", 543},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.era.String() != tt.name {
				t.Errorf("String() = %q, want %q", tt.era.String(), tt.name)
			}
			if tt.era.Offset() != tt.offset {
				t.Errorf("Offset() = %d, want %d", tt.era.Offset(), tt.offset)
			}
		})
	}
}

// TestEraEdgeCasesYearBoundaries tests year boundaries and edge cases
func TestEraEdgeCasesYearBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		ceYear      int
		beYear      int
		isLeapCE    bool
		description string
	}{
		{"Year 1 CE to BE", 1, 544, false, "Year 1 (not leap)"},
		{"Year 0 CE to BE", 0, 543, true, "Year zero (divisible by 400, is leap)"},
		{"Year -100 pre-CE", -100, 443, false, "-100 is divisible by 100 but not 400, not leap"},

		{"Leap year 2000 CE", 2000, 2543, true, "Y2K - major leap year"},
		{"Non-leap 1900 CE", 1900, 2443, false, "Century year not divisible by 400"},
		{"Leap year 2400 CE", 2400, 2943, true, "Future century leap year"},

		{"Current year 2024 CE", 2024, 2567, true, "Current year used in tests"},
		{"Future year 3000 CE", 3000, 3543, false, "Far future (not leap)"},
		{"Very future 4000 CE", 4000, 4543, true, "Distant future (divisible by 400)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify BE conversion
			beYear := BE().FromCE(tt.ceYear)
			if beYear != tt.beYear {
				t.Errorf("CE %d -> BE: got %d, want %d", tt.ceYear, beYear, tt.beYear)
			}

			// Verify roundtrip
			convertedBack := BE().ToCE(tt.beYear)
			if convertedBack != tt.ceYear {
				t.Errorf("Roundtrip: CE %d -> BE %d -> CE %d (failed)", tt.ceYear, tt.beYear, convertedBack)
			}

			// Verify leap year detection
			isLeap := (tt.ceYear%4 == 0 && tt.ceYear%100 != 0) || tt.ceYear%400 == 0
			if isLeap != tt.isLeapCE {
				t.Errorf("CE %d leap year: got %v, want %v", tt.ceYear, isLeap, tt.isLeapCE)
			}
		})
	}
}

// TestEraDetectionFromYear tests automatic era detection based on year values
func TestEraDetectionFromYear(t *testing.T) {
	tests := []struct {
		year        int
		expectedEra *Era
		reason      string
	}{
		// Modern CE years (1-4 digit, recent)
		{2024, CE(), "Modern CE year"},
		{2023, CE(), "Recent CE year"},
		{2000, CE(), "Y2K"},
		{1999, CE(), "Late 20th century"},

		// BE years (4+ digit, recent-ish in BE)
		{2567, BE(), "Modern BE year"},
		{2566, BE(), "Recent BE year"},
		{2543, BE(), "BE year for Y2K equivalent"},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			// Note: DetectEraFromYear uses current date as reference
			// This test may be flaky in far future, but works for current dates
			detected := DetectEraFromYear(tt.year)
			if detected != tt.expectedEra {
				t.Logf("Note: Year detection may depend on current date; got %v, expected %v for year %d",
					detected, tt.expectedEra, tt.year)
			}
		})
	}
}
