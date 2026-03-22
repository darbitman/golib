package async

import "sync"

// BufferQueue is a thread-safe, generic slice-backed queue. It is optimized
// for scenarios where a producer rapidly pushes items one at a time, and a
// consumer periodically pops all available items in a single batch.
//
// To minimize allocations, the internal slice capacity is heuristically preserved 
// across PopAll calls to maintain blazing-fast, allocation-free Push() speeds
// while cleanly passing completely isolated heap slices back to the consumer.
type BufferQueue[T any] struct {
	mu     sync.Mutex
	active []T
}

// NewBufferQueue creates a new empty BufferQueue.
func NewBufferQueue[T any]() *BufferQueue[T] {
	return &BufferQueue[T]{}
}

// Push adds a single item to the queue in a thread-safe manner.
func (b *BufferQueue[T]) Push(item T) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.active = append(b.active, item)
}

// PopAll removes and returns all items currently in the queue.
// If the queue is empty, it returns nil.
//
// The returned slice is passed entirely to the caller, meaning the caller
// assumes ownership and may safely modify it. The internal queue is atomically
// swapped with a fresh array holding the identical capacity to optimize
// rapid subsequent Push() memory allocations.
func (b *BufferQueue[T]) PopAll() []T {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.active) == 0 {
		return nil
	}

	// Swap out the active slice entirely to exit the critical section instantly.
	// This delegates all O(N) cleanup natively to the Go Garbage Collector!
	res := b.active

	// Re-allocate a fresh slice. We intentionally preserve the peak capacity
	// of the old slice to prevent multiple dynamic array re-allocations
	// during the next burst of rapid Push() calls.
	newCap := cap(res)

	// If the queue was heavily under-utilized during this batch
	// (less than 25% of the total capacity was filled before PopAll was called)
	if len(res) > 0 && len(res)*4 < newCap {
		// Actively shrink the next allocation's memory footprint by half to gracefully
		// discard hoarded RAM if the queue handles "Black Swan" traffic spikes.
		newCap /= 2
	}

	b.active = make([]T, 0, newCap)

	return res
}
