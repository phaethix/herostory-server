package main_thread

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessWait_RunsTaskBeforeReturning(t *testing.T) {
	var ran atomic.Bool

	ProcessWait(func() { ran.Store(true) })

	assert.True(t, ran.Load())
}
