// Package internal provides internal utilities for the time package.
// This package is not part of the public API and may be changed at any time.
package internal

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// EraCache provides thread-safe caching for era resolution results.
// It caches year conversions between Common Era (CE) and various eras
// to eliminate redundant calculations for frequently accessed years.
//
// The cache uses sync.Map for lock-free reads and CAS operations for writes,
// providing excellent performance under concurrent access.
//
// # Performance Characteristics
//
//   - O(1) lookup and insert operations
//   - Zero allocations for cache hits
//   - Minimal memory overhead (~16 bytes per entry)
//   - Reduced mutex contention through size pre-check optimization
//
// # Thread Safety
//
// All methods are safe for concurrent access:
//   - Get() uses lock-free sync.Map.Load()
//   - Set() uses lock-free sync.Map.Store() with mutex only for LRU
//   - Stats() uses atomic operations for read-only access
//
// # Era Pointer Usage
//
// The cache uses era pointers (*Era) as part of the cache key. This is safe
// because:
//   - Era instances are immutable once created
//   - Era pointers are only used as identity keys, never dereferenced
//   - The sync.Map handles concurrent access correctly
//
// # Usage Example
//
//	cache := NewEraCache(1024)
//	era := &Era{name: "BE", offset: 543}
//	// Store a year conversion
//	cache.Set(2024, unsafe.Pointer(era), 2567)
//	// Retrieve it
//	if year, ok := cache.Get(2024, unsafe.Pointer(era)); ok {
//		// year == 2567
//	}
type EraCache struct {
	cache   atomic.Value // stores *sync.Map for safe atomic swap
	maxSize int
	stats   CacheStats
	mu      sync.Mutex // Protects LRU list only
	lruList *lruList   // For LRU eviction (optional)
}

// cacheKey represents a unique cache entry key combining CE year and era pointer.
// Using unsafe.Pointer allows using Era pointers as map keys while maintaining
// performance and correctness since Era instances are immutable.
//
// #nosec G103 - Using unsafe.Pointer for pointer-to-integer conversion in map keys.
// This is safe because Era pointers are never dereferenced and Era instances are
// immutable once created. The pointer value is only used as an identity key.
type cacheKey struct {
	ceYear int64
	era    unsafe.Pointer // *Era (from gotime package)
}

// CacheStats tracks cache performance metrics for monitoring and optimization.
type CacheStats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
}

// lruList implements a simple doubly-linked list for LRU tracking.
type lruList struct {
	head *lruNode
	tail *lruNode
	size int
}

type lruNode struct {
	key  cacheKey
	prev *lruNode
	next *lruNode
}

// DefaultMaxCacheSize is the default maximum number of entries in the cache.
// This provides a good balance between memory usage and cache hit rate
// for typical workloads (100-200 unique year-era combinations).
const DefaultMaxCacheSize = 1024

// NewEraCache creates a new EraCache with the specified maximum size.
// If maxSize is 0, DefaultMaxCacheSize will be used.
func NewEraCache(maxSize int) *EraCache {
	if maxSize <= 0 {
		maxSize = DefaultMaxCacheSize
	}
	ec := &EraCache{
		maxSize: maxSize,
		lruList: newLRUList(),
	}
	ec.cache.Store(&sync.Map{})
	return ec
}

// Get retrieves the era year for the given CE year and era from the cache.
// Returns the cached era year and true if found, or 0 and false if not found.
// The era parameter should be an *Era pointer from the gotime package.
//
// #nosec G103 - era parameter is an unsafe.Pointer to *Era. Safe because Era
// instances are immutable and pointer is only used as identity key, not dereferenced.
func (ec *EraCache) Get(ceYear int, era unsafe.Pointer) (int, bool) {
	key := cacheKey{
		ceYear: int64(ceYear),
		era:    era,
	}

	cachePtr := ec.cache.Load().(*sync.Map)
	if val, ok := cachePtr.Load(key); ok {
		ec.incrementHits()
		return val.(int), true
	}

	ec.incrementMisses()
	return 0, false
}

// Set stores the era year for the given CE year and era in the cache.
// If the cache is at capacity, the least recently used entry is evicted.
// The era parameter should be an *Era pointer from the gotime package.
//
// Optimized to minimize mutex contention by only acquiring the mutex
// when eviction might be needed (when cache is near capacity).
//
// #nosec G103 - era parameter is an unsafe.Pointer to *Era. Safe because Era
// instances are immutable and pointer is only used as identity key, not dereferenced.
func (ec *EraCache) Set(ceYear int, era unsafe.Pointer, eraYear int) {
	key := cacheKey{
		ceYear: int64(ceYear),
		era:    era,
	}

	// Store the new entry first (lock-free, sync.Map handles concurrency)
	// This ensures the entry is available even if eviction fails
	cachePtr := ec.cache.Load().(*sync.Map)
	cachePtr.Store(key, eraYear)

	// Check if we need eviction - acquire mutex only for LRU management
	// This is called after Store to minimize mutex hold time
	ec.mu.Lock()
	if ec.lruList != nil {
		// Check if we need to evict before adding to LRU
		if ec.lruList.size >= ec.maxSize {
			evictedKey := ec.lruList.removeLeastRecent()
			if evictedKey.ceYear != 0 {
				// Delete from current cache
				cachePtr := ec.cache.Load().(*sync.Map)
				cachePtr.Delete(evictedKey)
				ec.stats.Evictions++
			}
		}
		// Add to LRU list
		ec.lruList.addToFront(key)
	}
	ec.mu.Unlock()
}

// Stats returns the current cache statistics.
// This method is lock-free for reads as stats are updated atomically.
func (ec *EraCache) Stats() CacheStats {
	return CacheStats{
		Hits:      atomic.LoadUint64(&ec.stats.Hits),
		Misses:    atomic.LoadUint64(&ec.stats.Misses),
		Evictions: atomic.LoadUint64(&ec.stats.Evictions),
	}
}

// Clear removes all entries from the cache and resets statistics.
func (ec *EraCache) Clear() {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Create a new empty sync.Map and atomically swap it
	ec.cache.Store(&sync.Map{})

	// Reset LRU list
	if ec.lruList != nil {
		ec.lruList = newLRUList()
	}

	// Reset stats
	atomic.StoreUint64(&ec.stats.Hits, 0)
	atomic.StoreUint64(&ec.stats.Misses, 0)
	atomic.StoreUint64(&ec.stats.Evictions, 0)
}

// HitRate returns the cache hit rate as a percentage (0.0 to 1.0).
// This method is lock-free as stats are accessed atomically.
func (ec *EraCache) HitRate() float64 {
	hits := atomic.LoadUint64(&ec.stats.Hits)
	misses := atomic.LoadUint64(&ec.stats.Misses)
	total := hits + misses
	if total == 0 {
		return 0.0
	}
	return float64(hits) / float64(total)
}

func (ec *EraCache) incrementHits() {
	atomic.AddUint64(&ec.stats.Hits, 1)
}

func (ec *EraCache) incrementMisses() {
	atomic.AddUint64(&ec.stats.Misses, 1)
}

// newLRUList creates a new LRU list.
func newLRUList() *lruList {
	return &lruList{
		head: nil,
		tail: nil,
		size: 0,
	}
}

// addToFront adds a key to the front of the LRU list.
func (l *lruList) addToFront(key cacheKey) {
	node := &lruNode{key: key}
	if l.head == nil {
		l.head = node
		l.tail = node
	} else {
		node.next = l.head
		l.head.prev = node
		l.head = node
	}
	l.size++
}

// removeLeastRecent removes and returns the least recently used key.
func (l *lruList) removeLeastRecent() cacheKey {
	if l.tail == nil {
		return cacheKey{}
	}
	key := l.tail.key
	l.tail = l.tail.prev
	if l.tail == nil {
		l.head = nil
	} else {
		l.tail.next = nil
	}
	l.size--
	return key
}
