package main_thread

import (
	"sync"
)

const MainQueueSize = 2048

var (
	queue = make(chan func(), MainQueueSize)
	once  sync.Once
)

// Process enqueues task for serial execution on the main goroutine.
func Process(task func()) {
	if task == nil {
		return
	}
	once.Do(func() { go execute() })
	queue <- task
}

// ProcessWait runs task on the main goroutine and blocks until it finishes.
// It deadlocks if called from the main goroutine itself.
func ProcessWait(task func()) {
	if task == nil {
		return
	}

	done := make(chan struct{})
	Process(func() {
		task()
		close(done)
	})
	<-done
}

func execute() {
	for task := range queue {
		if task != nil {
			task()
		}
	}
}
