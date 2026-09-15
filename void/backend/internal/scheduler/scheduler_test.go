package scheduler

import (
	"testing"
	"time"
)

func TestOperatingHoursOpenClosed(t *testing.T) {
	// Monday 09:00-17:00 UTC, evaluated at a known Monday noon and a known Sunday.
	oh := NewOperatingHours("UTC", []Window{
		{Weekday: time.Monday, StartHour: 9, StartMin: 0, EndHour: 17, EndMin: 0},
	})

	monday := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC) // a Monday
	status := oh.Evaluate(monday)
	if !status.Open {
		t.Fatalf("expected open on Monday noon, got closed")
	}
	if status.TimeUntilNext <= 0 {
		t.Fatalf("expected positive time until close")
	}

	sunday := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC) // a Sunday
	status = oh.Evaluate(sunday)
	if status.Open {
		t.Fatalf("expected closed on Sunday, got open")
	}
	if status.TimeUntilNext <= 0 {
		t.Fatalf("expected positive time until next open window")
	}
}

func TestJobRetryOnFailure(t *testing.T) {
	s := New()
	s.Enqueue(&Job{ID: "j1", RunAt: time.Now().Add(-time.Second), MaxRetries: 2})
	j := s.Next(time.Now())
	if j == nil || j.ID != "j1" {
		t.Fatal("expected job j1 to be due")
	}
	s.Complete("j1", errBoom)
	jobs := s.List()
	if jobs[0].Status != JobRetrying {
		t.Fatalf("expected retrying status, got %v", jobs[0].Status)
	}
}

type boomErr string

func (e boomErr) Error() string { return string(e) }

var errBoom = boomErr("boom")
