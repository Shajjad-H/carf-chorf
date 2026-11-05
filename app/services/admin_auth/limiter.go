package admin_auth

import (
	"sync"
	"time"
)

// FailureLimiter tracks failed admin login attempts across all clients.
type FailureLimiter struct {
	mu       sync.Mutex
	failures []time.Time
	max      int
	window   time.Duration
	now      func() time.Time
}

// NewFailureLimiter creates a global sliding-window failure limiter.
func NewFailureLimiter(max int, window time.Duration) *FailureLimiter {
	return &FailureLimiter{
		max:    max,
		window: window,
		now:    time.Now,
	}
}

// Blocked reports whether the limit has been reached.
func (l *FailureLimiter) Blocked() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked()
	return len(l.failures) >= l.max
}

// Fail records a failed attempt.
func (l *FailureLimiter) Fail() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked()
	l.failures = append(l.failures, l.now())
}

// Reset clears failed attempts after a successful login.
func (l *FailureLimiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.failures = nil
}

// RetryAfter returns how long until the oldest failure leaves the window.
func (l *FailureLimiter) RetryAfter() time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked()
	if len(l.failures) < l.max {
		return 0
	}
	return l.failures[0].Add(l.window).Sub(l.now())
}

func (l *FailureLimiter) pruneLocked() {
	cutoff := l.now().Add(-l.window)
	first := 0
	for first < len(l.failures) && l.failures[first].Before(cutoff) {
		first++
	}
	if first > 0 {
		l.failures = append([]time.Time(nil), l.failures[first:]...)
	}
}

// KeyedFailureLimiter tracks failures independently per key (for example, per
// client IP) so one bad actor cannot lock every administrator out.
type KeyedFailureLimiter struct {
	mu       sync.Mutex
	failures map[string][]time.Time
	max      int
	window   time.Duration
	now      func() time.Time
}

// NewKeyedFailureLimiter creates a keyed sliding-window failure limiter.
func NewKeyedFailureLimiter(max int, window time.Duration) *KeyedFailureLimiter {
	return &KeyedFailureLimiter{
		failures: make(map[string][]time.Time),
		max:      max,
		window:   window,
		now:      time.Now,
	}
}

// Blocked reports whether the key reached the limit.
func (l *KeyedFailureLimiter) Blocked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked(key)
	return len(l.failures[key]) >= l.max
}

// Fail records a failure for the key.
func (l *KeyedFailureLimiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked(key)
	l.failures[key] = append(l.failures[key], l.now())
}

// Reset clears failures for the key.
func (l *KeyedFailureLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
}

// RetryAfter returns how long until the oldest failure leaves the window.
func (l *KeyedFailureLimiter) RetryAfter(key string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked(key)
	failures := l.failures[key]
	if len(failures) < l.max {
		return 0
	}
	return failures[0].Add(l.window).Sub(l.now())
}

func (l *KeyedFailureLimiter) pruneLocked(key string) {
	cutoff := l.now().Add(-l.window)
	failures := l.failures[key]
	first := 0
	for first < len(failures) && failures[first].Before(cutoff) {
		first++
	}
	if first > 0 {
		l.failures[key] = append([]time.Time(nil), failures[first:]...)
	}
	if len(l.failures[key]) == 0 {
		delete(l.failures, key)
	}
}
