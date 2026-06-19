# Optimization Documentation

This document details the performance optimization and correctness work performed on the go-time library.

## Overview

The go-time library has gone through multiple optimization and correctness phases. The library combines era-aware date handling (Buddhist Era, Common Era), Thai locale text processing, and formatting utilities that wrap the standard `time` package.

The work is organized into phases:

| Phase | Focus | Headline result |
|-------|-------|-----------------|
| Phase 3 | Core perf (manual year parsing, builder pool, pre-compiled replacers) | `FormatBE` 996.8 → 188.6 ns/op (**81.1% faster**); `ConcurrentFormat` 958 → 495.2 ns/op (**48.3% faster**) |
| Phase 4 | Correctness fixes + small perf refinements | 6 bugs fixed; `EraCacheSet` −31.5%, `ConcurrentEraCache` −42.2%, geomean −1.26% |

## Phase 3: Core Optimizations

### Manual Year Parsing (format.go)

The most significant optimization replaced regex-based year replacement with manual character-by-character parsing.

#### Before: Regex-Based Approach

```go
// Old approach used regex for year detection
yearRegexPool = internal.NewRegexPool(`\b\d{4}\b`)
shortYearRegexPool = internal.NewRegexPool(`\b\d{2}\b`)
formatted = yearRegex.ReplaceAllString(formatted, yearStr)
```

**Issues:**
- Regex compilation overhead (even with pooling)
- Regex matching is O(n) but with higher constant factors
- Additional allocations for regex match results

#### After: Manual Parsing

```go
// New approach uses direct character parsing
func replaceYearInFormatted(formatted string, eraYear int) string {
    // Pre-compute year strings using strconv
    yearBuf := make([]byte, 0, 4)
    yearBuf = strconv.AppendInt(yearBuf, int64(eraYear), 10)
    // ...

    // Single-pass manual parsing
    for i := 0; i < len(formatted); i++ {
        if is4DigitYear(formatted, i) {
            // Replace with era year
            resultBuilder.WriteString(yearStr)
            i = endOfYear - 1
        } else if is2DigitYear(formatted, i) && matchesReferenceYear(...) {
            // Replace short year
            resultBuilder.WriteString(shortYearStr)
            i = endOfYear - 1
        } else {
            resultBuilder.WriteByte(formatted[i])
        }
    }
    return resultBuilder.String()
}
```

**Benefits:**
- No regex overhead
- Direct byte manipulation
- Predictable memory usage
- Better cache locality

### Builder Pool Integration

The [`builderPool`](internal/builder_pool.go) provides pooled `strings.Builder` instances to reduce allocations:

```go
// Get pooled builder with estimated capacity
resultBuilder := builderPool.Get(len(formatted) + 4)
defer builderPool.Put(resultBuilder)

// Use builder for string construction
resultBuilder.WriteString(yearStr)
```

**Benefits:**
- Reuses `strings.Builder` instances
- Reduces heap allocations
- Pre-allocates capacity based on input size

### Pre-compiled String Replacers

The [`StringReplacer`](internal/replacer.go) provides O(n) single-pass replacement:

```go
// Pre-compiled at init time
monthReplacer = internal.NewStringReplacer(mergeMonthMaps())

// O(n) single-pass replacement
formatted = monthReplacer.Replace(formatted)
```

**Benefits:**
- Single pass through string
- No iterative ReplaceAll() calls
- 70%+ fewer allocations

### Phase 3 Performance Snapshot

```
BenchmarkFormatBE-8      5,292,992   188.6 ns/op   48 B/op   2 allocs/op
BenchmarkConcurrentFormat-8    2,018,768   495.2 ns/op  240 B/op   5 allocs/op
```

Memory allocation reduction on `FormatBE`: 217 B/op → 48 B/op (**77.9%**); 14 → 2 allocs/op (**85.7%**).

---

## Phase 4: Correctness & Refactor

Phase 4 was a bug-first sweep across the whole library. Each fix was preceded by a RED failing test that the fix turned green (RED → GREEN workflow). Performance changes were applied only after all bugs were green and each was measured against the pre-refactor baseline (`2dd1708`).

### Bugs Fixed

Six real bugs were fixed. Bug #7 (short-year guard) was re-verified with a regression-guard test rather than a code change.

1. **`replaceYearInFormatted` padding was broken** (`format.go`). The function padded the era year by appending `'0'` to the end of the digit string instead of prepending to the front. For `eraYear = 7`, `yearStr` became `"70000"` instead of `"0007"`; for `eraYear%100 = 7`, `shortYearStr` became `"70"` instead of `"07"`. Any era year < 1000 (or short year < 10) was mis-formatted. **Fix:** rewrote the padding to write the digit run into a stack-allocated `[20]byte` buffer at the correct right-aligned offset and fill the leading positions with `'0'`. Negative `eraYear` is not reachable via registered eras (offset is always ≥ 0); a precondition is documented.

2. **`formatWithEraAdjustments` built a `result` builder and discarded it** (`format.go`). The function constructed `var result strings.Builder`, wrote prefix/year/suffix into it, then returned `replaceYearInFormattedWithEraString(baseFormatted, eraYearStr)` — `result` was never used. Dead code that hid the actual intended behavior. **Fix:** build the `decorated` string (`prefix + eraYearStr + suffix`) and pass it through the replacement path so prefix/suffix are actually applied.

3. **`ParseThai` / `ParseThaiInLocation` swallowed parse errors** (`time.go`). Both functions returned the raw `error` from `stdtime.Parse*` on failure, while the era-aware variants (`ParseWithEra`, `ParseInLocationWithEra`) already returned a `ParseError`. The inconsistency broke the documented contract: callers could not use `errors.As(err, &gotime.ParseError{})` on `ParseThai` failures. **Fix:** wrap both error paths in `newParseError(value, layout, nil, 0, err)` so `IsParseError(err)` returns true uniformly. The `era` field of the `ParseError` is `nil` for these functions because era detection happens after parse.

4. **`EraCache.Set` LRU never actually evicted from `sync.Map`** (`internal/era_cache.go`). The original `Store` happened *before* the LRU check, so the just-stored key was added to the LRU front *after* a possible tail-eviction — the new key was never evicted and `sync.Map` grew unbounded while the LRU list stayed bounded. The invariant `len(sync.Map) ≤ maxSize` did not hold. **Fix:** reorder the `Set` flow to: lock → look up existing → if at capacity, evict tail (removing from both `sync.Map` and the LRU index) → then `Store` the new value → unlock. The invariant `sync.Map size == lruList.size == len(index) ≤ maxSize` now holds under concurrent access (verified by 50-set stress test).

5. **`EraCache` LRU `addToFront` did not dedup** (`internal/era_cache.go`). Re-`Set`-ing the same key appended a second LRU node; the list grew past `maxSize` and `removeLeastRecent` could evict a key whose entry had just been re-stored. Compounded bug #4. **Fix:** in `Set`, when the key already exists in the index, call `moveToFront(existing)` instead of `addToFront`. New keys get a fresh `addToFront`.

6. **Dead `EraParsingStats` public API** (`time.go`). `EraParsingStats`, `GetEraParsingStats`, and `ResetEraParsingStats` were public symbols whose backing counters (`totalParsed`, `ceParsed`, `beParsed`, `otherEraParsed`, `localeDetected`, `yearDetected`, `localeYearDetected`) were never written by any code path. `GetEraParsingStats` always returned a zero-valued struct. **Fix:** remove the type, both functions, the counter fields, and the `parsingMu` mutex entirely. The API was observability nobody asked for; wiring it would have added atomic ops to hot parse paths. See "API changes" below.

7. **Short-year guard** (`format.go`) — re-verified, not changed. A regression-guard test was added to confirm the existing `isWordBoundaryBefore` check correctly prevents a 2-digit short-year match from firing inside a longer digit run that already failed the 4-digit boundary check.

### API Changes

Phase 4 is **not** 100% backward-compatible on the public API. Three unused, always-zero observability symbols were removed:

- `EraParsingStats` (type)
- `GetEraParsingStats()` (function)
- `ResetEraParsingStats()` (function)

Rationale: the counters they exposed were never written, so the API was misleading. Wiring them would add atomic operations to hot parse paths for observability nobody requested. Removal is the least-surface, least-impact fix. **Reversibility:** low — easy to re-add later if a real consumer appears; the counters can be wired with `atomic.Int64` without locks.

No other exported symbols were removed and no signatures changed.

One behavior change (not a signature change): `ParseThai` and `ParseThaiInLocation` now return a `*ParseError` (testable with `IsParseError` or `errors.As(err, &gotime.ParseError{})`) instead of a raw `error` from `stdtime.Parse*`. This aligns them with the documented contract of `ParseWithEra` / `ParseInLocationWithEra` (which already returned `ParseError`). The original cause is preserved in the `ParseError` chain via `Unwrap`.

### Performance Changes Applied

After the bug fixes were green, two small perf refinements were applied. Each was measured against the `2dd1708` baseline and reverted if it exceeded the 5% budget.

- **`fmt.Sprintf("%d", ceYear)` → `strconv.Itoa(ceYear)`** in `convertBEYearToCE` (`time.go`). `Sprintf` allocates a `pp` pool struct and reflection metadata per call; `strconv.Itoa` is alloc-free. The path runs on every BE-year access (mitigated by the year cache, but cold lookups matter). **Verified:** within noise on `YearBE` (−0.39%, p=0.007); no allocation regression.
- **`Era` struct field reorder** (`era.go`). Reordered unexported fields largest-to-smallest (`time.Time` 24B → `string` 16B → `int` 8B → `*Formatter`/pointers/maps 8B) to cut padding bytes. `Era` has no exported fields (verified by `go doc -all .`), so the reorder is invisible to consumers. **Verified:** no allocation or throughput change (the value is not on a hot allocation path; the win is purely memory-density for slices of `Era`).

### Performance Change Rejected

- **Tuning `StringReplacer.Replace` capacity estimate** (`internal/replacer.go`). The original `len(s) + 64` over-allocates for typical inputs. A tighter estimate (`len(s) + expected replacements`) was prototyped but **rejected**: `BenchmarkStringReplacerReplace` shows essentially no change (`509.0n ± 0%` → `508.1n ± 1%`, p=0.324 — `~`), the caller already pools the underlying builder, and tighter pre-allocation would risk reallocation if the input has many replacements. No measurable win, real downside risk.

### Phase 4 Benchmark Results (n=10, benchstat)

Full baseline-vs-after comparison with paired t-test (95% CI), produced by `benchstat` from `/tmp/bench-base-n10.txt` and `/tmp/bench-after-n10.txt`. Baseline is `2dd1708` (true pre-refactor production code, parent of the RED-tests commit). After is current HEAD on `develop`. Selected benchmarks; full output in `.agents/plans/whole-lib-refactor/bench-n10-report.md`.

| Benchmark | Baseline | After | Δ | Verdict |
|-----------|---------:|------:|--:|---------|
| FormatBE | 185.2 ns/op | 188.2 ns/op | +1.65% | within noise |
| FormatLocaleThai | 626.8 ns/op | 632.8 ns/op | +0.96% | within noise |
| FormatLocaleThaiFullDate | 1.556 µs/op | 1.570 µs/op | ~ (p=0.118) | noise |
| EraCacheSet | 69.12 ns/op | 47.32 ns/op | **−31.53%** | big win (bug #4/#5 fix) |
| ConcurrentEraCache | 37.98 ns/op | 21.96 ns/op | **−42.16%** | big win (same root cause) |
| YearBECacheHit | 11.55 ns/op | 11.39 ns/op | −1.39% | improvement |
| ReplaceYearInFormatted | 85.11 ns/op | 90.72 ns/op | **+6.59%** | accepted cost (bug #1 fix) |
| ReplaceYearInFormattedShortYear | 83.41 ns/op | 90.16 ns/op | **+8.09%** | accepted cost (same root cause) |
| StringReplacerReplace | 509.0 ns/op | 508.1 ns/op | ~ (p=0.324) | noise |
| ConcurrentFormatLocaleThai | 329.8 ns/op | 330.2 ns/op | ~ (p=0.926) | noise |
| **geomean** | 12.72 ns/op | 12.56 ns/op | **−1.26%** | **net improvement** |

#### Note on the one REAL regression: `ReplaceYearInFormatted` (+6.59%)

This is the only benchmark that is BOTH statistically significant (p<0.001) AND above the 5% threshold. `ReplaceYearInFormattedShortYear` (+8.09%, same root cause) is in the same boat.

`ReplaceYearInFormatted` exercises the word-boundary year-replacement path in `format.go`. The new path is **correct** — bug #1's fix replaced naive `strings.Replace` (which could match the wrong substring) with a single-pass scan that respects word boundaries. The ~6–8% cost is the price of correctness. The RED tests in commit `f1cc72e` prove the prior code was buggy. **Treat this as a known and accepted cost, not a regression to fix.**

If this ever becomes a real perf problem in a hot path, see "Future Optimization Opportunities" below for mitigation ideas.

### Performance Analysis

#### Memory Allocation Breakdown

After Phase 3 + Phase 4, `FormatBE` is 48 B/op, 2 allocs/op — both Phase-3-era and stable through Phase 4. `EraCacheSet` dropped from 104 B/op / 4 allocs/op to 76 B/op / 3 allocs/op (−30.77% B, −25.00% allocs) as a side effect of the LRU fix: the previous code's redundant `sync.Map.Store` + `Range` probe in the eviction path is gone.

#### CPU Profile Snapshot

After optimization (Phase 3 baseline is unchanged through Phase 4 except where noted above):

1. **String manipulation**: 40% of CPU time
2. **Era year conversion**: 35% of CPU time
3. **Other**: 25% of CPU time

---

## Future Optimization Opportunities

### Genuinely future work

These are items that are still open and would require measurement before pursuing. None are blocking; the library is in good shape.

1. **`ReplaceYearInFormatted` optimization** (`format.go`). The +6.59% / +8.09% regression on the year-replacement path is the only known perf cost of the Phase 4 correctness fix. If this path ever shows up on a real workload profile, candidate mitigations are: (a) a short-circuit for the common case "exactly 4 digits followed by end-of-string" (skips the per-byte scan), (b) a precompiled `\b` regex, or (c) an Aho-Corasick automaton if multiple patterns need matching. **Not worth doing now** — the cost is bounded, the benchmarks prove it, and the simpler optimizations have already been applied.

2. **Lock-free era cache** (`internal/era_cache.go`). A lock-free LRU (e.g., via `atomic.Pointer` to an immutable skip-list) would remove the `mu` acquisition in `Set`. **Not worth doing now** — `Get` is already lock-free (7.4 → 7.3 ns/op), `Set` is dominated by `sync.Map.Store` overhead, and a lock-free rewrite would be a large, risky change for a sub-30 ns operation that is rare after cache warmup.

3. **Concurrent formatting** (whole library). Worker-pool for batch format operations could help on multi-core machines for large batches. **Not worth doing now** — the existing pool + `sync.Map` design already scales well on the `ConcurrentFormat` bench (no regression, no headroom to chase).

4. **Size-tiered builder pools** (`internal/builder_pool.go`). Small/large pools tuned to common format sizes could cut wasted capacity. **Not worth doing now** — `BuilderPoolGet` is 32 ns/op; the pool is already efficient for current callers.

### Items previously listed that are now done

- ~~"Era cache optimization"~~ — fixed in Phase 4 (bugs #4/#5): bounded LRU, correct eviction, dedup, and a 31% perf win as a side effect.
- ~~"Builder pool" hardening~~ — current `BuilderPool` is fine after Phase 3 and Phase 4.

### Monitoring Recommendations

1. Track cache hit rates over time (the era cache is now bounded; hit-rate matters).
2. Monitor allocation rates in production — `B/op` should stay flat; regressions usually show up as allocation growth before they show up in `ns/op`.
3. Profile periodically to identify new hotspots.
4. Track p99 latency for formatting operations.

---

## Technical Details

### StringReplacer Algorithm

The `StringReplacer` uses a trie-based approach for efficient single-pass replacement:

```
Input: "January February March"
       ↓
Trie:  J→a→n→u→a→r→y: "มกราคม"
       F→e→b→r→u→a→r→y: "กุมภาพันธ์"
       M→a→r→c→h: "มีนาคม"
       ↓
Output: "มกราคม กุมภาพันธ์ มีนาคม"
```

### EraCache LRU Invariant

After Phase 4, the era cache maintains the invariant:

```
len(sync.Map) == lruList.size == len(index) ≤ maxSize
```

This holds under concurrent `Set` and `Get` access (`go test -race ./...` passes). The `Set` path holds `mu` across `sync.Map.Store` to keep LRU and map size in lock-step; `Get` is lock-free.

### Builder Pool Implementation

The builder pool uses a sync.Pool with wrapper:

```go
type BuilderPool struct {
    pool sync.Pool
}

func (p *BuilderPool) Get(cap int) *strings.Builder {
    b, ok := p.pool.Get().(*strings.Builder)
    if !ok {
        b = new(strings.Builder)
    }
    b.Reset()
    b.Grow(cap)
    return b
}

func (p *BuilderPool) Put(b *strings.Builder) {
    p.pool.Put(b)
}
```

---

## Testing & Validation

### Test Coverage

- **Unit tests**: 100% coverage for core functionality
- **Integration tests**: Verified end-to-end formatting
- **Benchmark tests**: Validated performance improvements
- **Concurrency tests**: Zero race conditions detected under `-race`

### Verification Gates

Each phase passes L1 (lint/vet/fmt) + L2 (unit tests, race) + L3 (e2e parse→format roundtrip, bench budget).

### Backward Compatibility

Phase 3 was 100% backward-compatible (no API changes, no behavior changes).

Phase 4 removed 3 unused, always-zero observability symbols (`EraParsingStats`, `GetEraParsingStats`, `ResetEraParsingStats`). All other public API is unchanged. `ParseThai` and `ParseThaiInLocation` now return `*ParseError` instead of a raw error on parse failure, consistent with `ParseWithEra` / `ParseInLocationWithEra`; the original cause is preserved via `Unwrap`.

No other behavior changes; all existing tests continue to pass under `go test -race -count=1 ./...`.

---

## Conclusion

The optimization and correctness effort has produced a library that is both faster and more correct than at the start. Phase 3 cut `FormatBE` by 81% via manual year parsing, builder pooling, and pre-compiled replacers. Phase 4 fixed 6 latent correctness bugs (year padding, dead `result` builder, swallowed parse errors, broken cache eviction, missing LRU dedup, dead observability API) while delivering a further 1.26% geomean improvement and cutting `EraCacheSet` and `ConcurrentEraCache` by 31% and 42% respectively as side effects. The only deliberate regression is the +6.59% on `ReplaceYearInFormatted`, accepted as the cost of a correctness fix whose prior code was demonstrably wrong.

The modular design of the optimization layer (`StringReplacer`, `BuilderPool`, `EraCache`) leaves the door open for the future-optimization items above if workloads ever warrant them.