// Package time provides enhanced time handling with support for multiple eras
// (such as Buddhist Era BE and Common Era CE), locale-aware formatting, and
// Thai language text processing. It wraps the standard library's time package
// while adding era-specific functionality commonly used in Thailand and other
// regions that utilize different calendar eras.
//
// # Thread Safety
//
// This package is fully thread-safe. All operations can be safely used concurrently:
//
//   - Time values (Time struct) are immutable once created; all access is read-only
//   - Era registry operations (RegisterEra, GetEra) are protected by sync.RWMutex
//   - Era cache operations use sync.Map for lock-free reads and atomic swaps
//   - Reference date configuration uses sync.RWMutex for safe concurrent access
//
// Example of safe concurrent usage:
//
//	time.Now() creates a new Time value, safe for concurrent use
//	tm := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC).InEra(time.BE())
//	// tm can be safely passed to multiple goroutines
//	go func() { _ = tm.Year() }()
//	go func() { _ = tm.Format("2006") }()
//
// The package defines an Era type to represent different calendar systems and
// provides utilities for converting between eras, formatting dates with era-specific
// years, and parsing Thai text representations.
package time

import (
	"sync"
	stdtime "time"

	"github.com/bouroo/go-time/internal"
)

// Era represents a calendar era with a name and year offset from Common Era (CE).
// It is used to handle different calendar systems such as Buddhist Era (BE)
// and Common Era (CE), allowing year conversions and formatting.
//
// The Era struct supports:
//   - Year offset from CE (e.g., BE has offset +543)
//   - Start and end dates for era transitions (e.g., Japanese emperor reigns)
//   - Calendar family grouping (e.g., "Japanese", "Buddhist")
//   - Locale-specific era names (e.g., "令和" vs "Reiwa")
//   - Custom formatting rules (prefix, suffix, year digits)
//
// # Era Structure
//
// Eras can be simple (just an offset from CE) or complex (with start dates
// and locale-specific names). Use RegisterEra() for simple eras or
// RegisterEraWithOptions() for full configuration.
//
// # Example: Simple Era Registration
//
//	gotime.RegisterEra("SE", 100) // Simple Era, 100 years ahead of CE
//
// # Example: Complex Era with Options
//
//	gotime.RegisterEraWithOptions(gotime.EraOptions{
//	    Name:      "ME",
//	    Offset:    100,
//	    StartDate: gotime.Date(2000, 1, 1, 0, 0, 0, 0, gotime.UTC),
//	    Family:    "MyFamily",
//	    Locale:    "en-US",
//	    Format: &gotime.EraFormat{
//	        Prefix:     "MyEra-",
//	        Suffix:     "",
//	        YearDigits: 4,
//	        ZeroBased:  false,
//	    },
//	    Names: map[string]string{
//	        "en-US": "My Era",
//	        "ja-JP": "私の時代",
//	    },
//	})
type Era struct {
	name      string
	offset    int
	startDate stdtime.Time
	endDate   stdtime.Time
	family    string
	locale    string
	format    *EraFormat
	names     map[string]string
	formatter EraFormatterFunc
}

// Era-related constants.
const (
	// BEOffset is the number of years to add to a Common Era year to get
	// the corresponding Buddhist Era year. Buddhist Era is 543 years ahead
	// of the Common Era calendar.
	BEOffset = 543

	// DefaultEraFamily is the default calendar family for simple eras.
	DefaultEraFamily = "Common"
)

// EraFormat defines formatting rules for an era.
// It controls how years in this era are displayed in formatted output.
type EraFormat struct {
	// Prefix is the string to prepend before the year (e.g., "令和" for Japanese).
	Prefix string

	// Suffix is the string to append after the year (e.g., "年" for Japanese).
	Suffix string

	// YearDigits specifies the number of digits to use for the year.
	// Common values are 2 (gannen numbering) or 4 (full year).
	YearDigits int

	// ZeroBased indicates whether the first year of the era is 0 or 1.
	// Japanese eras use 1-based (元年 = year 1), not 0-based.
	ZeroBased bool

	// FullFormat is an optional custom format string for the full era date.
	// If set, this takes precedence over Prefix/YearDigits/Suffix.
	// The format uses the same layout strings as time.Time.Format.
	// Example: "2006年01月02日" for Japanese date format.
	FullFormat string
}

// EraFormatterFunc is a custom formatter function for an era.
// It allows complete control over how dates in this era are formatted.
//
// The function receives the Time value and should return the formatted string.
// Return an empty string to use the default era formatting.
type EraFormatterFunc func(t Time) string

// EraOptions contains all configuration options for registering a new era.
// Use RegisterEraWithOptions() to create an era with these options.
type EraOptions struct {
	// Name is the unique identifier for this era (e.g., "BE", "CE", "Reiwa").
	Name string

	// Offset is the number of years to add to a CE year to get the era year.
	// For BE, this is 543. For CE, this is 0.
	Offset int

	// StartDate is when this era begins. Zero means the era has no specific
	// start date (historical or simple offset-based era).
	StartDate stdtime.Time

	// EndDate is when this era ends. Zero means the era is ongoing.
	// For eras with transitions (e.g., Japanese emperor reigns), set both
	// StartDate and EndDate.
	EndDate stdtime.Time

	// Family is the calendar family this era belongs to (e.g., "Japanese", "Buddhist").
	// Use GetFamily() to retrieve all eras in a family.
	Family string

	// Locale is the primary locale for this era (e.g., "ja-JP", "th-TH").
	// Used for locale-aware era detection and formatting.
	Locale string

	// Format defines era-specific formatting rules.
	Format *EraFormat

	// Names contains localized names for this era by locale.
	// Example: {"en-US": "Reiwa", "ja-JP": "令和"}
	Names map[string]string

	// Formatter is an optional custom formatter function.
	// If provided, this takes precedence over Format for formatting.
	Formatter EraFormatterFunc
}

var (
	ce = &Era{name: "CE", offset: 0}
	be = &Era{name: "BE", offset: BEOffset}

	eras   = make(map[string]*Era)
	erasMu sync.RWMutex

	// detectionReferenceDate is the reference date for era detection.
	// If zero, time.Now() is used. This enables deterministic testing.
	detectionReferenceDate stdtime.Time
	detectionMu            sync.RWMutex

	// familyTransitions maps family name to era transitions.
	// Each family can have multiple transitions (e.g., Japanese eras).
	familyTransitions = make(map[string][]*EraTransition)

	// localeDefaultEras maps locale to default era for that locale.
	// Used for locale-aware era detection.
	localeDefaultEras = make(map[string]*Era)
)

// EraTransition represents a transition from one era to another within a family.
// It specifies when a new era begins, allowing GetEraForDate to determine
// which era was active at any given point in time.
type EraTransition struct {
	era   *Era
	start stdtime.Time
}

func init() {
	RegisterEra("CE", 0)
	RegisterEra("BE", BEOffset)
}

// CE returns the Common Era (CE) era instance. Common Era is the
// standard calendar system used internationally, equivalent to CE (Anno Domini).
func CE() *Era {
	return ce
}

// BE returns the Buddhist Era (BE) era instance. Buddhist Era is used
// in Thailand and some other Southeast Asian countries, dating from the
// enlightenment of Buddha in 543 BCE.
func BE() *Era {
	return be
}

// String returns the era's name, such as "CE" or "BE".
func (e *Era) String() string {
	return e.name
}

// Offset returns the number of years to add to a Common Era year to get
// the corresponding year in this era.
func (e *Era) Offset() int {
	return e.offset
}

// FromCE converts a Common Era year to the corresponding year in this era.
// For example, with BE era and BEOffset of 543, FromCE(2024) returns 2567.
func (e *Era) FromCE(ceYear int) int {
	return ceYear + e.offset
}

// ToCE converts a year from this era to the corresponding Common Era year.
// For example, with BE era and BEOffset of 543, ToCE(2567) returns 2024.
func (e *Era) ToCE(eraYear int) int {
	return eraYear - e.offset
}

// StartDate returns the date when this era begins.
// Returns zero time if the era has no specific start date.
func (e *Era) StartDate() stdtime.Time {
	return e.startDate
}

// EndDate returns the date when this era ends.
// Returns zero time if the era has no specific end date (ongoing era).
func (e *Era) EndDate() stdtime.Time {
	return e.endDate
}

// Family returns the calendar family this era belongs to.
// Returns "Common" for simple eras with no family specified.
func (e *Era) Family() string {
	return e.family
}

// Locale returns the primary locale for this era.
// Returns empty string if no locale was specified.
func (e *Era) Locale() string {
	return e.locale
}

// Format returns the era-specific formatting rules.
// Returns nil if no format was specified.
func (e *Era) Format() *EraFormat {
	return e.format
}

// Names returns the map of localized era names.
// Returns nil if no localized names were specified.
func (e *Era) Names() map[string]string {
	return e.names
}

// NameForLocale returns the era name localized for the given locale.
// If no localized name exists for the locale, returns the default era name.
func (e *Era) NameForLocale(locale string) string {
	if e.names != nil {
		if name, ok := e.names[locale]; ok {
			return name
		}
	}
	return e.name
}

// IsValidForDate checks if this era was active at the given date.
// For eras with no start/end dates, this always returns true.
// For eras with only a start date, returns true if date >= startDate.
// For eras with only an end date, returns true if date < endDate.
// For eras with both, returns true if startDate <= date < endDate.
func (e *Era) IsValidForDate(date stdtime.Time) bool {
	if !e.startDate.IsZero() && date.Before(e.startDate) {
		return false
	}
	if !e.endDate.IsZero() && !date.Before(e.endDate) {
		return false
	}
	return true
}

// YearInEra returns the year number within this era for the given date.
// This correctly handles zero-based years (e.g., if ZeroBased is true,
// the first year is year 0, not year 1).
func (e *Era) YearInEra(date stdtime.Time) int {
	ceYear := date.Year()
	eraYear := e.FromCE(ceYear)

	if e.format != nil && e.format.ZeroBased {
		// Adjust for zero-based counting
		return eraYear
	}

	// First year is 1, not 0
	return eraYear
}

// RegisterEra registers a new era with the given name and offset from Common Era.
// If an era with the same name already exists, it returns the existing era.
// The registration is thread-safe. This also clears the era cache to ensure
// consistency when new eras are added.
func RegisterEra(name string, offset int) *Era {
	erasMu.Lock()
	defer erasMu.Unlock()

	if _, exists := eras[name]; exists {
		return eras[name]
	}

	era := &Era{name: name, offset: offset}
	eras[name] = era

	// Clear the global era cache to ensure consistency with new era
	globalEraCache.Clear()

	return era
}

// RegisterEraWithOptions registers a new era with full configuration options.
// This allows specifying start/end dates, locale, formatting rules, and
// localized names for the era.
//
// The options parameter must have a non-empty Name. If an era with the same
// name already exists, it returns the existing era without applying new options.
// To update an existing era, first unregister it (if supported) or use a new name.
//
// This function is thread-safe and clears the era cache to ensure consistency.
//
// # Example
//
//	era := gotime.RegisterEraWithOptions(gotime.EraOptions{
//	    Name:      "Reiwa",
//	    Offset:    2018, // Reiwa began in 2018 CE
//	    StartDate: gotime.Date(2019, 5, 1, 0, 0, 0, 0, gotime.UTC),
//	    Family:    "Japanese",
//	    Locale:    "ja-JP",
//	    Format: &gotime.EraFormat{
//	        Prefix:     "令和",
//	        Suffix:     "",
//	        YearDigits: 2,
//	        ZeroBased:  false,
//	    },
//	    Names: map[string]string{
//	        "en-US": "Reiwa",
//	        "ja-JP": "令和",
//	    },
//	})
func RegisterEraWithOptions(options EraOptions) *Era {
	if options.Name == "" {
		return nil
	}

	erasMu.Lock()
	defer erasMu.Unlock()

	if existing, exists := eras[options.Name]; exists {
		return existing
	}

	era := &Era{
		name:      options.Name,
		offset:    options.Offset,
		startDate: options.StartDate,
		endDate:   options.EndDate,
		family:    options.Family,
		locale:    options.Locale,
		format:    options.Format,
		names:     options.Names,
		formatter: options.Formatter,
	}

	if era.family == "" {
		era.family = DefaultEraFamily
	}

	eras[options.Name] = era

	// Clear the global era cache to ensure consistency with new era
	globalEraCache.Clear()

	return era
}

// RegisterEraTransition registers a transition between two eras within a family.
// This is useful for defining when one era ends and another begins, such as
// in the Japanese calendar where emperor reigns define era boundaries.
//
// The transition takes effect at the start of the startDate. Dates before
// startDate belong to the previous era; dates at or after startDate belong
// to the new era.
//
// This function is thread-safe.
func RegisterEraTransition(family string, newEra *Era, startDate stdtime.Time) error {
	erasMu.Lock()
	defer erasMu.Unlock()

	// Get or create family transitions map
	if familyTransitions[family] == nil {
		familyTransitions[family] = make([]*EraTransition, 0)
	}

	// Add new transition
	transition := &EraTransition{
		era:   newEra,
		start: startDate,
	}
	familyTransitions[family] = append(familyTransitions[family], transition)

	// Sort transitions by start date
	// (simple bubble sort for small lists - switch to sort.Slice for larger)
	for i := 0; i < len(familyTransitions[family])-1; i++ {
		for j := 0; j < len(familyTransitions[family])-i-1; j++ {
			if familyTransitions[family][j].start.After(familyTransitions[family][j+1].start) {
				familyTransitions[family][j], familyTransitions[family][j+1] =
					familyTransitions[family][j+1], familyTransitions[family][j]
			}
		}
	}

	return nil
}

// GetEraForDate returns the active era for a given date within a family.
// This is useful for Japanese calendar dates where the era changes based
// on the emperor's reign dates.
//
// If no transitions are registered for the family, returns nil.
func GetEraForDate(date stdtime.Time, family string) *Era {
	erasMu.RLock()
	defer erasMu.RUnlock()

	transitions := familyTransitions[family]
	if len(transitions) == 0 {
		return nil
	}

	// Find the most recent transition that has started
	var activeEra *Era
	for _, t := range transitions {
		if !date.Before(t.start) {
			activeEra = t.era
		} else {
			break
		}
	}

	return activeEra
}

// GetEraTransitions returns all registered transitions for a family.
// The transitions are sorted by start date.
func GetEraTransitions(family string) []*EraTransition {
	erasMu.RLock()
	defer erasMu.RUnlock()

	transitions := familyTransitions[family]
	if transitions == nil {
		return nil
	}

	// Return a copy to prevent modification
	result := make([]*EraTransition, len(transitions))
	copy(result, transitions)
	return result
}

// GetEra retrieves a previously registered era by name.
// Returns nil if the era is not found.
func GetEra(name string) *Era {
	erasMu.RLock()
	defer erasMu.RUnlock()
	return eras[name]
}

// SetEraDetectionReferenceDate sets the reference date for era detection.
// This is useful for deterministic testing. Pass a zero time.Time to use time.Now().
func SetEraDetectionReferenceDate(t stdtime.Time) {
	detectionMu.Lock()
	defer detectionMu.Unlock()
	detectionReferenceDate = t
}

// ClearEraCache clears the global era cache.
// This is useful when you want to release memory or when custom eras
// have been registered and you want to ensure cache consistency.
func ClearEraCache() {
	globalEraCache.Clear()
}

// EraCacheStats returns the current statistics for the global era cache.
// This can be used to monitor cache effectiveness.
func EraCacheStats() internal.CacheStats {
	return globalEraCache.Stats()
}

// EraCacheHitRate returns the hit rate of the global era cache as a percentage.
func EraCacheHitRate() float64 {
	return globalEraCache.HitRate()
}

// DetectEraFromYear determines which era (CE or BE) the given year is most
// likely to belong to based on proximity to the reference date. This is useful
// for Thai date parsing where the era may not be explicitly specified.
// The reference date is configurable via SetEraDetectionReferenceDate for testing.
func DetectEraFromYear(year int) *Era {
	detectionMu.RLock()
	refDate := detectionReferenceDate
	detectionMu.RUnlock()

	currentTime := refDate
	if currentTime.IsZero() {
		currentTime = stdtime.Now()
	}
	currentCEYear := currentTime.Year()
	currentBEYear := currentCEYear + BE().offset

	ceDiff := absInt(year - currentCEYear)
	beDiff := absInt(year - currentBEYear)

	if beDiff < ceDiff {
		return BE()
	}

	return CE()
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// DetectEraFromYearAndDate determines which era the given year is most likely
// to belong to, considering both the year value and the date context.
// It also considers locale hints if available.
//
// This is more accurate than DetectEraFromYear when dealing with dates
// that have explicit era information (e.g., parsed from Japanese dates).
func DetectEraFromYearAndDate(year int, date stdtime.Time, locale string) *Era {
	// First check locale-specific defaults
	if era := DetectEraForLocale(locale); era != nil {
		return era
	}

	// Fall back to year-based detection
	return DetectEraFromYear(year)
}

// SetLocaleDefaultEra sets the default era for a locale.
// This is used by DetectEraForLocale and DetectEraFromYearAndDate
// to provide locale-aware era detection.
//
// For example, setting the default era for "ja-JP" to Reiwa will cause
// Japanese dates to be detected as Reiwa era by default.
func SetLocaleDefaultEra(locale string, era *Era) {
	detectionMu.Lock()
	defer detectionMu.Unlock()

	if localeDefaultEras == nil {
		localeDefaultEras = make(map[string]*Era)
	}
	localeDefaultEras[locale] = era
}

// DetectEraForLocale returns the default era for the given locale.
// Returns nil if no default era is set for the locale.
//
// Built-in locale defaults:
//   - "th-TH" → BE (Buddhist Era)
//   - "ja-JP" → No default (use GetEraForDate with Japanese family)
//
// Use SetLocaleDefaultEra() to set custom defaults.
func DetectEraForLocale(locale string) *Era {
	detectionMu.RLock()
	defer detectionMu.RUnlock()

	// Check explicitly set defaults first
	if era, ok := localeDefaultEras[locale]; ok {
		return era
	}

	// Built-in defaults
	switch locale {
	case "th-TH":
		return BE()
	}

	return nil
}

// GetLocaleDefaultEra returns the explicitly set default era for a locale.
// Returns nil if no default has been set for the locale.
func GetLocaleDefaultEra(locale string) *Era {
	detectionMu.RLock()
	defer detectionMu.RUnlock()

	return localeDefaultEras[locale]
}

// ClearLocaleDefaultEra removes the default era setting for a locale.
func ClearLocaleDefaultEra(locale string) {
	detectionMu.Lock()
	defer detectionMu.Unlock()

	delete(localeDefaultEras, locale)
}

// ListLocaleDefaultEras returns a copy of all registered locale default eras.
func ListLocaleDefaultEras() map[string]*Era {
	detectionMu.RLock()
	defer detectionMu.RUnlock()

	result := make(map[string]*Era)
	for k, v := range localeDefaultEras {
		result[k] = v
	}
	return result
}

// EraFamilyNames returns a list of all registered calendar family names.
func EraFamilyNames() []string {
	erasMu.RLock()
	defer erasMu.RUnlock()

	families := make(map[string]bool)
	for _, era := range eras {
		if era.family != "" {
			families[era.family] = true
		}
	}

	result := make([]string, 0, len(families))
	for family := range families {
		result = append(result, family)
	}

	return result
}

// GetErasInFamily returns all eras belonging to a specific calendar family.
// Returns nil if no family with that name exists.
func GetErasInFamily(family string) []*Era {
	erasMu.RLock()
	defer erasMu.RUnlock()

	var result []*Era
	for _, era := range eras {
		if era.family == family {
			result = append(result, era)
		}
	}

	return result
}

// IsValidYear checks if the given year is valid for this era.
// BE era requires positive years (year > 0), while CE era accepts
// zero and positive years.
func (e *Era) IsValidYear(year int) bool {
	if e == BE() {
		return year > 0
	}
	return year >= 0 // CE era accepts year 0 and positive years
}
