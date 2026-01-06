// Package gotime provides enhanced time handling with support for multiple eras
// (such as Buddhist Era BE and Common Era CE), locale-aware formatting, and
// Thai language text processing. It wraps the standard library's time package
// while adding era-specific functionality commonly used in Thailand and other
// regions that utilize different calendar eras.
//
// The package defines an Era type to represent different calendar systems and
// provides utilities for converting between eras, formatting dates with era-specific
// years, and parsing Thai text representations.
package gotime

import (
	"sync"
	"time"
)

// Era represents a calendar era with a name and year offset from Common Era (CE).
// It is used to handle different calendar systems such as Buddhist Era (BE)
// and Common Era (CE), allowing year conversions and formatting.
type Era struct {
	name   string
	offset int
}

// Era-related constants.
const (
	// BEOffset is the number of years to add to a Common Era year to get
	// the corresponding Buddhist Era year. Buddhist Era is 543 years ahead
	// of the Common Era calendar.
	BEOffset = 543
)

var (
	ce = &Era{name: "CE", offset: 0}
	be = &Era{name: "BE", offset: BEOffset}

	eras   = make(map[string]*Era)
	erasMu sync.RWMutex
)

func init() {
	RegisterEra("CE", 0)
	RegisterEra("BE", BEOffset)
}

// CE returns the Common Era (CE) era instance. Common Era is the
// standard calendar system used internationally, equivalent to AD (Anno Domini).
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

// RegisterEra registers a new era with the given name and offset from Common Era.
// If an era with the same name already exists, it returns the existing era.
// The registration is thread-safe.
func RegisterEra(name string, offset int) *Era {
	erasMu.Lock()
	defer erasMu.Unlock()

	if _, exists := eras[name]; exists {
		return eras[name]
	}

	era := &Era{name: name, offset: offset}
	eras[name] = era
	return era
}

// GetEra retrieves a previously registered era by name.
// Returns nil if the era is not found.
func GetEra(name string) *Era {
	erasMu.RLock()
	defer erasMu.RUnlock()
	return eras[name]
}

// DetectEraFromYear determines which era (CE or BE) the given year is most
// likely to belong to based on proximity to the current year. This is useful
// for Thai date parsing where the era may not be explicitly specified.
func DetectEraFromYear(year int) *Era {
	currentTime := time.Now()
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

// IsValidYear checks if the given year is valid for this era.
// BE era requires positive years (year > 0), while CE era accepts
// zero and positive years.
func (e *Era) IsValidYear(year int) bool {
	if e == BE() {
		return year > 0
	}
	return year >= 0 // CE era accepts year 0 and positive years
}
