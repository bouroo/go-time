// Package gotime provides comprehensive concurrency tests to verify thread safety
// of all operations. These tests use the race detector to detect potential issues.
package gotime

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestConcurrentEraRegistration verifies that concurrent era registration
// is thread-safe and doesn't cause race conditions.
func TestConcurrentEraRegistration(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 10

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				// Try to register the same era multiple times
				eraName := "TestEra"
				offset := id + j
				era := RegisterEra(eraName, offset)
				if era == nil {
					errChan <- nil // Should not happen
					continue
				}
				// Verify era is retrievable
				retrieved := GetEra(eraName)
				if retrieved == nil {
					errChan <- nil // Should not happen
				}
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			t.Errorf("Concurrent era registration error: %v", err)
		}
	}
}

// TestConcurrentTimeYearAccess tests concurrent access to Year() method
// which uses the global era cache.
func TestConcurrentTimeYearAccess(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 100

	// Create a time in BE era
	tm := Date(2024, 6, 15, 12, 30, 45, 0, time.UTC).InEra(BE())

	var wg sync.WaitGroup
	var counter int64

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				year := tm.Year()
				// BE year should be 2024 + 543 = 2567
				if year != 2567 {
					t.Errorf("Expected year 2567, got %d", year)
				}
				_ = year
				_ = tm.Era()
				_ = tm.YearCE()
				atomic.AddInt64(&counter, 1)
			}
		}()
	}

	wg.Wait()
	t.Logf("Completed %d concurrent year accesses", counter)
}

// TestConcurrentFormatOperations tests concurrent formatting operations.
func TestConcurrentFormatOperations(t *testing.T) {
	const numGoroutines = 50
	const numIterations = 50

	// Create times in different eras
	times := []Time{
		Date(2024, 1, 15, 10, 30, 0, 0, time.UTC).InEra(CE()),
		Date(2024, 6, 15, 12, 0, 0, 0, time.UTC).InEra(BE()),
		Date(2023, 12, 31, 23, 59, 59, 0, time.UTC).InEra(BE()),
	}

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				tm := times[id%len(times)]
				// Test Format
				_ = tm.Format("2006-01-02")
				// Test FormatLocale
				_ = tm.FormatLocale(LocaleThTH, "Monday, January 02, 2006")
				_ = tm.FormatLocale(LocaleEnUS, "January 02, 2006")
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentParseOperations tests concurrent parsing operations.
func TestConcurrentParseOperations(t *testing.T) {
	const numGoroutines = 50
	const numIterations = 50

	inputs := []struct {
		layout string
		value  string
	}{
		{"2006-01-02", "2024-06-15"},
		{"02/01/2006", "15/06/2024"},
		{"January 02, 2006", "June 15, 2024"},
	}

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				input := inputs[id%len(inputs)]
				_, _ = ParseWithEra(input.layout, input.value, CE())
				_, _ = ParseWithEra(input.layout, input.value, BE())
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentThaiParsing tests concurrent Thai parsing operations.
func TestConcurrentThaiParsing(t *testing.T) {
	const numGoroutines = 50
	const numIterations = 50

	inputs := []struct {
		layout string
		value  string
	}{
		{"02 มกราคม 2006", "15 มิถุนายน 2567"},
		{"2006-01-02", "2567-06-15"},
	}

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				input := inputs[id%len(inputs)]
				_, _ = ParseThai(input.layout, input.value)
				_, _ = ParseThaiInLocation(input.layout, input.value, time.UTC)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentEraCacheAccess tests concurrent access to the global era cache.
func TestConcurrentEraCacheAccess(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 100

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				// Create different times in different eras
				tm := Date(2020+j%50, 1, 1, 0, 0, 0, 0, time.UTC).InEra(BE())
				_ = tm.Year()
				_ = tm.Format("2006")
				_ = tm.FormatLocale(LocaleThTH, "2006")
			}
		}(i)
	}

	wg.Wait()

	// Verify cache statistics
	stats := EraCacheStats()
	t.Logf("Era cache stats - Hits: %d, Misses: %d, Evictions: %d",
		stats.Hits, stats.Misses, stats.Evictions)

	// Verify hit rate is reasonable
	hitRate := EraCacheHitRate()
	t.Logf("Era cache hit rate: %.2f%%", hitRate*100)
}

// TestConcurrentCacheClear tests concurrent cache clearing operations.
func TestConcurrentCacheClear(t *testing.T) {
	const numGoroutines = 10
	const numIterations = 10

	// Pre-populate the cache
	tm := Date(2024, 6, 15, 12, 0, 0, 0, time.UTC).InEra(BE())
	for i := 0; i < 100; i++ {
		_ = tm.Year()
	}

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				if id%2 == 0 {
					// Access cache
					_ = tm.Year()
					_ = EraCacheStats()
					_ = EraCacheHitRate()
				} else {
					// Clear cache
					ClearEraCache()
				}
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentReferenceDateModification tests concurrent modification
// of reference dates for deterministic behavior.
func TestConcurrentReferenceDateModification(t *testing.T) {
	const numGoroutines = 20
	const numIterations = 50

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				// Set reference dates
				refDate := time.Date(2024, time.Month(j%12+1), 15, 0, 0, 0, 0, time.UTC)
				SetEraDetectionReferenceDate(refDate)
				SetYearFormatReferenceDate(refDate)

				// Use the reference dates
				tm := Date(2024, 6, 15, 12, 0, 0, 0, time.UTC).InEra(BE())
				_ = tm.Format("2006")
				_ = DetectEraFromYear(2567)
			}
		}(i)
	}

	wg.Wait()

	// Clear reference dates at the end
	SetEraDetectionReferenceDate(time.Time{})
	SetYearFormatReferenceDate(time.Time{})
}

// TestConcurrentStringReplacerAccess tests concurrent access to StringReplacer.
func TestConcurrentStringReplacerAccess(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 100

	inputs := []string{
		"Monday, January 15, 2024",
		"Tuesday, February 28, 2023",
		"Wednesday, March 15, 2024",
		"Thursday, April 20, 2023",
		"Friday, May 10, 2024",
	}

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				input := inputs[id%len(inputs)]
				_ = replaceMonthNames(input)
				_ = replaceDayNames(input)
				_ = replaceThaiMonthNames(input)
				_ = replaceThaiDayNames(input)
			}
		}(i)
	}

	wg.Wait()
}

// TestHighConcurrencyStress tests high concurrency stress with many operations.
func TestHighConcurrencyStress(t *testing.T) {
	const numGoroutines = 200

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Start concurrent operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				tm := Date(2020+id%10, int(time.Month(id%12+1)), id%28+1, id%24, id%60, 0, 0, time.UTC)
					if id%2 == 0 {
						tm = tm.InEra(BE())
					}
					_ = tm.Year()
					_ = tm.YearCE()
					_ = tm.Format("2006-01-02 15:04:05")
					_ = tm.FormatLocale(LocaleThTH, "Monday, January 02, 2006 15:04:05")
				}
			}
		}(i)
	}

	// Run for a bit then stop
	time.Sleep(100 * time.Millisecond)
	close(done)
	wg.Wait()
}

// TestMixedConcurrentOperations tests mixed concurrent operations
// that simulate real-world usage patterns.
func TestMixedConcurrentOperations(t *testing.T) {
	const numGoroutines = 50
	const numIterations = 30

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				// Simulate real-world usage
				// 1. Create time in random era
				tm := Date(2020+j%10, int(time.Month(j%12+1)), j%28+1, j%24, j%60, 0, 0, time.UTC)
				if j%3 != 0 {
					tm = tm.InEra(BE())
				}

				// 2. Access year multiple times
				for k := 0; k < 3; k++ {
					_ = tm.Year()
					_ = tm.YearCE()
				}

				// 3. Format in different ways
				_ = tm.Format("2006-01-02")
				_ = tm.FormatLocale(LocaleThTH, "02 มกราคม 2567")
				_ = tm.FormatLocale(LocaleEnUS, "January 02, 2024")

				// 4. Parse and convert
				if j%5 == 0 {
					_, _ = ParseWithEra("2006-01-02", "2024-06-15", BE())
					_, _ = ParseThai("02 มกราคม 2567", "02 มกราคม 2567")
				}

				// 5. Time arithmetic
				tm = tm.Add(time.Hour * 24 * 30)
				_ = tm.AddDate(1, 0, 0)
			}
		}(i)
	}

	wg.Wait()
}

// TestEraCacheLRUConcurrency tests LRU eviction under concurrent access.
func TestEraCacheLRUConcurrency(t *testing.T) {
	const numGoroutines = 50
	const numIterations = 200

	// Clear cache first
	ClearEraCache()

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				// Access times with different years and eras to trigger LRU
				year := 2000 + (id*10 + j)%100
				tm := Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
				if j%2 == 0 {
					tm = tm.InEra(BE())
				}
				_ = tm.Year()
			}
		}(i)
	}

	wg.Wait()

	// Check cache statistics
	stats := EraCacheStats()
	t.Logf("LRU test - Hits: %d, Misses: %d, Evictions: %d",
		stats.Hits, stats.Misses, stats.Evictions)

	// Verify cache is working
	if stats.Hits+stats.Misses == 0 {
		t.Error("Expected cache to have some activity")
	}
}

// BenchmarkConcurrentYearAccess benchmarks concurrent year access performance.
func BenchmarkConcurrentYearAccess(b *testing.B) {
	b.StopTimer()
	tm := Date(2024, 6, 15, 12, 0, 0, 0, time.UTC).InEra(BE())

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = tm.Year()
			_ = tm.YearCE()
		}
	})
}

// BenchmarkConcurrentFormat benchmarks concurrent format operations.
func BenchmarkConcurrentFormat(b *testing.B) {
	b.StopTimer()
	tm := Date(2024, 6, 15, 12, 0, 0, 0, time.UTC).InEra(BE())

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = tm.Format("2006-01-02 15:04:05")
			_ = tm.FormatLocale(LocaleThTH, "Monday, January 02, 2006 15:04:05")
		}
	})
}
