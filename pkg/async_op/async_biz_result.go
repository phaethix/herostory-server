package asyncop

import (
	"herostory-server/pkg/main_thread"
	"sync/atomic"
)

// AsyncBizResult is a one-shot future whose callback always runs on the
// main goroutine, so game-state mutations in OnComplete stay single-threaded.
type AsyncBizResult[T any] struct {
	returnedObj        atomic.Pointer[T]
	completeFunc       atomic.Pointer[func()]
	hasResult          atomic.Bool
	completeFuncCalled atomic.Bool
}

// GetReturnedObj returns the worker's result, or nil if it has not been set.
func (r *AsyncBizResult[T]) GetReturnedObj() *T {
	return r.returnedObj.Load()
}

// SetReturnedObj publishes val once. Later calls are ignored so a
// retried worker cannot overwrite a result the main thread already saw.
func (r *AsyncBizResult[T]) SetReturnedObj(val *T) {
	if r.hasResult.CompareAndSwap(false, true) {
		r.returnedObj.Store(val)
		r.doComplete()
	}
}

// OnComplete registers fn once. If the worker already finished, fn still
// runs on the main goroutine — never inline on the caller.
func (r *AsyncBizResult[T]) OnComplete(fn func()) {
	if fn == nil {
		return
	}
	if r.completeFunc.CompareAndSwap(nil, &fn) && r.hasResult.Load() {
		r.doComplete()
	}
}

func (r *AsyncBizResult[T]) doComplete() {
	fnPtr := r.completeFunc.Load()
	if fnPtr == nil {
		return
	}
	if r.completeFuncCalled.CompareAndSwap(false, true) {
		main_thread.Process(*fnPtr)
	}
}
