package async

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TaskFunc defines the signature for an asynchronous task. It accepts a
// context that may be cancelled or time out, and yields a generic result
// along with any execution error.
type TaskFunc[T any] func(context.Context) (T, error)

// Future represents a thread-safe handle to an active asynchronous task.
// It allows the caller to await the completion of the background goroutine,
// intercept timeouts, or explicitly cancel the pending chore.
type Future[T any] struct {
	respch     chan response[T]
	ctx        context.Context
	cancelFunc context.CancelFunc

	once   sync.Once
	result response[T]
}

// response wraps the output or error of an asynchronous task for safe
// transport across channels.
type response[T any] struct {
	value T
	err   error
}

// Submit begins executing the given task asynchronously in a newly spawned
// goroutine using the provided context. It returns a Future[T]
// which can be used to Await the result or explicitly Cancel the operation.
func Submit[T any](ctx context.Context, task TaskFunc[T]) *Future[T] {
	ctx, cancel := context.WithCancel(ctx)
	return submit(ctx, cancel, task)
}

// SubmitWithTimeout begins executing the given task asynchronously,
// wrapping the provided context with the specified timeout duration.
// If the task surpasses this timeout, it will be automatically cancelled.
func SubmitWithTimeout[T any](ctx context.Context, task TaskFunc[T], d time.Duration) *Future[T] {
	ctx, cancel := context.WithTimeout(ctx, d)
	return submit(ctx, cancel, task)
}

// Await blocks the calling goroutine until the background task successfully
// completes, returns an error, or the context is cancelled/timed out.
// It is perfectly safe and idempotent to call Await multiple times concurrently;
// all callers will correctly receive the cached identical result.
func (f *Future[T]) Await() (T, error) {
	f.once.Do(func() {
		defer f.cancelFunc()

		select {
		case resp := <-f.respch:
			f.result = resp
		case <-f.ctx.Done():
			// Double-check `respch` just in case the task successfully completed
			// and triggered `defer cancel()` exactly when we entered this select.
			select {
			case resp := <-f.respch:
				f.result = resp
			default:
				f.result = response[T]{err: f.ctx.Err()}
			}
		}
	})
	return f.result.value, f.result.err
}

// Cancel immediately aborts the underlying context for the asynchronous task.
// Any active or future calls to Await will return the context cancellation error.
func (f *Future[T]) Cancel() {
	f.cancelFunc()
}

// submit initializes the task execution pipeline, wiring up standard internal
// contexts, cancellation propagation, and spinning up the internal goroutine.
func submit[T any](ctx context.Context, cancel context.CancelFunc, task TaskFunc[T]) *Future[T] {
	respch := make(chan response[T], 1)

	go func() {
		defer cancel()

		var val T
		var err error

		// Safely recover from panics inside the async task so it doesn't crash the host application
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("async task panicked: %v", r)
			}
			respch <- response[T]{
				value: val,
				err:   err,
			}
			close(respch)
		}()

		val, err = task(ctx)
	}()

	return &Future[T]{
		respch:     respch,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}
