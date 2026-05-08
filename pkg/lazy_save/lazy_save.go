// Package lazysave provides a delayed-save mechanism (write coalescing).
//
// Multiple SaveOrUpdate calls for the same id collapse into a single
// pending record. A background goroutine flushes records that have been
// quiet for at least the configured quiet period by invoking their
// persist function, which is expected to dispatch the actual I/O off
// this goroutine (typically onto a worker pool keyed by entity id).
//
// The intent is to throttle write amplification for hot game state such
// as HP that may change many times per second. Trade-off: state changed
// within the quiet window may be lost on crash. Use FlushAll at shutdown
// to minimise that loss window.
//
// This package favours a function-valued API (id + func) over an
// interface with virtual methods: callers can register anonymous work
// via a closure without declaring a named type for every persistable
// entity.
package lazysave

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// DefaultQuietPeriod is how long a record must be untouched before the
// flusher writes it out.
const DefaultQuietPeriod = 20 * time.Second

// DefaultTickInterval is how often the flusher scans for stale records.
const DefaultTickInterval = 1 * time.Second

// store is an isolated lazy-save instance. Most callers use the package
// default via SaveOrUpdate / FlushAll; tests allocate their own store to
// keep state local and deterministic.
type store struct {
	quietPeriod  time.Duration
	tickInterval time.Duration

	mu      sync.Mutex
	records map[string]*record

	startOnce sync.Once
}

// record tracks a pending entry. lastUpdate is only touched while the
// owning store's mu is held, so no atomics are needed.
type record struct {
	persist    func()
	lastUpdate time.Time
}

// newStore builds an independent store. Zero or negative durations fall
// back to their defaults so callers can pass time.Duration(0) to mean
// "use the default".
func newStore(quiet, tick time.Duration) *store {
	if quiet <= 0 {
		quiet = DefaultQuietPeriod
	}
	if tick <= 0 {
		tick = DefaultTickInterval
	}
	return &store{
		quietPeriod:  quiet,
		tickInterval: tick,
		records:      make(map[string]*record),
	}
}

// saveOrUpdate registers persist under id or refreshes the pending
// entry. The most recent persist closure wins, which is what callers
// want: it captures the latest snapshot of the entity state.
func (s *store) saveOrUpdate(id string, persist func()) {
	if id == "" || persist == nil {
		return
	}

	s.startOnce.Do(func() { go s.run() })

	s.mu.Lock()
	defer s.mu.Unlock()

	if r, ok := s.records[id]; ok {
		r.persist = persist
		r.lastUpdate = time.Now()
		return
	}
	s.records[id] = &record{persist: persist, lastUpdate: time.Now()}
}

// flushAll synchronously runs every pending persist func and empties
// the store. It does NOT stop the background flusher: the caller is
// expected to be terminating the process.
func (s *store) flushAll() {
	s.mu.Lock()
	pending := s.records
	s.records = make(map[string]*record)
	s.mu.Unlock()

	for _, r := range pending {
		r.persist()
	}
}

// run is the background flusher loop.
func (s *store) run() {
	ticker := time.NewTicker(s.tickInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.flushStale()
	}
}

// flushStale persists every record whose last update is older than the
// quiet period. The work is done in two phases to keep the lock short
// and to avoid racing with concurrent saveOrUpdate calls:
//
//  1. under the lock, move stale records into a local slice and delete
//     them from the map;
//  2. outside the lock, invoke each persist function.
//
// A saveOrUpdate that arrives after phase 1 but before phase 2 cannot
// lose data: it will simply re-register under the same id.
func (s *store) flushStale() {
	cutoff := time.Now().Add(-s.quietPeriod)

	var ready []*record
	s.mu.Lock()
	for id, r := range s.records {
		if r.lastUpdate.After(cutoff) {
			continue
		}
		delete(s.records, id)
		ready = append(ready, r)
	}
	s.mu.Unlock()

	for _, r := range ready {
		r.persist()
	}

	if len(ready) > 0 {
		log.Debug().
			Int("count", len(ready)).
			Msg("lazy_save flushed stale records")
	}
}

// defaultStore is the process-wide instance used by the exported API.
var defaultStore = newStore(DefaultQuietPeriod, DefaultTickInterval)

// SaveOrUpdate registers persist for delayed execution under id.
// Re-registering under the same id before the quiet period elapses
// replaces the pending persist func and resets the timer. The persist
// function MUST NOT block on I/O; dispatch the write onto a worker
// pool instead.
func SaveOrUpdate(id string, persist func()) {
	defaultStore.saveOrUpdate(id, persist)
}

// FlushAll synchronously runs every pending persist func and clears the
// default store. Intended for graceful shutdown.
func FlushAll() {
	defaultStore.flushAll()
}
