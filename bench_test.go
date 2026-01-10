package time

import (
	"testing"
	stdtime "time"
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
