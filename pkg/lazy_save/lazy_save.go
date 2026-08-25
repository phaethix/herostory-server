// Package lazysave coalesces rapid writes (HP ticks) into one flush
// after a quiet period. Persist funcs must not block — hand I/O to a
// worker pool. A crash during the quiet window can lose the last
// write; call FlushAll on shutdown.
package lazysave

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// DefaultQuietPeriod is how long an id must be untouched before it is flushed.
const DefaultQuietPeriod = 20 * time.Second

// DefaultTickInterval is how often the flusher scans for stale records.
const DefaultTickInterval = 1 * time.Second

// store is isolated so tests do not share the process-wide default.
type store struct {
	quietPeriod  time.Duration
	tickInterval time.Duration

	mu      sync.Mutex
	records map[string]*record

	startOnce sync.Once
}

type record struct {
	persist    func()
	lastUpdate time.Time // guarded by store.mu; no atomics needed
}

// newStore treats non-positive durations as the package defaults.
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

// cancel drops a pending write without running it, so a newer
// synchronous persist (disconnect) is not overwritten later.
func (s *store) cancel(id string) bool {
	if id == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.records[id]; !ok {
		return false
	}
	delete(s.records, id)
	return true
}

func (s *store) flushAll() {
	s.mu.Lock()
	pending := s.records
	s.records = make(map[string]*record)
	s.mu.Unlock()

	for _, r := range pending {
		r.persist()
	}
}

func (s *store) run() {
	ticker := time.NewTicker(s.tickInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.flushStale()
	}
}

// flushStale snapshots under the lock, then persists outside it so a
// concurrent saveOrUpdate can re-register instead of racing the write.
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
		log.Debug().Int("count", len(ready)).Msg("lazy_save flushed stale records")
	}
}

var defaultStore = newStore(DefaultQuietPeriod, DefaultTickInterval)

// SaveOrUpdate's persist must not block the flusher goroutine.
func SaveOrUpdate(id string, persist func()) {
	defaultStore.saveOrUpdate(id, persist)
}

// FlushAll runs every pending persist and clears the store. Call on shutdown.
func FlushAll() {
	defaultStore.flushAll()
}

// Cancel drops a pending record without running it.
func Cancel(id string) bool {
	return defaultStore.cancel(id)
}
