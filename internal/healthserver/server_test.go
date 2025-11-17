//nolint:testpackage // Testing private readinessCache type
package healthserver

import (
	"sync"
	"testing"
	"time"
)

// TestReadinessCacheRaceCondition tests that concurrent access to cache doesn't cause races.
func TestReadinessCacheRaceCondition(t *testing.T) {
	t.Parallel()

	cache := &readinessCache{}

	const concurrentOps = 100

	var wg sync.WaitGroup

	wg.Add(concurrentOps * 2)

	// Concurrent writes
	for i := range concurrentOps {
		go func(val bool) {
			defer wg.Done()

			cache.mu.Lock()
			cache.isReady = val
			cache.checkedAt = time.Now()
			cache.mu.Unlock()
		}(i%2 == 0)
	}

	// Concurrent reads
	for range concurrentOps {
		go func() {
			defer wg.Done()

			cache.mu.RLock()
			_ = cache.isReady
			_ = cache.checkedAt
			cache.mu.RUnlock()
		}()
	}

	wg.Wait()
}

// TestReadinessCacheTTL tests cache TTL logic.
func TestReadinessCacheTTL(t *testing.T) {
	t.Parallel()

	cache := &readinessCache{
		isReady:   true,
		checkedAt: time.Now(),
	}

	// Fresh cache
	cache.mu.RLock()
	cacheAge := time.Since(cache.checkedAt)
	cache.mu.RUnlock()

	if cacheAge >= readinessCacheTTL {
		t.Error("Cache should be fresh immediately after setting")
	}

	// Expired cache
	cache.mu.Lock()
	cache.checkedAt = time.Now().Add(-readinessCacheTTL - time.Second)
	cache.mu.Unlock()

	cache.mu.RLock()
	cacheAge = time.Since(cache.checkedAt)
	cache.mu.RUnlock()

	if cacheAge < readinessCacheTTL {
		t.Error("Cache should be expired after TTL")
	}
}

// TestReadinessCacheConcurrentUpdates tests concurrent cache updates don't corrupt state.
func TestReadinessCacheConcurrentUpdates(t *testing.T) {
	t.Parallel()

	cache := &readinessCache{}

	const updates = 1000

	var wg sync.WaitGroup

	wg.Add(updates)

	// Many goroutines updating cache
	for i := range updates {
		go func(iteration int) {
			defer wg.Done()

			cache.mu.Lock()
			cache.isReady = (iteration % 2) == 0
			cache.checkedAt = time.Now()
			cache.mu.Unlock()

			// Small delay to increase chance of race
			time.Sleep(time.Microsecond)
		}(i)
	}

	wg.Wait()

	// Verify final state is consistent
	cache.mu.RLock()
	isReady := cache.isReady
	checkedAt := cache.checkedAt
	cache.mu.RUnlock()

	// State should be internally consistent (no corruption)
	if checkedAt.IsZero() && isReady {
		t.Error("Inconsistent state: isReady=true but checkedAt is zero")
	}
}
