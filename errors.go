// Package time provides error types for time parsing and formatting operations.
// These errors provide detailed context about parsing failures, including
// the input value, expected layout, era, and the underlying error.
package time

import (
	"errors"
	"fmt"
)

// ErrorCode represents a category of errors for programmatic handling.
type ErrorCode string

const (
	// ErrCodeInvalidFormat indicates the format string is invalid.
	ErrCodeInvalidFormat ErrorCode = "invalid_format"
	// ErrCodeInvalidTime indicates the time value is invalid.
	ErrCodeInvalidTime ErrorCode = "invalid_time"
	// ErrCodeInvalidEra indicates an invalid era was specified.
	ErrCodeInvalidEra ErrorCode = "invalid_era"
	// ErrCodeEraMismatch indicates an era/time mismatch.
	ErrCodeEraMismatch ErrorCode = "era_mismatch"
	// ErrCodeThaiText indicates a Thai text processing error.
	ErrCodeThaiText ErrorCode = "thai_text_error"
	// ErrCodeOutOfBounds indicates a value is out of valid bounds.
	ErrCodeOutOfBounds ErrorCode = "out_of_bounds"
	// ErrCodeUnknown indicates an unknown error.
	ErrCodeUnknown ErrorCode = "unknown"
)

// baseError provides common error functionality.
type baseError struct {
	code     ErrorCode
	message  string
	original error
	context  map[string]any
}

// Error returns a human-readable description of the error.
func (e *baseError) Error() string {
	if e.original != nil {
		return fmt.Sprintf("%s: %v", e.message, e.original)
	}
	return e.message
}

// Unwrap returns the underlying error.
func (e *baseError) Unwrap() error {
	return e.original
}

// Code returns the error code.
func (e *baseError) Code() ErrorCode {
	return e.code
}

// Context returns additional context information about the error.
func (e *baseError) Context() map[string]any {
	return e.context
}

// TimeError is the common interface for all time errors.
type TimeError interface {
	error
	Unwrap() error
	Code() ErrorCode
	Context() map[string]any
}

// ParseError represents an error that occurred while parsing a time value.
// It contains the input string that failed to parse, the expected layout,
// the era context, and the original underlying error.
type ParseError struct {
	baseError
	Input    string
	Layout   string
	Era      *Era
	Original error // Kept for backward compatibility
	Position int   // Line number where the error occurred (1-based)
}

// newParseError creates a new ParseError with the specified parameters.
func newParseError(input, layout string, era *Era, pos int, original error) *ParseError {
	eraStr := "CE"
	if era != nil {
		eraStr = era.String()
	}

	return &ParseError{
		baseError: baseError{
			code:     ErrCodeInvalidFormat,
			message:  "failed to parse time",
			original: original,
			context: map[string]any{
				"input":    input,
				"layout":   layout,
				"era":      eraStr,
				"position": pos,
			},
		},
		Input:    input,
		Layout:   layout,
		Era:      era,
		Original: original, // For backward compatibility
		Position: pos,
	}
}

// Line returns the 1-based line number where the error occurred.
// Returns 0 if position information is not available.
func (e *ParseError) Line() int {
	return e.Position
}

// Column returns the column position where the error occurred.
// Currently returns the same value as Line() for compatibility.
func (e *ParseError) Column() int {
	return e.Position
}

// Error returns a human-readable description of the parse error,
// including the input, layout, era, and original error message.
func (e *ParseError) Error() string {
	eraStr := "CE"
	if e.Era != nil {
		eraStr = e.Era.String()
	}
	return fmt.Sprintf("cannot parse %q as %q with era %s: %v", e.Input, e.Layout, eraStr, e.original)
}

// ThaiTextError represents an error related to Thai text processing,
// such as invalid Thai month or day names.
type ThaiTextError struct {
	baseError
	Input  string
	Reason string
}

// newThaiTextError creates a new ThaiTextError with the specified parameters.
//
//nolint:unused
func newThaiTextError(input, reason string, original error) *ThaiTextError {
	return &ThaiTextError{
		baseError: baseError{
			code:     ErrCodeThaiText,
			message:  "invalid Thai text",
			original: original,
			context: map[string]any{
				"input":  input,
				"reason": reason,
			},
		},
		Input:  input,
		Reason: reason,
	}
}

// Error returns a human-readable description of the Thai text error.
func (e *ThaiTextError) Error() string {
	return fmt.Sprintf("invalid Thai text %q: %s", e.Input, e.Reason)
}

// ValidationError represents a validation error for era-related operations.
type ValidationError struct {
	baseError
	Field      string
	Value      any
	Constraint string
}

// newValidationError creates a new ValidationError with the specified parameters.
//
//nolint:unused
func newValidationError(field string, value any, constraint string) *ValidationError {
	return &ValidationError{
		baseError: baseError{
			code:     ErrCodeInvalidEra,
			message:  fmt.Sprintf("validation failed for %s", field),
			original: nil,
			context: map[string]any{
				"field":      field,
				"value":      value,
				"constraint": constraint,
			},
		},
		Field:      field,
		Value:      value,
		Constraint: constraint,
	}
}

// Error returns a human-readable description of the validation error.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s (value=%v)", e.Field, e.Constraint, e.Value)
}

// TimeValidationError represents an error for invalid time values.
type TimeValidationError struct {
	baseError
	Field    string
	Value    any
	MinValue any
	MaxValue any
}

// newTimeValidationError creates a new TimeValidationError with the specified parameters.
//
//nolint:unused
func newTimeValidationError(field string, value, min, max any) *TimeValidationError {
	return &TimeValidationError{
		baseError: baseError{
			code:     ErrCodeOutOfBounds,
			message:  fmt.Sprintf("time value out of bounds for %s", field),
			original: nil,
			context: map[string]any{
				"field": field,
				"value": value,
				"min":   min,
				"max":   max,
			},
		},
		Field:    field,
		Value:    value,
		MinValue: min,
		MaxValue: max,
	}
}

// Error returns a human-readable description of the time validation error.
func (e *TimeValidationError) Error() string {
	return fmt.Sprintf("time value out of bounds for %s: %v (valid range: %v to %v)", e.Field, e.Value, e.MinValue, e.MaxValue)
}

// EraMismatchError represents an error when an era/time mismatch is detected.
type EraMismatchError struct {
	baseError
	ExpectedEra *Era
	ActualEra   *Era
	Details     string
}

// newEraMismatchError creates a new EraMismatchError with the specified parameters.
//
//nolint:unused
func newEraMismatchError(expectedEra, actualEra *Era, details string) *EraMismatchError {
	expectedStr := "CE"
	if expectedEra != nil {
		expectedStr = expectedEra.String()
	}
	actualStr := "CE"
	if actualEra != nil {
		actualStr = actualEra.String()
	}

	return &EraMismatchError{
		baseError: baseError{
			code:     ErrCodeEraMismatch,
			message:  "era mismatch",
			original: nil,
			context: map[string]any{
				"expected_era": expectedStr,
				"actual_era":   actualStr,
				"details":      details,
			},
		},
		ExpectedEra: expectedEra,
		ActualEra:   actualEra,
		Details:     details,
	}
}

// Error returns a human-readable description of the era mismatch error.
func (e *EraMismatchError) Error() string {
	return fmt.Sprintf("era mismatch: expected %s, got %s: %s",
		e.getEraName(e.ExpectedEra), e.getEraName(e.ActualEra), e.Details)
}

func (e *EraMismatchError) getEraName(era *Era) string {
	if era == nil {
		return "CE"
	}
	return era.String()
}

// MultiError aggregates multiple errors for batch operations.
type MultiError struct {
	errors []error
}

// NewMultiError creates a new MultiError.
func NewMultiError() *MultiError {
	return &MultiError{}
}

// Add adds an error to the collection if it's not nil.
func (e *MultiError) Add(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}

// AddAll adds multiple errors to the collection.
func (e *MultiError) AddAll(errs ...error) {
	for _, err := range errs {
		e.Add(err)
	}
}

// Errors returns the list of errors.
func (e *MultiError) Errors() []error {
	return e.errors
}

// Error returns a string representation of all errors.
// If there are no errors, it returns an empty string.
// If there is one error, it returns that error's string.
// If there are multiple errors, it returns a summary.
func (e *MultiError) Error() string {
	if len(e.errors) == 0 {
		return ""
	}
	if len(e.errors) == 1 {
		return e.errors[0].Error()
	}
	return fmt.Sprintf("%d errors occurred: %v", len(e.errors), e.errors[0])
}

// HasErrors returns true if there are any errors in the collection.
func (e *MultiError) HasErrors() bool {
	return len(e.errors) > 0
}

// Count returns the number of errors in the collection.
func (e *MultiError) Count() int {
	return len(e.errors)
}

// Range calls f for each error in the collection.
func (e *MultiError) Range(f func(index int, err error)) {
	for i, err := range e.errors {
		f(i, err)
	}
}

// Is reports whether any error in the collection matches target.
func (e *MultiError) Is(target error) bool {
	for _, err := range e.errors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// As reports whether any error in the collection can be assigned to target.
func (e *MultiError) As(target any) bool {
	for _, err := range e.errors {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}

// --- Error Helper Functions ---

// IsParseError reports whether err is a ParseError.
func IsParseError(err error) bool {
	var pe *ParseError
	return errors.As(err, &pe)
}

// IsThaiTextError reports whether err is a ThaiTextError.
func IsThaiTextError(err error) bool {
	var te *ThaiTextError
	return errors.As(err, &te)
}

// IsValidationError reports whether err is a ValidationError.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

// IsTimeValidationError reports whether err is a TimeValidationError.
func IsTimeValidationError(err error) bool {
	var tve *TimeValidationError
	return errors.As(err, &tve)
}

// IsEraMismatchError reports whether err is an EraMismatchError.
func IsEraMismatchError(err error) bool {
	var eme *EraMismatchError
	return errors.As(err, &eme)
}

// IsMultiError reports whether err is a MultiError.
func IsMultiError(err error) bool {
	var me *MultiError
	return errors.As(err, &me)
}

// GetErrorCode returns the error code for the given error.
// Returns ErrCodeUnknown if the error doesn't have a code.
func GetErrorCode(err error) ErrorCode {
	var ge TimeError
	if errors.As(err, &ge) {
		return ge.Code()
	}
	return ErrCodeUnknown
}

// GetErrorPosition returns the line and column where the error occurred.
// For ParseError, returns the position information.
// For other errors, returns (0, 0).
func GetErrorPosition(err error) (line, column int) {
	var pe *ParseError
	if errors.As(err, &pe) {
		return pe.Line(), pe.Column()
	}
	return 0, 0
}

// GetErrorContext returns the context map for the given error.
// Returns nil if the error doesn't have context.
func GetErrorContext(err error) map[string]any {
	var ge TimeError
	if errors.As(err, &ge) {
		return ge.Context()
	}
	return nil
}

// GetParseInput returns the input string that caused a ParseError.
// Returns empty string if err is not a ParseError.
func GetParseInput(err error) string {
	var pe *ParseError
	if errors.As(err, &pe) {
		return pe.Input
	}
	return ""
}

// GetParseLayout returns the layout string used in a ParseError.
// Returns empty string if err is not a ParseError.
func GetParseLayout(err error) string {
	var pe *ParseError
	if errors.As(err, &pe) {
		return pe.Layout
	}
	return ""
}

// UnwrapErrors unwraps an error and returns all errors in the chain.
// If the error is a MultiError, returns all errors from it.
// Otherwise, returns a slice containing the single error.
func UnwrapErrors(err error) []error {
	if me, ok := err.(*MultiError); ok {
		return me.Errors()
	}
	return []error{err}
}
