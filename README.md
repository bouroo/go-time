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
t := gotime.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)

// Convert to Buddhist Era
beTime := t.InEra(gotime.BE())
fmt.Printf("BE Year: %d\n", beTime.Year())    // 2567
fmt.Printf("CE Year: %d\n", beTime.YearCE())  // 2024
```

### Thai Date Formatting

```go
beTime := gotime.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC).InEra(gotime.BE())

// Format with Thai month names
thai := beTime.FormatLocale(gotime.LocaleThTH, "02 January 2006")
// Output: "2 มกราคม 2549"

// Format day of week in Thai
dayName := beTime.FormatLocale(gotime.LocaleThTH, "Monday")
// Output: "วันจันทร์"
```

### Parsing Thai Dates

```go
// Parse with explicit era
t, err := gotime.ParseWithEra("02 January 2006", "15 มกราคม 2567", gotime.BE())
if err != nil {
    log.Fatal(err)
}
fmt.Printf("CE Year: %d\n", t.YearCE())  // 2024

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
t, err := gotime.ParseWithEra("2006-01-02", "invalid-date", gotime.BE())
if err != nil {
    fmt.Printf("Error at offset %d: %v\n", err.Offset, err.Err)
}
```

## Performance

Benchmarks run on Go 1.25.5, darwin/arm64:

```
BenchmarkDate
BenchmarkDate-14                188699502                6.146 ns/op           0 B/op          0 allocs/op
BenchmarkNow
BenchmarkNow-14                 43979469                26.24 ns/op            0 B/op          0 allocs/op
BenchmarkInEraCE
BenchmarkInEraCE-14             742275505                1.590 ns/op           0 B/op          0 allocs/op
BenchmarkInEraBE
BenchmarkInEraBE-14             757681784                1.625 ns/op           0 B/op          0 allocs/op
BenchmarkYearCE
BenchmarkYearCE-14              329784787                3.642 ns/op           0 B/op          0 allocs/op
BenchmarkYearBE
BenchmarkYearBE-14              322362559                3.692 ns/op           0 B/op          0 allocs/op
BenchmarkIsLeap
BenchmarkIsLeap-14              291164432                4.120 ns/op           0 B/op          0 allocs/op
BenchmarkIsCE
BenchmarkIsCE-14                720262174                1.598 ns/op           0 B/op          0 allocs/op
BenchmarkIsBE
BenchmarkIsBE-14                744628588                1.600 ns/op           0 B/op          0 allocs/op
BenchmarkFormat
BenchmarkFormat-14              17508992                66.68 ns/op           24 B/op          1 allocs/op
BenchmarkFormatBE
BenchmarkFormatBE-14             1201922               996.8 ns/op           217 B/op         14 allocs/op
BenchmarkString
BenchmarkString-14              13287068                87.05 ns/op           32 B/op          1 allocs/op
BenchmarkAdd
BenchmarkAdd-14                 411226130                2.902 ns/op           0 B/op          0 allocs/op
BenchmarkSub
BenchmarkSub-14                 235008325                5.109 ns/op           0 B/op          0 allocs/op
BenchmarkBefore
BenchmarkBefore-14              681499858                1.765 ns/op           0 B/op          0 allocs/op
BenchmarkAfter
BenchmarkAfter-14               667542660                1.771 ns/op           0 B/op          0 allocs/op
BenchmarkEqual
BenchmarkEqual-14               675964306                1.784 ns/op           0 B/op          0 allocs/op
BenchmarkMarshalJSON
BenchmarkMarshalJSON-14         46805750                24.97 ns/op           48 B/op          1 allocs/op
BenchmarkUnmarshalJSON
BenchmarkUnmarshalJSON-14       69837481                16.95 ns/op            0 B/op          0 allocs/op
BenchmarkGobEncode
BenchmarkGobEncode-14           100000000               10.17 ns/op           16 B/op          1 allocs/op
BenchmarkGobDecode
BenchmarkGobDecode-14           486095229                2.480 ns/op           0 B/op          0 allocs/op
BenchmarkUnix
BenchmarkUnix-14                757660658                1.580 ns/op           0 B/op          0 allocs/op
BenchmarkUnixNano
BenchmarkUnixNano-14            741669543                1.611 ns/op           0 B/op          0 allocs/op
BenchmarkEraCE
BenchmarkEraCE-14               728046922                1.602 ns/op           0 B/op          0 allocs/op
BenchmarkEraBE
BenchmarkEraBE-14               755261457                1.586 ns/op           0 B/op          0 allocs/op
BenchmarkLocation
BenchmarkLocation-14            719289343                1.589 ns/op           0 B/op          0 allocs/op
BenchmarkDay
BenchmarkDay-14                 320286014                3.745 ns/op           0 B/op          0 allocs/op
BenchmarkMonth
BenchmarkMonth-14               307425740                3.903 ns/op           0 B/op          0 allocs/op
BenchmarkHour
BenchmarkHour-14                411464016                2.929 ns/op           0 B/op          0 allocs/op
BenchmarkMinute
BenchmarkMinute-14              413787930                2.925 ns/op           0 B/op          0 allocs/op
BenchmarkSecond
BenchmarkSecond-14              452139042                2.663 ns/op           0 B/op          0 allocs/op
BenchmarkNanosecond
BenchmarkNanosecond-14          759756940                1.574 ns/op           0 B/op          0 allocs/op
```

### Performance Notes

- Standard CE formatting: ~68ns/op (minimal overhead)
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
