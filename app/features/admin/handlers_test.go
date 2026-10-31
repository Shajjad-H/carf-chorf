package admin

import (
	"testing"
	"time"
)

func TestParseDailyDate(t *testing.T) {
	now := time.Date(2026, 10, 31, 14, 30, 0, 0, time.Local)
	today := time.Date(2026, 10, 31, 0, 0, 0, 0, time.Local)

	if got := parseDailyDate("", now); !got.Equal(today) {
		t.Fatalf("empty date = %v, want %v", got, today)
	}
	if got := parseDailyDate("invalid", now); !got.Equal(today) {
		t.Fatalf("invalid date = %v, want %v", got, today)
	}
	if got := parseDailyDate("2026-11-01", now); !got.Equal(today) {
		t.Fatalf("future date = %v, want %v", got, today)
	}

	want := time.Date(2026, 10, 29, 0, 0, 0, 0, time.Local)
	if got := parseDailyDate("2026-10-29", now); !got.Equal(want) {
		t.Fatalf("past date = %v, want %v", got, want)
	}
}
