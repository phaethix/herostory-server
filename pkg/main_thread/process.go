package main_thread

import (
	"sync"
)

const MainQueueSize = 2048

var (
	queue = make(chan func(), MainQueueSize)
	once  sync.Once
)

func Process(task func()) {
	if task == nil {
		return
	}

	once.Do(func() { go execute() })

	queue <- task
}

// ProcessWait runs task on the main thread and blocks until it finishes.
// Must not be called from the main thread itself or it will deadlock.
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
