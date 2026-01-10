# go-time

A Go library for Thai/Buddhist Era (BE) date support with seamless `time.Time` compatibility.

[![Go Reference](https://pkg.go.dev/badge/github.com/bouroo/go-time.svg)](https://pkg.go.dev/github.com/bouroo/go-time)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## What is go-time?

go-time provides comprehensive Thai calendar support for Go applications. It enables developers to work with Buddhist Era dates naturally while maintaining full compatibility with the standard library's `time.Time` type.

### Key Capabilities

- **Era Conversion**: Convert between Anno Domini (CE) and Buddhist Era (BE) effortlessly
- **Thai Localization**: Full Thai month and day name translations
- **Flexible Parsing**: Parse dates written in Thai format or auto-detect the era
- **Drop-in Replacement**: Use familiar `time.Time` methods without learning a new API
- **Minimal Overhead**: Performance tuned to match standard library operations

## Installation

```bash
go get github.com/bouroo/go-time
```

## Quick Examples

### Converting Between Eras

```go
import (
    "fmt"
    "time"
    "github.com/bouroo/go-time"
)

// Create a standard time
t := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)

// Convert to Buddhist Era
beTime := t.InEra(time.BE())
fmt.Printf("BE Year: %d\n", beTime.Year())    // 2567
fmt.Printf("CE Year: %d\n", beTime.YearCE())  // 2024
```

### Thai Date Formatting

```go
beTime := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC).InEra(time.BE())

// Format with Thai month names
thai := beTime.FormatLocale(time.LocaleThTH, "02 January 2006")
// Output: "2 มกราคม 2549"

// Format day of week in Thai
dayName := beTime.FormatLocale(time.LocaleThTH, "Monday")
// Output: "จันทร์"
```

### Parsing Thai Dates

```go
// Parse with explicit era
t, err := time.ParseWithEra("02 January 2006", "15 มกราคม 2567", time.BE())
if err != nil {
    log.Fatal(err)
}
fmt.Printf("CE Year: %d\n", t.YearCE())  // 2024

// Auto-detect era from year value
t, err := time.ParseThai("02/01/2006", "15/01/2567")
// Years 2501-2599 detected as BE automatically
fmt.Printf("Era: %v\n", t.Era())        // BE
```

## Why go-time?

### Use Cases

- **Financial Applications**: Generate Thai tax invoices and receipts with correct BE dates
- **Government Systems**: Handle official documents requiring Buddhist Era dating
- **Localization Projects**: Display dates in Thai format for Thai-speaking users
- **Date Validation**: Parse and validate Thai-entered dates in forms

### Design Principles

1. **Compatibility First**: Every function is designed to work with `time.Time` seamlessly
2. **Zero Surprises**: Behavior matches the standard library wherever possible
3. **Type Safety**: Strong typing prevents era confusion at compile time
4. **Performance Conscious**: Benchmarks included; Thai operations add minimal overhead

## Supported Formats

The library recognizes all Thai calendar components:

### Thai Month Names

| English    | Thai Full    | Thai Short |
|------------|--------------|------------|
| January    | มกราคม       | ม.ค.       |
| February   | กุมภาพันธ์    | ก.พ.       |
| March      | มีนาคม       | มี.ค.      |
| April      | เมษายน       | เม.ย.      |
| May        | พฤษภาคม      | พ.ค.       |
| June       | มิถุนายน      | มิ.ย.      |
| July       | กรกฎาคม      | ก.ค.       |
| August     | สิงหาคม      | ส.ค.       |
| September  | กันยายน      | ก.ย.       |
| October    | ตุลาคม       | ต.ค.       |
| November   | พฤศจิกายน     | พ.ย.       |
| December   | ธันวาคม      | ธ.ค.       |

### Thai Day Names

| English    | Thai Full    | Thai Short |
|------------|--------------|------------|
| Monday     | จันทร์      | จ.         |
| Tuesday    | อังคาร     | อ.         |
| Wednesday  | พุธ        | พ.         |
| Thursday   | พฤหัสบดี    | พฤ.        |
| Friday     | ศุกร์       | ศ.         |
| Saturday   | เสาร์       | ส.         |
| Sunday     | อาทิตย์     | อา.        |

## API Reference

### Core Types

| Type     | Description                                    |
|----------|------------------------------------------------|
| `Time`   | Wraps `time.Time` with era support             |
| `Era`    | Represents calendar eras (CE, BE)              |
| `Locale` | Locale identifiers for formatting              |

### Parsing Functions

| Function                                   | Returns      | Description                          |
|--------------------------------------------|--------------|--------------------------------------|
| `Parse(layout, value string)`              | `time.Time`  | Drop-in for `time.Parse()`           |
| `ParseInLocation(layout, value, loc)`      | `time.Time`  | Drop-in for `time.ParseInLocation()` |
| `ParseWithEra(layout, value string, era)`  | `Time`       | Parse with explicit era              |
| `ParseInLocationWithEra(layout, value, loc, era)` | `Time` | Parse in location with era           |
| `ParseThai(layout, value string)`          | `Time`       | Auto-detect era from year            |
| `ParseThaiInLocation(layout, value, loc)`  | `Time`       | Auto-detect era in location          |

### Time Methods

| Method                    | Returns   | Description                     |
|---------------------------|-----------|---------------------------------|
| `Year() int`              | `int`     | Year in current era             |
| `YearCE() int`            | `int`     | Year in CE regardless of era    |
| `InEra(e *Era) Time`      | `Time`    | Convert to different era        |
| `Format(layout string)`   | `string`  | Format using Go layout          |
| `FormatLocale(locale, layout)` | `string` | Format with locale translations |

### Era Functions

| Function                       | Returns   | Description                          |
|--------------------------------|-----------|--------------------------------------|
| `CE() *Era`                    | `*Era`    | Returns CE era instance              |
| `BE() *Era`                    | `*Era`    | Returns BE era instance              |
| `DetectEraFromYear(year int)`  | `*Era`    | Detect era from year value (2501-2599 → BE) |

### Error Handling

```go
// ParseError provides detailed error context
type ParseError struct {
    // Contains layout, value, and offset information
}

// Example error handling
t, err := time.ParseWithEra("2006-01-02", "invalid-date", time.BE())
if err != nil {
    fmt.Printf("Error at offset %d: %v\n", err.Offset, err.Err)
}
```

## Performance

Benchmarks run on Go 1.25.5, darwin/arm64:

```
BenchmarkDate                      188,699,502    6.146 ns/op    0 B/op    0 allocs
BenchmarkNow                        43,979,469   26.24 ns/op    0 B/op    0 allocs
BenchmarkInEraCE                   742,275,505    1.590 ns/op    0 B/op    0 allocs
BenchmarkInEraBE                   757,681,784    1.625 ns/op    0 B/op    0 allocs
BenchmarkYearCE                    329,784,787    3.642 ns/op    0 B/op    0 allocs
BenchmarkYearBE                    322,362,559    3.692 ns/op    0 B/op    0 allocs
BenchmarkFormat                     17,508,992   66.68 ns/op   24 B/op    1 allocs
BenchmarkFormatBE                    1,201,922  996.8 ns/op  217 B/op   14 allocs
BenchmarkString                     13,287,068   87.05 ns/op   32 B/op    1 allocs
```

### Performance Optimization

The gotime package includes comprehensive performance optimizations:

- **Single-Pass String Replacement**: O(n) algorithm replacing iterative O(n×m) approach
- **Regex Compilation Caching**: Pre-compiled patterns eliminate runtime compilation
- **Era Year Caching**: LRU cache reduces redundant FromCE() calculations by 80%+
- **Builder Pooling**: Reuses strings.Builder instances for reduced allocations

See [OPTIMIZATION.md](OPTIMIZATION.md) for detailed documentation.

### Performance Metrics

| Operation | Time | Allocations | Notes |
|-----------|------|-------------|-------|
| Year() (cached) | ~5ns | 0 B | 90% faster for repeated calls |
| String replacement | O(n) | 1 alloc | 70%+ fewer allocs than iterative |
| CE formatting | ~68ns | 24 B | Minimal overhead |
| BE formatting | ~997ns | 217 B | Era conversion required |

## Optimization Documentation

The package includes comprehensive optimization documentation:

### Main Documentation

- **[OPTIMIZATION.md](OPTIMIZATION.md)** - Complete overview of refactoring and optimization work

### Component Documentation

- **[docs/string-replacer.md](docs/string-replacer.md)** - Single-pass string replacement algorithm
- **[docs/regex-pool.md](docs/regex-pool.md)** - Regex compilation caching
- **[docs/era-cache.md](docs/era-cache.md)** - Era year conversion caching
- **[docs/builder-pool.md](docs/builder-pool.md)** - strings.Builder pooling

### Architecture

The optimization layer includes four internal components:

```
┌─────────────────────────────────────────────────────────────┐
│                    gotime Package                            │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│  │StringReplacer│ │ RegexPool   │ │ EraCache    │            │
│  │(single-pass)│ │(pre-compile)│ │(LRU cache)  │            │
│  └─────────────┘ └─────────────┘ └─────────────┘            │
└─────────────────────────────────────────────────────────────┘
```

### Key Optimizations

| Component | Improvement | Location |
|-----------|-------------|----------|
| StringReplacer | 70% fewer allocations | [`internal/replacer.go`](internal/replacer.go) |
| RegexPool | 60% fewer allocations | [`internal/regex_pool.go`](internal/regex_pool.go) |
| EraCache | 80%+ cache hit rate | [`internal/era_cache.go`](internal/era_cache.go) |
| BuilderPool | 40% fewer allocations | [`internal/builder_pool.go`](internal/builder_pool.go) |

## Advanced Features

### Error Codes

The package provides structured error handling with error codes:

```go
import "github.com/bouroo/go-time"

// Check error type
if time.IsParseError(err) {
    // Handle parse error
}

if time.IsValidationError(err) {
    // Handle validation error
}

// Get error code for programmatic handling
code := time.GetErrorCode(err)
switch code {
case time.ErrCodeInvalidFormat:
    // ...
case time.ErrCodeInvalidEra:
    // ...
}
```

### Error Codes Reference

| Code | Description |
|------|-------------|
| `ErrCodeInvalidFormat` | Invalid format string |
| `ErrCodeInvalidTime` | Invalid time value |
| `ErrCodeInvalidEra` | Invalid era specified |
| `ErrCodeEraMismatch` | Era/time mismatch |
| `ErrCodeThaiText` | Thai text processing error |
| `ErrCodeOutOfBounds` | Value out of bounds |

### Validation Errors

```go
// ValidationError for field validation failures
type ValidationError struct {
    Field      string
    Value      any
    Constraint string
}

// TimeValidationError for time value bounds
type TimeValidationError struct {
    Field    string
    Value    any
    MinValue any
    MaxValue any
}

// MultiError for batch operations
me := time.NewMultiError()
me.Add(err1)
me.Add(err2)
if me.HasErrors() {
    // Handle errors
}
```

### Configuration Functions

For deterministic testing and cache management:

```go
import "time"

// Set reference date for era detection (default: time.Now())
time.SetEraDetectionReferenceDate(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

// Set reference date for year formatting
time.SetYearFormatReferenceDate(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

// Clear the global era cache
time.ClearEraCache()

// Get cache statistics
stats := time.EraCacheStats()
fmt.Printf("Hits: %d, Misses: %d\n", stats.Hits, stats.Misses)

// Get cache hit rate
hitRate := time.EraCacheHitRate()
fmt.Printf("Hit Rate: %.1f%%\n", hitRate*100)
```

## Compatibility

| Requirement | Details                                  |
|-------------|------------------------------------------|
| Go Version  | 1.18+                                    |
| `time.Time` | 100% compatible                          |
| Backward    | Yes - existing code requires no changes  |

## Contributing

Contributions are welcome! Please read the contributing guidelines before submitting PRs.

## License

MIT License - see the [LICENSE](LICENSE) file for details.
