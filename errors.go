// Package gotime provides error types for time parsing and formatting operations.
// These errors provide detailed context about parsing failures, including
// the input value, expected layout, era, and the underlying error.
package gotime

import "fmt"

// ParseError represents an error that occurred while parsing a time value.
// It contains the input string that failed to parse, the expected layout,
// the era context, and the original underlying error.
type ParseError struct {
	Input    string
	Layout   string
	Era      *Era
	Original error
}

// Error returns a human-readable description of the parse error,
// including the input, layout, era, and original error message.
func (e *ParseError) Error() string {
	eraStr := "AD"
	if e.Era != nil {
		eraStr = e.Era.String()
	}
	return fmt.Sprintf("cannot parse %q as %q with era %s: %v", e.Input, e.Layout, eraStr, e.Original)
}

// Unwrap returns the underlying error that caused the parse failure.
func (e *ParseError) Unwrap() error {
	return e.Original
}

// ThaiTextError represents an error related to Thai text processing,
// such as invalid Thai month or day names.
type ThaiTextError struct {
	Input    string
	Reason   string
	Original error
}

// Error returns a human-readable description of the Thai text error.
func (e *ThaiTextError) Error() string {
	return fmt.Sprintf("invalid Thai text %q: %s", e.Input, e.Reason)
}

// Unwrap returns the underlying error that caused the Thai text processing failure.
func (e *ThaiTextError) Unwrap() error {
	return e.Original
}
