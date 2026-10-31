package admin_auth

import (
	"testing"
	"time"
)

func TestFailureLimiter(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	limiter := NewFailureLimiter(5, 5*time.Minute)
	limiter.now = func() time.Time { return now }

	for i := 0; i < 5; i++ {
		if limiter.Blocked() {
			t.Fatalf("blocked before attempt %d", i+1)
		}
		limiter.Fail()
	}

	if !limiter.Blocked() {
		t.Fatal("limiter should be blocked after 5 failures")
	}
	if retry := limiter.RetryAfter(); retry <= 0 {
		t.Fatalf("retry after = %v, want > 0", retry)
	}

	now = now.Add(5*time.Minute + time.Second)
	if limiter.Blocked() {
		t.Fatal("limiter should unblock after the window")
	}

	limiter.Fail()
	limiter.Reset()
	if limiter.Blocked() {
		t.Fatal("limiter should be clear after Reset")
	}
}
