package ratelimit

import (
	"math/rand"
	"sync"
	"time"
)

// EveryN is a strategy that runs the function every Nth time it is called.
// It is thread-safe.
type EveryN struct {
	mu sync.Mutex
	// N is the frequency. The function runs on the 1st call and every Nth call thereafter.
	N         int
	remaining int
}

// Handle executes the function based on the EveryN frequency.
func (e *EveryN) Handle(fn func()) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.remaining <= 0 {
		e.remaining = e.N
		fn()
	}
	e.remaining--
}

// Interval is a strategy that runs the function at most N times within a given time limit.
// It is thread-safe.
type Interval struct {
	mu sync.Mutex
	// N is the maximum number of times to run the function within the duration.
	N int
	// Duration is the time limit for the N executions.
	Duration time.Duration

	startTime time.Time
	remaining int
}

// Handle executes the function if the current time window hasn't exceeded the N-execution limit.
func (i *Interval) Handle(fn func()) {
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()

	if i.startTime.IsZero() || now.Sub(i.startTime) >= i.Duration {
		i.startTime = now
		i.remaining = i.N
	}

	if i.remaining > 0 {
		fn()
		i.remaining--
	}
}

// Random is a strategy that runs the function on average 1 out of N times.
// It is thread-safe.
type Random struct {
	N int
}

// Handle executes the function based on a uniform random distribution.
func (r *Random) Handle(fn func()) {
	// math/rand's global source is thread-safe.
	if 1+rand.Intn(r.N) == r.N {
		fn()
	}
}

// Quota is a strategy that runs the function until a specific quota is exhausted.
// It is thread-safe.
type Quota struct {
	mu sync.Mutex
	// Remaining is the number of times the function is still allowed to run.
	Remaining int
}

// Handle executes the function only if there is still quota remaining.
func (q *Quota) Handle(fn func()) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.Remaining > 0 {
		fn()
		q.Remaining--
	}
}
