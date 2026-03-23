package asyncop

import (
	"hash/crc32"

	"herostory-server/pkg/main_thread"
)

// Default pool configuration.
const (
	defaultPoolSize  = 2048 // number of worker goroutines
	defaultQueueSize = 2048 // buffered channel size per worker
)

// WorkerPool dispatches async tasks across a fixed set of goroutines.
// Tasks sharing the same bind key are always routed to the same goroutine,
// preserving per-key FIFO ordering.
type WorkerPool struct {
	queues []chan func()
}

// NewWorkerPool creates a pool of n goroutines, each backed by a buffered channel.
// If n <= 0, defaultPoolSize is used.
func NewWorkerPool(n int) *WorkerPool {
	if n <= 0 {
		n = defaultPoolSize
	}

	p := &WorkerPool{
		queues: make([]chan func(), n),
	}

	for i := range p.queues {
		p.queues[i] = make(chan func(), defaultQueueSize)
		go consume(p.queues[i])
	}

	return p
}

// consume drains tasks from a channel until it is closed.
func consume(q <-chan func()) {
	for fn := range q {
		fn()
	}
}

// Process dispatches asyncOp to the goroutine selected by bindID.
// If continueWith is non-nil, it will be scheduled on the main thread
// after asyncOp completes.
func (p *WorkerPool) Process(bindID int, asyncOp func(), continueWith func()) {
	if asyncOp == nil {
		return
	}

	idx := (bindID & 0x7FFFFFFF) % len(p.queues)

	p.queues[idx] <- func() {
		asyncOp()
		if continueWith != nil {
			main_thread.Process(continueWith)
		}
	}
}


var defaultPool = NewWorkerPool(defaultPoolSize)

// Process dispatches asyncOp via the default global WorkerPool.
func Process(bindID int, asyncOp func(), continueWith func()) {
	defaultPool.Process(bindID, asyncOp, continueWith)
}

// StrToBindID converts a string key to a non-negative bind ID using CRC32,
// suitable for worker pool routing.
func StrToBindID(s string) int {
	return int(crc32.ChecksumIEEE([]byte(s)) & 0x7FFFFFFF)
}