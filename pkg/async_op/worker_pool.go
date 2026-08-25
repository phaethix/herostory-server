package asyncop

import (
	"hash/crc32"

	"herostory-server/pkg/main_thread"
)

const (
	defaultPoolSize  = 2048
	defaultQueueSize = 2048
)

// WorkerPool pins tasks with the same bind key onto one goroutine so
// per-user DB writes stay FIFO and never run in parallel.
type WorkerPool struct {
	queues []chan func()
}

// NewWorkerPool starts n workers. n <= 0 uses defaultPoolSize.
func NewWorkerPool(n int) *WorkerPool {
	if n <= 0 {
		n = defaultPoolSize
	}

	p := &WorkerPool{queues: make([]chan func(), n)}
	for i := range p.queues {
		p.queues[i] = make(chan func(), defaultQueueSize)
		go consume(p.queues[i])
	}
	return p
}

func consume(q <-chan func()) {
	for fn := range q {
		fn()
	}
}

// Process runs asyncOp on the worker selected by bindID.
// continueWith, if non-nil, is then posted to the main goroutine.
func (p *WorkerPool) Process(bindID int, asyncOp, continueWith func()) {
	if asyncOp == nil {
		return
	}

	// Mask the sign bit: Go's remainder of a negative bindID is negative,
	// which is not a valid slice index.
	idx := (bindID & 0x7FFFFFFF) % len(p.queues)
	p.queues[idx] <- func() {
		asyncOp()
		if continueWith != nil {
			main_thread.Process(continueWith)
		}
	}
}

var defaultPool = NewWorkerPool(defaultPoolSize)

// Process is Process on the process-wide pool.
func Process(bindID int, asyncOp, continueWith func()) {
	defaultPool.Process(bindID, asyncOp, continueWith)
}

// StrToBindID is stable for a given string so the same username always
// lands on the same worker.
func StrToBindID(s string) int {
	return int(crc32.ChecksumIEEE([]byte(s)) & 0x7FFFFFFF)
}
