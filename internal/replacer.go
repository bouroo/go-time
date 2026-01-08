// Package internal provides utility components for the gotime package.
// This package is not meant to be imported directly by users of gotime.
package internal

import (
	"sort"
)

// builderPool provides a shared pool of strings.Builder instances.
// This reduces allocations when performing string replacements.
var builderPool = NewBuilderPool()

// StringReplacer performs multiple string replacements in a single pass.
// This provides O(n) complexity instead of O(n*m) where n is the input
// length and m is the number of replacement pairs.
//
// Thread Safety: StringReplacer is read-only after initialization,
// making it safe for concurrent access from multiple goroutines.
type StringReplacer struct {
	replacements []replacement
}

// replacement represents a single string replacement pair.
// The 'from' field is the string to find, and 'to' is the replacement.
// 'len' is cached for performance.
type replacement struct {
	from string
	to   string
	len  int
}

// NewStringReplacer creates a new StringReplacer with the given replacement
// map. The map keys are the strings to find, and the values are their
// replacements.
//
// IMPORTANT: Replacements are sorted by length (longest first) to ensure
// that longer patterns are matched before shorter ones that might be
// substrings of them (e.g., "February" before "Feb", "May" full before "May" short).
// For patterns of equal length, they are sorted alphabetically to ensure
// deterministic behavior.
//
// Performance characteristics:
// - Time: O(n) single pass through the input
// - Space: O(n) for the output string
// - Allocations: Single allocation for the result string
func NewStringReplacer(replacements map[string]string) *StringReplacer {
	// Convert map to slice of replacements
	reps := make([]replacement, 0, len(replacements))
	for from, to := range replacements {
		reps = append(reps, replacement{
			from: from,
			to:   to,
			len:  len(from),
		})
	}

	// Sort by length descending (longest first) to avoid partial matches.
	// For equal lengths, sort alphabetically to ensure deterministic behavior.
	// This ensures "February" is matched before "Feb", and full "May" before short "May".
	sort.Slice(reps, func(i, j int) bool {
		if reps[i].len != reps[j].len {
			return reps[i].len > reps[j].len
		}
		// Secondary sort: alphabetically, with a preference for the original
		// English month names (non-abbreviated) to come first
		return reps[i].from > reps[j].from
	})

	return &StringReplacer{
		replacements: reps,
	}
}

// Replace performs all replacements on the input string and returns
// the result. This method is thread-safe and can be called concurrently.
//
// The algorithm iterates through the input string once, at each position
// checking if any replacement matches. The longest matching replacement
// at each position is applied first.
//
// Example:
//
//	sr := internal.NewStringReplacer(map[string]string{
//	    "January": "มกราคม",
//	    "February": "กุมภาพันธ์",
//	})
//	result := sr.Replace("January and February")
//	// result: "มกราคม and กุมภาพันธ์"
func (sr *StringReplacer) Replace(s string) string {
	// Fast path: if no replacements, return input
	if len(sr.replacements) == 0 {
		return s
	}

	// Estimate result size to minimize allocations
	// Start with input length and add room for expansions
	estimatedCap := len(s) + 64
	if estimatedCap < 64 {
		estimatedCap = 64
	}

	// Use pooled builder for reduced allocations
	sb := builderPool.Get(estimatedCap)
	defer builderPool.Put(sb)

	i := 0
	for i < len(s) {
		matched := false

		// Check all replacements at current position
		// Try longest matches first (already sorted by length)
		for _, rep := range sr.replacements {
			if len(s)-i >= rep.len && s[i:i+rep.len] == rep.from {
				sb.WriteString(rep.to)
				i += rep.len
				matched = true
				break
			}
		}

		// No match found, copy current character
		if !matched {
			sb.WriteByte(s[i])
			i++
		}
	}

	return sb.String()
}

// ReplaceAll is an alias for Replace for clarity.
func (sr *StringReplacer) ReplaceAll(s string) string {
	return sr.Replace(s)
}
