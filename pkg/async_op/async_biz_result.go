package asyncop

import (
	"herostory-server/pkg/main_thread"
	"sync/atomic"
)

// AsyncBizResult holds the result of an asynchronous business operation.
// It allows the caller to register an OnComplete callback (similar to a Future),
// and once the async operation sets the returned object, the callback is
// dispatched to the main thread for execution.
//
// T is the type of the result produced by the async operation,
// eliminating the need for runtime type assertions.
type AsyncBizResult[T any] struct {
	// returnedObj is the result produced by the async operation.
	returnedObj atomic.Pointer[T]
	// completeFunc is the callback to invoke when the result is ready.
	completeFunc atomic.Pointer[func()]
	// hasResult is set to true once returnedObj has been assigned.
	hasResult atomic.Bool
	// completeFuncCalled guards doComplete to fire at most once.
	completeFuncCalled atomic.Bool
}

// GetReturnedObj returns the result object set by the async operation.
// Returns nil if the result has not been set yet.
func (r *AsyncBizResult[T]) GetReturnedObj() *T {
	return r.returnedObj.Load()
}

// SetReturnedObj stores the result and triggers the completion callback
// if it has already been registered. This method is safe to call from
// any goroutine but will only take effect on the first invocation.
func (r *AsyncBizResult[T]) SetReturnedObj(val *T) {
	if r.hasResult.CompareAndSwap(false, true) {
		r.returnedObj.Store(val)
		r.doComplete()
	}
}

// OnComplete registers a callback that will be dispatched to the main
// thread once the async result is available. If the result has already
// been set, the callback fires immediately (still on the main thread).
// Only the first call to OnComplete takes effect; subsequent calls are ignored.
func (r *AsyncBizResult[T]) OnComplete(fn func()) {
	if fn == nil {
		return
	}
	if r.completeFunc.CompareAndSwap(nil, &fn) {
		// If the returned object was already set before OnComplete was called,
		// trigger the callback now.
		if r.hasResult.Load() {
			r.doComplete()
		}
	}
}

// doComplete dispatches completeFunc to the main thread exactly once.
func (r *AsyncBizResult[T]) doComplete() {
	fnPtr := r.completeFunc.Load()
	if fnPtr == nil {
		return
	}

	// use CAS to ensure completeFunc is dispatched exactly once.
	if r.completeFuncCalled.CompareAndSwap(false, true) {
		main_thread.Process(*fnPtr)
	}
}
