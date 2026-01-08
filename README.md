# go-time

A Go library for Thai/Buddhist Era (BE) date support with seamless `time.Time` compatibility.

[![Go Reference](https://pkg.go.dev/badge/github.com/bouroo/go-time.svg)](https://pkg.go.dev/github.com/bouroo/go-time)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Get Started in 5 Minutes

```go
import (
    "fmt"
    "time"
    "github.com/bouroo/go-time"
)
```

### Convert Between Eras

```go
// Create a standard time
t := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)

// Convert to Buddhist Era
beTime := t.InEra(time.BE())
fmt.Printf("BE Year: %d\n", beTime.Year())    // 2567
fmt.Printf("CE Year: %d\n", beTime.YearCE())  // 2024
```

### Format Thai Dates

```go
beTime := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC).InEra(time.BE())

// Format with Thai month names
thai := beTime.FormatLocale(time.LocaleThTH, "02 January 2006")
// Output: "2 มกราคม 2549"

// Format day of week in Thai
dayName := beTime.FormatLocale(time.LocaleThTH, "Monday")
// Output: "จันทร์"
```

### Parse Thai Dates

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

---

## Installation

```bash
go get github.com/bouroo/go-time
```

---

## Parsing Thai Dates

```go
// Parse with explicit era
t, err := time.ParseWithEra("02 January 2006", "15 มกราคม 2567", time.BE())

// Parse in location with era
t, err := time.ParseInLocationWithEra("02 January 2006", "15 มกราคม 2567", time.UTC, time.BE())

// Auto-detect BE era from year (2501-2599)
t, err := time.ParseThai("02/01/2006", "15/01/2567")
t, err := time.ParseThaiInLocation("02/01/2006", "15/01/2567", time.UTC)
```

**Error handling:**

```go
t, err := time.ParseWithEra("2006-01-02", "invalid-date", time.BE())
if err != nil {
    fmt.Printf("Error at offset %d: %v\n", err.Offset, err.Err)
}
```

---

## Formatting Dates

```go
beTime := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC).InEra(time.BE())

// Standard Go layout (uses BE year)
beTime.Format("2006-01-02")  // "2567-02-29"

// Thai locale translations
beTime.FormatLocale(time.LocaleThTH, "02 January 2006")  // "2 มกราคม 2549"
beTime.FormatLocale(time.LocaleThTH, "Monday")           // "จันทร์"
beTime.FormatLocale(time.LocaleThTH, "Mon")              // "จ."
```

**Thai month names:**

| English | Thai Full | Thai Short |
|---------|-----------|------------|
| January | มกราคม | ม.ค. |
| February | กุมภาพันธ์ | ก.พ. |
| March | มีนาคม | มี.ค. |
| April | เมษายน | เม.ย. |
| May | พฤษภาคม | พ.ค. |
| June | มิถุนายน | มิ.ย. |
| July | กรกฎาคม | ก.ค. |
| August | สิงหาคม | ส.ค. |
| September | กันยายน | ก.ย. |
| October | ตุลาคม | ต.ค. |
| November | พฤศจิกายน | พ.ย. |
| December | ธันวาคม | ธ.ค. |

---

## Era Conversion

```go
t := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)

// Convert to Buddhist Era
beTime := t.InEra(time.BE())
beTime.Year()     // 2567 (BE year)
beTime.YearCE()   // 2024 (CE year always)

// Convert back to CE
ceTime := beTime.InEra(time.CE())
ceTime.Year()     // 2024
```

**Era functions:**

```go
time.BE()  // Buddhist Era
time.CE()  // Anno Domini
time.DetectEraFromYear(2567)  // Returns BE
```

---

## Error Codes

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
    // Invalid format string
case time.ErrCodeInvalidTime:
    // Invalid time value
case time.ErrCodeInvalidEra:
    // Invalid era specified
case time.ErrCodeEraMismatch:
    // Era/time mismatch
case time.ErrCodeThaiText:
    // Thai text processing error
case time.ErrCodeOutOfBounds:
    // Value out of bounds
}
```

**Validation errors:**

```go
type ValidationError struct {
    Field      string
    Value      any
    Constraint string
}

type TimeValidationError struct {
    Field    string
    Value    any
    MinValue any
    MaxValue any
}

me := time.NewMultiError()
me.Add(err1)
me.Add(err2)
if me.HasErrors() {
    // Handle errors
}
```

---

## Performance

Benchmarks (Go 1.25.5, darwin/arm64):

| Operation | Time | Allocations |
|-----------|------|-------------|
| Date creation | 6.1 ns/op | 0 B |
| InEra conversion | 1.6 ns/op | 0 B |
| Year() (cached) | 3.7 ns/op | 0 B |
| CE formatting | 67 ns/op | 24 B |
| BE formatting | 189 ns/op | 48 B |

**Key optimizations (81% faster BE formatting):**

- Single-pass string replacement (O(n) vs O(n×m))
- Manual year parsing (no regex overhead)
- LRU cache for era conversions (80%+ hit rate)
- strings.Builder pooling

See [OPTIMIZATION.md](OPTIMIZATION.md) for details.

---

## Configuration

```go
import "time"

// Set reference date for era detection (default: time.Now())
time.SetEraDetectionReferenceDate(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

// Clear the global era cache
time.ClearEraCache()

// Get cache statistics
stats := time.EraCacheStats()
fmt.Printf("Hits: %d, Misses: %d\n", stats.Hits, stats.Misses)
```

---

## Testing

```bash
# Run all tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -cover ./...

# Run benchmarks
go test -bench=. -benchmem ./...
```

**Test coverage:** 79.5% overall, 95% internal packages.

---

## Compatibility

| Requirement | Details |
|-------------|---------|
| Go Version | 1.18+ |
| `time.Time` | 100% compatible |
| Backward Compatible | Yes |

---

## License

MIT License - see [LICENSE](LICENSE) file.
