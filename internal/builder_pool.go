// Package internal provides internal utilities for the gotime package.
// This package is not part of the public API and may be changed at any time.
package internal

import (
	"strings"
	"sync"
	"sync/atomic"
)

// BuilderPool provides thread-safe pooling of strings.Builder instances.
// This reduces memory allocations by reusing builder objects across multiple
// formatting operations, eliminating the overhead of creating new builders
// for each string construction.
//
// Performance characteristics:
//   - Get/Put operations: O(1) amortized
//   - Thread safety: Uses sync.Pool for concurrent access
//   - Memory: Reduces allocations through object reuse
//
// Usage:
//
//	bp := NewBuilderPool()
//	builder := bp.Get(256) // 256 is the capacity hint
//	defer bp.Put(builder)
//	builder.WriteString("part1")
//	builder.WriteString("part2")
//	result := builder.String()
type BuilderPool struct {
	pool  sync.Pool
	stats poolStats
}

// poolStats tracks pool performance metrics for monitoring and optimization.
type poolStats struct {
	Gets      int64
	Puts      int64
	Allocates int64
}

// DefaultBuilderCapacity is the default pre-allocated capacity for new builders.
// This provides a good balance between memory usage and performance for typical
// formatting operations (date strings, year formatting, etc.).
const DefaultBuilderCapacity = 256

// MaxBuilderCapacity is the maximum capacity for pooled builders.
// Builders larger than this are not returned to the pool to prevent memory bloat
// from retaining very large buffers.
const MaxBuilderCapacity = 4096

// NewBuilderPool creates a new BuilderPool with default settings.
// The pool will pre-allocate builders with DefaultBuilderCapacity.
func NewBuilderPool() *BuilderPool {
	bp := &BuilderPool{}
	bp.pool.New = func() any {
		atomic.AddInt64(&bp.stats.Allocates, 1)
		b := &strings.Builder{}
		b.Grow(DefaultBuilderCapacity)
		return b
	}
	return bp
}

// Get retrieves a strings.Builder from the pool.
// If the pool is empty, a new builder is created with the default capacity.
//
// The capacityHint parameter provides a hint about the expected size of the
// resulting string. If greater than 0, the builder will be grown to accommodate
// at least this many bytes, reducing reallocations during string building.
//
// The caller must call Put() to return the builder to the pool when done.
// The builder is reset before being returned, but retains its allocated capacity
// for efficient reuse.
func (bp *BuilderPool) Get(capacityHint int) *strings.Builder {
	atomic.AddInt64(&bp.stats.Gets, 1)

	b := bp.pool.Get().(*strings.Builder)
	b.Reset()

	// Grow the builder if the capacity hint exceeds current capacity
	// This reduces reallocations during string building
	if capacityHint > b.Cap() {
		b.Grow(capacityHint - b.Cap())
	}

	return b
}

// Put returns a strings.Builder to the pool for reuse.
// The builder is reset before being stored.
//
// Builders with capacity exceeding MaxBuilderCapacity are not pooled
// to prevent memory bloat from retaining very large buffers. These
// are left for the garbage collector to reclaim.
func (bp *BuilderPool) Put(b *strings.Builder) {
	if b == nil {
		return
	}

	atomic.AddInt64(&bp.stats.Puts, 1)

	// Don't pool very large builders to prevent memory bloat
	if b.Cap() > MaxBuilderCapacity {
		return
	}

	bp.pool.Put(b)
}

// Stats returns the current pool statistics.
func (bp *BuilderPool) Stats() PoolStats {
	return PoolStats{
		Gets:      atomic.LoadInt64(&bp.stats.Gets),
		Puts:      atomic.LoadInt64(&bp.stats.Puts),
		Allocates: atomic.LoadInt64(&bp.stats.Allocates),
	}
}

// ResetStats resets all pool statistics to zero.
func (bp *BuilderPool) ResetStats() {
	atomic.StoreInt64(&bp.stats.Gets, 0)
	atomic.StoreInt64(&bp.stats.Puts, 0)
	atomic.StoreInt64(&bp.stats.Allocates, 0)
}

// PoolStats is a snapshot of pool statistics at a point in time.
type PoolStats struct {
	// Gets is the total number of times Get() was called.
	Gets int64
	// Puts is the total number of times Put() was called.
	Puts int64
	// Allocates is the total number of new builders created by the pool.
	Allocates int64
}

// HitRate returns the ratio of pool hits to total gets as a percentage (0.0 to 1.0).
// A higher hit rate indicates better pool utilization.
func (s PoolStats) HitRate() float64 {
	total := s.Gets
	if total == 0 {
		return 0.0
	}
	// Hits = Gets - Allocates (allocates are cache misses that had to create new objects)
	hits := total - s.Allocates
	return float64(hits) / float64(total)
}
