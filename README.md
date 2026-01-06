# go-time

A Go library for Thai/Buddhist Era (BE) date support with seamless `time.Time` compatibility.

[![Go Reference](https://pkg.go.dev/badge/github.com/bouroo/go-time.svg)](https://pkg.go.dev/github.com/bouroo/go-time)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## What is go-time?

go-time provides comprehensive Thai calendar support for Go applications. It enables developers to work with Buddhist Era dates naturally while maintaining full compatibility with the standard library's `time.Time` type.

### Key Capabilities

- **Era Conversion**: Convert between Anno Domini (AD) and Buddhist Era (BE) effortlessly
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
    "github.com/bouroo/go-time/gotime"
)

// Create a standard time
t := gotime.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)

// Convert to Buddhist Era
beTime := t.InEra(gotime.BE())
fmt.Printf("BE Year: %d\n", beTime.Year())    // 2567
fmt.Printf("AD Year: %d\n", beTime.YearAD())  // 2024
```

### Thai Date Formatting

```go
beTime := gotime.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC).InEra(gotime.BE())

// Format with Thai month names
thai := beTime.FormatLocale(gotime.LocaleThTH, "02 January 2006")
// Output: "29 กุมภาพันธ์ 2567"

// Format day of week in Thai
dayName := beTime.FormatLocale(gotime.LocaleThTH, "Monday")
// Output: "วันพฤหัสบดี"
```

### Parsing Thai Dates

```go
// Parse with explicit era
t, err := gotime.ParseWithEra("02 January 2006", "15 มกราคม 2567", gotime.BE())
if err != nil {
    log.Fatal(err)
}
fmt.Printf("AD Year: %d\n", t.YearAD())  // 2024

// Auto-detect era from year value
t, err := gotime.ParseThai("02/01/2006", "15/01/2567")
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
| Monday     | วันจันทร์      | จ.         |
| Tuesday    | วันอังคาร     | อ.         |
| Wednesday  | วันพุธ        | พ.         |
| Thursday   | วันพฤหัสบดี    | พฤ.        |
| Friday     | วันศุกร์       | ศ.         |
| Saturday   | วันเสาร์       | ส.         |
| Sunday     | วันอาทิตย์     | อา.        |

## API Reference

### Core Types

| Type     | Description                                    |
|----------|------------------------------------------------|
| `Time`   | Wraps `time.Time` with era support             |
| `Era`    | Represents calendar eras (AD, BE)              |
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
| `YearAD() int`            | `int`     | Year in AD regardless of era    |
| `InEra(e *Era) Time`      | `Time`    | Convert to different era        |
| `Format(layout string)`   | `string`  | Format using Go layout          |
| `FormatLocale(locale, layout)` | `string` | Format with locale translations |

### Era Functions

| Function                       | Returns   | Description                          |
|--------------------------------|-----------|--------------------------------------|
| `AD() *Era`                    | `*Era`    | Returns AD era instance              |
| `BE() *Era`                    | `*Era`    | Returns BE era instance              |
| `DetectEraFromYear(year int)`  | `*Era`    | Detect era from year value (2501-2599 → BE) |

### Error Handling

```go
// ParseError provides detailed error context
type ParseError struct {
    // Contains layout, value, and offset information
}

// Example error handling
t, err := gotime.ParseWithEra("2006-01-02", "invalid-date", gotime.BE())
if err != nil {
    fmt.Printf("Error at offset %d: %v\n", err.Offset, err.Err)
}
```

## Performance

Benchmarks run on Go 1.25.5, darwin/arm64:

```
BenchmarkTimeFormatAD-14                 17.3M  68ns/op   24 B/op   1 allocs/op
BenchmarkTimeFormatBE-14                  1.2M  938ns/op  209 B/op  13 allocs/op
BenchmarkTimeFormatLocaleThai-14          901K  1310ns/op 315 B/op  12 allocs/op
BenchmarkParseThai-14                     2.1M  562ns/op   16 B/op   1 allocs/op
BenchmarkParseThaiAutoDetect-14           2.6M  464ns/op    0 B/op   0 allocs/op
BenchmarkParseWithEra-14                  1.3M  870ns/op   81 B/op   6 allocs/op
BenchmarkParseInLocationWithEra-14        1.3M  903ns/op   72 B/op   6 allocs/op
```

### Performance Notes

- Standard AD formatting: ~68ns/op (minimal overhead)
- BE formatting: ~938ns/op (era conversion required)
- Thai locale formatting: ~1310ns/op (string translation)
- Auto-detection parsing: ~464ns/op (fastest option)

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
