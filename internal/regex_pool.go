// Package internal provides utility components for the time package.
// This package is not meant to be imported directly by users of time.
package internal

import (
	"regexp"
	"sync"
)

// RegexPool provides a thread-safe pool of compiled regex objects.
// This eliminates the overhead of runtime regex compilation by reusing
// pre-compiled regex instances across multiple goroutines.
//
// Performance characteristics:
// - Get/Put operations: O(1) amortized
// - Regex operations: Same as stdlib regexp
// - Memory: Reduced allocations through object reuse
//
// Thread Safety: RegexPool uses sync.Pool internally, making it safe
// for concurrent access from multiple goroutines.
type RegexPool struct {
	// pattern holds the compiled regex pattern for creating new instances
	pattern *regexp.Regexp
	// pool provides thread-safe object pooling
	pool *sync.Pool
}

// NewRegexPool creates a new RegexPool with the given pattern.
// The pattern is pre-compiled once at initialization time.
// Returns an error if the pattern is invalid.
func NewRegexPool(pattern string) *RegexPool {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		// This is a programming error, panic during init
		panic("invalid regex pattern: " + pattern)
	}

	return &RegexPool{
		pattern: compiled,
		pool: &sync.Pool{
			New: func() any {
				// Create a fresh copy of the compiled regex for each pool entry.
				// Copy() is deprecated since Go 1.12 but still required here for thread safety
				// as each pool entry may be used concurrently by different goroutines.
				return compiled.Copy() //nolint:staticcheck
			},
		},
	}
}

// Get retrieves a regex instance from the pool.
// If the pool is empty, a new compiled copy is created.
// The caller must call Put() to return the instance to the pool when done.
func (rp *RegexPool) Get() *regexp.Regexp {
	return rp.pool.Get().(*regexp.Regexp)
}

// Put returns a regex instance to the pool for reuse.
// This must be called after Get() to avoid memory leaks.
// The regex instance must not be used after being put back.
func (rp *RegexPool) Put(re *regexp.Regexp) {
	rp.pool.Put(re)
}

// ReplaceAllStringFunc executes the given function on all matches
// in the input string and returns the modified string.
// This is a convenience method that handles Get/Put automatically.
//
// Example:
//
//	result := pool.ReplaceAllStringFunc("Year 2567", func(match string) string {
//	    year, _ := strconv.Atoi(match)
//	    return fmt.Sprintf("%d", year-543)
//	})
func (rp *RegexPool) ReplaceAllStringFunc(s string, fn func(string) string) string {
	re := rp.Get()
	defer rp.Put(re)
	return re.ReplaceAllStringFunc(s, fn)
}

// ReplaceAllString replaces all matches of the pattern in s with repl.
// This is a convenience method that handles Get/Put automatically.
func (rp *RegexPool) ReplaceAllString(s, repl string) string {
	re := rp.Get()
	defer rp.Put(re)
	return re.ReplaceAllString(s, repl)
}

// FindAllString finds all substrings in s that match the pattern.
// The n argument specifies the maximum number of matches to return:
// -1 means all matches.
// This is a convenience method that handles Get/Put automatically.
func (rp *RegexPool) FindAllString(s string, n int) []string {
	re := rp.Get()
	defer rp.Put(re)
	return re.FindAllString(s, n)
}

// FindString finds the first match of the pattern in s.
// Returns empty string if no match is found.
func (rp *RegexPool) FindString(s string) string {
	re := rp.Get()
	defer rp.Put(re)
	return re.FindString(s)
}

// MatchString reports whether the pattern matches s.
func (rp *RegexPool) MatchString(s string) bool {
	re := rp.Get()
	defer rp.Put(re)
	return re.MatchString(s)
}
