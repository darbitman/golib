package ratelimit

import (
	"errors"
	"sync"
)

var (
	// ErrKeyNotRegistered is returned by Execute when the provided key has not been pre-registered.
	ErrKeyNotRegistered = errors.New("ratelimit: key not registered")

	// ErrStrategyNil is returned by Register when a nil strategy is provided.
	ErrStrategyNil = errors.New("ratelimit: nil strategy given")

	// ErrKeyAlreadyRegistered is returned by Register when the provided key is already in use.
	ErrKeyAlreadyRegistered = errors.New("ratelimit: key already registered")
)

// Strategy is the interface for different rate-limiting algorithms.
type Strategy interface {
	Handle(fn func())
}

// Limiter manages a collection of named rate-limiting strategies.
// It is the main entry point for the ratelimit package.
type Limiter struct {
	store sync.Map
}

// New creates and returns a new Limiter instance.
func New() *Limiter {
	return &Limiter{}
}

// Register pre-registers a named strategy that can be used later by calling Execute.
// If the strategy is nil, it returns ErrStrategyNil. If the key is already
// registered, it returns ErrKeyAlreadyRegistered.
func (l *Limiter) Register(key string, strategy Strategy) error {
	if strategy == nil {
		return ErrStrategyNil
	}
	_, loaded := l.store.LoadOrStore(key, strategy)
	if loaded {
		return ErrKeyAlreadyRegistered
	}
	return nil
}

// Execute looks up a pre-registered strategy by key and runs the function accordingly.
// If no strategy is found for the given key, it returns ErrKeyNotRegistered.
func (l *Limiter) Execute(key string, fn func()) error {
	if fn == nil {
		return nil
	}

	actual, ok := l.store.Load(key)
	if !ok {
		return ErrKeyNotRegistered
	}

	s := actual.(Strategy)
	s.Handle(fn)
	return nil
}
