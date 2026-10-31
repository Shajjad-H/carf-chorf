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
