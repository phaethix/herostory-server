package lazysave

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// counter returns a persist func that increments an atomic counter on
// every call, and a pointer to the counter.
func counter() (func(), *atomic.Int32) {
	var n atomic.Int32
	return func() { n.Add(1) }, &n
}

func TestSaveOrUpdate_CoalescesById(t *testing.T) {
	s := newStore(time.Hour, time.Hour) // disable auto-flush

	persist, _ := counter()
	for range 100 {
		s.saveOrUpdate("user:1", persist)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	assert.Equal(t, 1, len(s.records), "saves for the same id must coalesce")
}

func TestSaveOrUpdate_LatestPersistWins(t *testing.T) {
	// Re-registering under the same id before the flush must replace
	// the stored persist func: it normally captures the newest state
	// snapshot, so the older closure is stale.
	s := newStore(time.Hour, time.Hour)

	oldPersist, oldCount := counter()
	newPersist, newCount := counter()

	s.saveOrUpdate("user:1", oldPersist)
	s.saveOrUpdate("user:1", newPersist)
	s.flushAll()

	assert.EqualValues(t, 0, oldCount.Load(), "stale persist should not run")
	assert.EqualValues(t, 1, newCount.Load(), "latest persist should run exactly once")
}

func TestFlushStale_SkipsRecentlyTouched(t *testing.T) {
	s := newStore(50*time.Millisecond, time.Hour)

	persist, count := counter()
	s.saveOrUpdate("user:1", persist)

	s.flushStale() // still fresh
	assert.EqualValues(t, 0, count.Load(), "persist must not run before quiet period")

	time.Sleep(80 * time.Millisecond)
	s.flushStale() // now stale
	assert.EqualValues(t, 1, count.Load(), "persist must run after quiet period")

	s.mu.Lock()
	defer s.mu.Unlock()
	assert.Empty(t, s.records, "flushed record must be removed")
}

func TestFlushStale_DoesNotLoseConcurrentRefresh(t *testing.T) {
	// A SaveOrUpdate that arrives after a flush must re-register
	// cleanly; otherwise updates could be silently dropped.
	s := newStore(time.Nanosecond, time.Hour) // everything is immediately stale

	persist, count := counter()
	s.saveOrUpdate("user:1", persist)
	s.flushStale()
	require.EqualValues(t, 1, count.Load())

	s.saveOrUpdate("user:1", persist)

	s.mu.Lock()
	_, present := s.records["user:1"]
	s.mu.Unlock()
	assert.True(t, present, "post-flush saveOrUpdate must re-register")
}

func TestFlushAll_PersistsAndClears(t *testing.T) {
	s := newStore(time.Hour, time.Hour)

	const n = 16
	counts := make([]*atomic.Int32, n)
	for i := range n {
		persist, c := counter()
		counts[i] = c
		s.saveOrUpdate("u"+strconv.Itoa(i), persist)
	}

	s.flushAll()

	for i, c := range counts {
		assert.EqualValuesf(t, 1, c.Load(), "obj %d: expected exactly one persist", i)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	assert.Empty(t, s.records)
}

func TestSaveOrUpdate_Concurrent(t *testing.T) {
	// Under -race this catches any unguarded access to the records map
	// or the lastUpdate field.
	s := newStore(time.Millisecond, time.Hour)

	persist, _ := counter()

	const goroutines = 32
	const perG = 1000

	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range perG {
				s.saveOrUpdate("user:hot", persist)
			}
		})
	}

	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				s.flushStale()
			}
		}
	}()

	wg.Wait()
	close(stop)
	// No assertion: the goal is to prove race-freedom, not to pin
	// down a specific persist count (non-deterministic by design).
}

func TestNewStore_AppliesDefaultsForNonPositive(t *testing.T) {
	s := newStore(0, -1)
	assert.Equal(t, DefaultQuietPeriod, s.quietPeriod)
	assert.Equal(t, DefaultTickInterval, s.tickInterval)
}
