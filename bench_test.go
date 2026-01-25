package time

import (
	"testing"
	stdtime "time"

	"github.com/bouroo/go-time/internal"
)

// Benchmark functions for time library operations

func BenchmarkDate(b *testing.B) {
	b.ReportAllocs()
	loc, _ := stdtime.LoadLocation("UTC")
	for b.Loop() {
		_ = Date(2024, 2, 29, 12, 30, 45, 123456789, loc)
	}
}

func BenchmarkNow(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = Now()
	}
}

func BenchmarkInEraCE(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	ce := CE()
	for b.Loop() {
		_ = tm.InEra(ce)
	}
}

func BenchmarkInEraBE(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	be := BE()
	for b.Loop() {
		_ = tm.InEra(be)
	}
}

func BenchmarkYearCE(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.YearCE()
	}
}

func BenchmarkYearBE(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	beTime := tm.InEra(BE())
	for b.Loop() {
		_ = beTime.Year()
	}
}

func BenchmarkIsLeap(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.IsLeap()
	}
}

func BenchmarkIsCE(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.IsCE()
	}
}

func BenchmarkIsBE(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	beTime := tm.InEra(BE())
	for b.Loop() {
		_ = beTime.IsBE()
	}
}

func BenchmarkFormat(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.Format("2006-01-02 15:04:05")
	}
}

func BenchmarkFormatBE(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	beTime := tm.InEra(BE())
	for b.Loop() {
		_ = beTime.Format("2006-01-02 15:04:05")
	}
}

func BenchmarkString(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.String()
	}
}

func BenchmarkAdd(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.Add(stdtime.Hour)
	}
}

func BenchmarkSub(b *testing.B) {
	b.ReportAllocs()
	tm1 := Date(2024, 2, 29, 12, 0, 0, 0, stdtime.UTC)
	tm2 := Date(2024, 2, 29, 13, 0, 0, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm1.Sub(tm2)
	}
}

func BenchmarkBefore(b *testing.B) {
	b.ReportAllocs()
	tm1 := Date(2024, 2, 29, 12, 0, 0, 0, stdtime.UTC)
	tm2 := Date(2024, 2, 29, 13, 0, 0, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm1.Before(tm2)
	}
}

func BenchmarkAfter(b *testing.B) {
	b.ReportAllocs()
	tm1 := Date(2024, 2, 29, 12, 0, 0, 0, stdtime.UTC)
	tm2 := Date(2024, 2, 29, 13, 0, 0, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm2.After(tm1)
	}
}

func BenchmarkEqual(b *testing.B) {
	b.ReportAllocs()
	tm1 := Date(2024, 2, 29, 12, 0, 0, 0, stdtime.UTC)
	tm2 := Date(2024, 2, 29, 12, 0, 0, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm1.Equal(tm2)
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_, _ = tm.MarshalJSON()
	}
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	b.ReportAllocs()
	data := []byte(`"2024-02-29T12:30:45Z"`)
	var tm Time
	for b.Loop() {
		_ = tm.UnmarshalJSON(data)
	}
}

func BenchmarkGobEncode(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 123456789, stdtime.UTC)
	for b.Loop() {
		_, _ = tm.GobEncode()
	}
}

func BenchmarkGobDecode(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 123456789, stdtime.UTC)
	data, _ := tm.GobEncode()
	for b.Loop() {
		var decoded Time
		_ = decoded.GobDecode(data)
	}
}

func BenchmarkUnix(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.Unix()
	}
}

func BenchmarkUnixNano(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.UnixNano()
	}
}

func BenchmarkEraCE(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.Era()
	}
}

func BenchmarkEraBE(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	beTime := tm.InEra(BE())
	for b.Loop() {
		_ = beTime.Era()
	}
}

func BenchmarkLocation(b *testing.B) {
	b.ReportAllocs()
	loc, _ := stdtime.LoadLocation("America/New_York")
	tm := Date(2024, 2, 29, 12, 30, 45, 0, loc)
	for b.Loop() {
		_ = tm.Location()
	}
}

func BenchmarkDay(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.Day()
	}
}

func BenchmarkMonth(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.Month()
	}
}

func BenchmarkHour(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.Hour()
	}
}

func BenchmarkMinute(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.Minute()
	}
}

func BenchmarkSecond(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	for b.Loop() {
		_ = tm.Second()
	}
}

func BenchmarkNanosecond(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 123456789, stdtime.UTC)
	for b.Loop() {
		_ = tm.Nanosecond()
	}
}

// Performance regression benchmarks for hot paths (Phase 3)

// BenchmarkFormatLocaleThai benchmarks FormatLocale with Thai locale - a hot path
func BenchmarkFormatLocaleThai(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	beTime := tm.InEra(BE())
	for b.Loop() {
		_ = beTime.FormatLocale(LocaleThTH, "02 January 2006")
	}
}

// BenchmarkFormatLocaleThaiFullDate benchmarks FormatLocale with full date format
func BenchmarkFormatLocaleThaiFullDate(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	beTime := tm.InEra(BE())
	for b.Loop() {
		_ = beTime.FormatLocale(LocaleThTH, "Monday, 02 January 2006 15:04:05")
	}
}

// BenchmarkYearBECacheHit benchmarks Year() for BE with cache hit (repeated calls)
func BenchmarkYearBECacheHit(b *testing.B) {
	b.ReportAllocs()
	tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
	beTime := tm.InEra(BE())
	// Warm up the cache
	_ = beTime.Year()
	for b.Loop() {
		_ = beTime.Year()
	}
}

// BenchmarkReplaceYearInFormatted benchmarks the year replacement hot path
func BenchmarkReplaceYearInFormatted(b *testing.B) {
	b.ReportAllocs()
	formatted := "29 February 2024 12:30:45"
	// Use a fixed reference date for consistent benchmarks
	SetYearFormatReferenceDate(stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC))
	for b.Loop() {
		_ = replaceYearInFormatted(formatted, 2567)
	}
}

// BenchmarkReplaceYearInFormattedShortYear benchmarks short year replacement
func BenchmarkReplaceYearInFormattedShortYear(b *testing.B) {
	b.ReportAllocs()
	formatted := "29/02/24 12:30:45"
	// Use a fixed reference date for consistent benchmarks
	SetYearFormatReferenceDate(stdtime.Date(2024, 1, 1, 0, 0, 0, 0, stdtime.UTC))
	for b.Loop() {
		_ = replaceYearInFormatted(formatted, 67)
	}
}

// BenchmarkBuilderPoolGetGet benchmarks BuilderPool.Get() performance
func BenchmarkBuilderPoolGet(b *testing.B) {
	b.ReportAllocs()
	bp := internal.NewBuilderPool()
	for b.Loop() {
		builder := bp.Get(256)
		builder.WriteString("test")
		_ = builder.String()
		bp.Put(builder)
	}
}

// BenchmarkEraCacheGet benchmarks EraCache.Get() with cache hit
func BenchmarkEraCacheGet(b *testing.B) {
	b.ReportAllocs()
	ec := internal.NewEraCache(1024)
	// Pre-populate cache
	ec.Set(2024, nil, 2567)
	for b.Loop() {
		_, _ = ec.Get(2024, nil)
	}
}

// BenchmarkEraCacheSet benchmarks EraCache.Set() performance
func BenchmarkEraCacheSet(b *testing.B) {
	b.ReportAllocs()
	ec := internal.NewEraCache(1024)
	for b.Loop() {
		ec.Set(2024, nil, 2567)
	}
}

// BenchmarkStringReplacerReplace benchmarks StringReplacer.Replace() performance
func BenchmarkStringReplacerReplace(b *testing.B) {
	b.ReportAllocs()
	sr := internal.NewStringReplacer(map[string]string{
		"January":   "มกราคม",
		"February":  "กุมภาพันธ์",
		"March":     "มีนาคม",
		"April":     "เมษายน",
		"May":       "พฤษภาคม",
		"June":      "มิถุนายน",
		"July":      "กรกฎาคม",
		"August":    "สิงหาคม",
		"September": "กันยายน",
		"October":   "ตุลาคม",
		"November":  "พฤศจิกายน",
		"December":  "ธันวาคม",
	})
	input := "January February March April May June July August September October November December"
	for b.Loop() {
		_ = sr.Replace(input)
	}
}

// BenchmarkCombinedThaiLocaleReplace benchmarks combined Thai locale replacement
func BenchmarkCombinedThaiLocaleReplace(b *testing.B) {
	b.ReportAllocs()
	input := "Monday, 29 February 2024"
	for b.Loop() {
		_ = replaceThaiLocale(input)
	}
}

// Concurrent benchmarks for thread safety verification

func BenchmarkConcurrentFormatLocaleThai(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
		beTime := tm.InEra(BE())
		for pb.Next() {
			_ = beTime.FormatLocale(LocaleThTH, "02 January 2006")
		}
	})
}

func BenchmarkConcurrentYearBE(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		tm := Date(2024, 2, 29, 12, 30, 45, 0, stdtime.UTC)
		beTime := tm.InEra(BE())
		for pb.Next() {
			_ = beTime.Year()
		}
	})
}

func BenchmarkConcurrentEraCache(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ec := internal.NewEraCache(1024)
		for pb.Next() {
			ec.Set(2024, nil, 2567)
			_, _ = ec.Get(2024, nil)
		}
	})
}
