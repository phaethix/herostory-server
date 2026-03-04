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

func execute() {
	for task := range queue {
		if task != nil {
			task()
		}
	}
}
