package internal

import (
	"strconv"
	"strings"
	"testing"
)

// BuilderPool tests

func TestBuilderPoolBasic(t *testing.T) {
	bp := NewBuilderPool()

	// Get a builder
	b := bp.Get(64)
	if b == nil {
		t.Fatal("Get returned nil")
	}

	// Write some data
	b.WriteString("test")
	if b.String() != "test" {
		t.Errorf("String() = %q, want %q", b.String(), "test")
	}

	// Return to pool
	bp.Put(b)

	// Get again - should be from pool
	b2 := bp.Get(64)
	if b2 == nil {
		t.Fatal("Second Get returned nil")
	}
	if b2.String() != "" {
		t.Errorf("Pooled builder should be empty, got %q", b2.String())
	}
	bp.Put(b2)

	// Check stats
	stats := bp.Stats()
	if stats.Gets < 2 {
		t.Errorf("Expected at least 2 gets, got %d", stats.Gets)
	}
	if stats.Puts < 2 {
		t.Errorf("Expected at least 2 puts, got %d", stats.Puts)
	}
}

func TestBuilderPoolNilPut(t *testing.T) {
	bp := NewBuilderPool()
	// Should not panic
	bp.Put(nil)
}

func TestBuilderPoolLargeBuilder(t *testing.T) {
	bp := NewBuilderPool()

	// Create a builder larger than MaxBuilderCapacity
	b := bp.Get(8192)
	b.WriteString(strings.Repeat("x", 4096))
	bp.Put(b)

	// Large builder should not be pooled
	stats := bp.Stats()
	if stats.Allocates < 1 {
		t.Errorf("Expected at least 1 allocate for large builder, got %d", stats.Allocates)
	}
}

func TestBuilderPoolHitRate(t *testing.T) {
	bp := NewBuilderPool()
	bp.ResetStats()

	// Do many operations to get a stable hit rate
	// sync.Pool may occasionally clear, so we use more iterations
	// and a lower threshold to account for GC behavior
	for i := 0; i < 100; i++ {
		b := bp.Get(64)
		b.WriteString("test")
		bp.Put(b)
	}

	stats := bp.Stats()
	hitRate := stats.HitRate()
	// With 100 iterations, expect >40% hit rate
	// sync.Pool may have some misses due to GC, but most should hit
	if hitRate < 0.4 {
		t.Errorf("Expected hit rate > 0.4, got %f", hitRate)
	}
}

func TestBuilderPoolResetStats(t *testing.T) {
	bp := NewBuilderPool()
	b := bp.Get(64)
	bp.Put(b)

	stats := bp.Stats()
	if stats.Gets == 0 {
		t.Error("Expected some gets before reset")
	}

	bp.ResetStats()
	stats = bp.Stats()
	if stats.Gets != 0 {
		t.Errorf("Expected 0 gets after reset, got %d", stats.Gets)
	}
}

// EraCache tests

func TestEraCacheBasic(t *testing.T) {
	ec := NewEraCache(100)

	// Test cache miss
	if _, ok := ec.Get(2024, nil); ok {
		t.Error("Expected cache miss for empty cache")
	}

	// Test cache miss with era pointer
	// We can't easily create an era pointer here, so we test with nil

	// Add a value
	ec.Set(2024, nil, 2567)

	// Test cache hit with nil
	if year, ok := ec.Get(2024, nil); !ok {
		t.Error("Expected cache hit after Set")
	} else if year != 2567 {
		t.Errorf("Year = %d, want %d", year, 2567)
	}
}

func TestEraCacheStats(t *testing.T) {
	ec := NewEraCache(100)

	// Initial stats
	stats := ec.Stats()
	if stats.Hits != 0 {
		t.Errorf("Expected 0 initial hits, got %d", stats.Hits)
	}

	// Add some entries
	for i := 0; i < 10; i++ {
		ec.Set(2000+i, nil, 2500+i)
	}

	// Access them to generate hits
	for i := 0; i < 10; i++ {
		ec.Get(2000+i, nil)
	}

	stats = ec.Stats()
	if stats.Hits < 5 {
		t.Errorf("Expected at least 5 hits, got %d", stats.Hits)
	}
}

func TestEraCacheHitRate(t *testing.T) {
	ec := NewEraCache(100)

	// Add entries
	for i := 0; i < 10; i++ {
		ec.Set(2000+i, nil, 2500+i)
	}

	// Some hits, some misses - use different keys for misses
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			ec.Get(2000+i, nil) // hits
		} else {
			ec.Get(3000+i, nil) // misses (different keys)
		}
	}

	hitRate := ec.HitRate()
	// Expected: 5 hits out of 10 = 0.5
	if hitRate < 0.4 || hitRate > 0.6 {
		t.Errorf("Expected hit rate ~0.5, got %f", hitRate)
	}
}

func TestEraCacheClear(t *testing.T) {
	ec := NewEraCache(100)

	// Add entries
	ec.Set(2024, nil, 2567)

	// Verify it's there
	if _, ok := ec.Get(2024, nil); !ok {
		t.Error("Expected entry before clear")
	}

	// Clear
	ec.Clear()

	// Should be gone
	if _, ok := ec.Get(2024, nil); ok {
		t.Error("Expected cache miss after clear")
	}
}

func TestEraCacheZeroMaxSize(t *testing.T) {
	// Should use default size
	ec := NewEraCache(0)
	if ec == nil {
		t.Error("Expected non-nil cache with zero max size")
	}
}

// RegexPool tests

func TestRegexPoolBasic(t *testing.T) {
	rp := NewRegexPool(`\d+`)

	// Get a regex
	re := rp.Get()
	if re == nil {
		t.Fatal("Get returned nil")
	}
	rp.Put(re)

	// Use ReplaceAllString
	result := rp.ReplaceAllString("abc123def456", "NUM")
	if result != "abcNUMdefNUM" {
		t.Errorf("ReplaceAllString = %q, want %q", result, "abcNUMdefNUM")
	}
}

func TestRegexPoolFindAllString(t *testing.T) {
	rp := NewRegexPool(`\d+`)

	matches := rp.FindAllString("a1b2c3", -1)
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}
}

func TestRegexPoolFindString(t *testing.T) {
	rp := NewRegexPool(`\d+`)

	match := rp.FindString("abc123def")
	if match != "123" {
		t.Errorf("FindString = %q, want %q", match, "123")
	}
}

func TestRegexPoolMatchString(t *testing.T) {
	rp := NewRegexPool(`\d+`)

	if !rp.MatchString("abc123") {
		t.Error("Expected match for string with digits")
	}
	if rp.MatchString("abc") {
		t.Error("Expected no match for string without digits")
	}
}

func TestRegexPoolReplaceAllStringFunc(t *testing.T) {
	rp := NewRegexPool(`\d+`)

	result := rp.ReplaceAllStringFunc("year 2024", func(match string) string {
		n, _ := strconv.Atoi(match)
		return strconv.Itoa(n + 543)
	})
	if result != "year 2567" {
		t.Errorf("ReplaceAllStringFunc = %q, want %q", result, "year 2567")
	}
}

// StringReplacer tests

func TestStringReplacerBasic(t *testing.T) {
	sr := NewStringReplacer(map[string]string{
		"January":  "มกราคม",
		"February": "กุมภาพันธ์",
	})

	result := sr.Replace("January and February")
	if result != "มกราคม and กุมภาพันธ์" {
		t.Errorf("Replace = %q, want %q", result, "มกราคม and กุมภาพันธ์")
	}
}

func TestStringReplacerNoReplacements(t *testing.T) {
	sr := NewStringReplacer(map[string]string{})

	result := sr.Replace("no replacements")
	if result != "no replacements" {
		t.Errorf("Replace with empty map = %q, want %q", result, "no replacements")
	}
}

func TestStringReplacerLongestMatch(t *testing.T) {
	// "May" full name should be matched before short name
	sr := NewStringReplacer(map[string]string{
		"May":     "พฤษภาคม",
		"January": "มกราคม",
	})

	result := sr.Replace("May and January")
	if result != "พฤษภาคม and มกราคม" {
		t.Errorf("Replace = %q, want %q", result, "พฤษภาคม and มกราคม")
	}
}

func TestStringReplacerMultiple(t *testing.T) {
	sr := NewStringReplacer(map[string]string{
		"January":  "JAN",
		"February": "FEB",
		"March":    "MAR",
	})

	tests := []struct {
		input    string
		expected string
	}{
		{"January", "JAN"},
		{"February", "FEB"},
		{"March", "MAR"},
		{"January February March", "JAN FEB MAR"},
		{"No match", "No match"},
	}

	for _, tt := range tests {
		result := sr.Replace(tt.input)
		if result != tt.expected {
			t.Errorf("Replace(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestStringReplacerEmptyInput(t *testing.T) {
	sr := NewStringReplacer(map[string]string{
		"test": "TEST",
	})

	result := sr.Replace("")
	if result != "" {
		t.Errorf("Replace empty string = %q, want %q", result, "")
	}
}

func TestStringReplacerNoMatch(t *testing.T) {
	sr := NewStringReplacer(map[string]string{
		"test": "TEST",
	})

	result := sr.Replace("no match here")
	if result != "no match here" {
		t.Errorf("Replace with no matches = %q, want %q", result, "no match here")
	}
}

func TestStringReplacerOverlap(t *testing.T) {
	// Test that overlapping patterns work correctly
	sr := NewStringReplacer(map[string]string{
		"ab":  "X",
		"abc": "Y",
	})

	// Should match "abc" first (longer), then no "ab" overlap
	result := sr.Replace("abc")
	if result != "Y" {
		t.Errorf("Overlap test = %q, want %q", result, "Y")
	}
}

func TestStringReplacerReplaceAll(t *testing.T) {
	// ReplaceAll is an alias for Replace, test both paths
	sr := NewStringReplacer(map[string]string{
		"January":  "JAN",
		"February": "FEB",
	})

	// Test ReplaceAll method
	result := sr.ReplaceAll("January and February")
	if result != "JAN and FEB" {
		t.Errorf("ReplaceAll = %q, want %q", result, "JAN and FEB")
	}

	// Verify Replace and ReplaceAll return identical results
	replaceResult := sr.Replace("January and February")
	if result != replaceResult {
		t.Errorf("Replace and ReplaceAll should return same result")
	}

	// Test with empty string
	emptyResult := sr.ReplaceAll("")
	if emptyResult != "" {
		t.Errorf("ReplaceAll empty string = %q, want %q", emptyResult, "")
	}

	// Test with no matches
	noMatchResult := sr.ReplaceAll("No matches here")
	if noMatchResult != "No matches here" {
		t.Errorf("ReplaceAll no match = %q, want %q", noMatchResult, "No matches here")
	}
}

func TestEraCacheLRUEviction(t *testing.T) {
	// Create a small cache to trigger LRU eviction
	ec := NewEraCache(5)

	// Fill the cache beyond capacity
	for i := 0; i < 10; i++ {
		ec.Set(2000+i, nil, 2500+i)
	}

	// Verify cache has entries and evictions occurred
	stats := ec.Stats()
	if stats.Evictions == 0 {
		t.Error("Expected evictions when cache exceeds capacity")
	}
	t.Logf("After filling cache: Evictions=%d", stats.Evictions)

	// The most recent entries should still be accessible
	// Entry 2009 should still be present (it was last in)
	if year, ok := ec.Get(2009, nil); !ok {
		t.Error("Most recent entry (2009) should still be in cache")
	} else if year != 2509 {
		t.Errorf("Entry 2009 year = %d, want 2509", year)
	}

	// Verify cache stats show activity
	finalStats := ec.Stats()
	t.Logf("Final cache stats - Hits: %d, Misses: %d, Evictions: %d",
		finalStats.Hits, finalStats.Misses, finalStats.Evictions)

	// Verify that evictions occurred (this is the main coverage goal)
	if finalStats.Evictions < 5 {
		t.Errorf("Expected at least 5 evictions, got %d", finalStats.Evictions)
	}

	// Test that cache still works after eviction
	ec.Set(3000, nil, 3500)
	if year, ok := ec.Get(3000, nil); !ok {
		t.Error("New entry should be in cache")
	} else if year != 3500 {
		t.Errorf("New entry year = %d, want 3500", year)
	}
}
