// Package scheduler implements VOID's Scheduler: it queues, prioritizes,
// retries and tracks long-running Simulation/Experiment Jobs, and separately
// manages fully user-configurable Operating Hours windows (e.g. "only run
// heavy simulations Sat-Wed 09:00-18:00 Asia/Baku") so the dashboard can
// show, live, whether the system is inside/outside its allowed window and
// exactly how long until the next open/close transition.
package scheduler

import (
	"container/heap"
	"sync"
	"time"
)

// JobStatus enumerates scheduler job lifecycle states.
type JobStatus string

const (
	JobQueued  JobStatus = "queued"
	JobRunning JobStatus = "running"
	JobDone    JobStatus = "done"
	JobFailed  JobStatus = "failed"
	JobRetrying JobStatus = "retrying"
)

// Job is one scheduled unit of work (a Simulation run, Experiment sweep,
// Replay, Export, ...).
type Job struct {
	ID         string
	Kind       string // "simulation" | "experiment" | "replay" | "export"
	Priority   int
	RunAt      time.Time
	Status     JobStatus
	Attempts   int
	MaxRetries int
	Payload    map[string]interface{}
	Err        string

	index int
}

type jobHeap []*Job

func (h jobHeap) Len() int { return len(h) }
func (h jobHeap) Less(i, j int) bool {
	if !h[i].RunAt.Equal(h[j].RunAt) {
		return h[i].RunAt.Before(h[j].RunAt)
	}
	return h[i].Priority > h[j].Priority
}
func (h jobHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i]; h[i].index, h[j].index = i, j }
func (h *jobHeap) Push(x interface{}) {
	j := x.(*Job)
	j.index = len(*h)
	*h = append(*h, j)
}
func (h *jobHeap) Pop() interface{} {
	old := *h
	n := len(old)
	j := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return j
}

// Scheduler queues jobs and hands due ones to a worker via Next().
type Scheduler struct {
	mu   sync.Mutex
	h    jobHeap
	byID map[string]*Job
}

func New() *Scheduler {
	s := &Scheduler{byID: map[string]*Job{}}
	heap.Init(&s.h)
	return s
}

func (s *Scheduler) Enqueue(j *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j.Status = JobQueued
	heap.Push(&s.h, j)
	s.byID[j.ID] = j
}

// Next pops the earliest due job, or nil if the earliest job's RunAt is
// still in the future (caller should sleep/poll).
func (s *Scheduler) Next(now time.Time) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.h) == 0 || s.h[0].RunAt.After(now) {
		return nil
	}
	j := heap.Pop(&s.h).(*Job)
	j.Status = JobRunning
	return j
}

func (s *Scheduler) Complete(id string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.byID[id]
	if !ok {
		return
	}
	if err == nil {
		j.Status = JobDone
		return
	}
	j.Attempts++
	j.Err = err.Error()
	if j.Attempts <= j.MaxRetries {
		j.Status = JobRetrying
		j.RunAt = time.Now().Add(time.Duration(j.Attempts) * 5 * time.Second)
		heap.Push(&s.h, j)
	} else {
		j.Status = JobFailed
	}
}

func (s *Scheduler) List() []*Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Job, 0, len(s.byID))
	for _, j := range s.byID {
		out = append(out, j)
	}
	return out
}

// ---------------------------------------------------------------------
// Operating Hours: fully user-configurable schedule windows.
// ---------------------------------------------------------------------

// Window is one recurring open window, e.g. Weekday=Monday 09:00-18:00.
type Window struct {
	Weekday   time.Weekday `json:"weekday"`
	StartHour int          `json:"start_hour"` // 0-23
	StartMin  int          `json:"start_min"`  // 0-59
	EndHour   int          `json:"end_hour"`
	EndMin    int          `json:"end_min"`
}

// OperatingHours is a fully user-supplied weekly schedule (which days are
// "open", and the open/close time on each) plus the IANA timezone it's
// evaluated in. Every field is provided by the user - VOID makes no
// assumptions about business hours.
type OperatingHours struct {
	mu       sync.RWMutex
	Timezone string   `json:"timezone"`
	Windows  []Window `json:"windows"`
}

func NewOperatingHours(timezone string, windows []Window) *OperatingHours {
	return &OperatingHours{Timezone: timezone, Windows: windows}
}

func (o *OperatingHours) Set(timezone string, windows []Window) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.Timezone = timezone
	o.Windows = windows
}

func (o *OperatingHours) loc() *time.Location {
	loc, err := time.LoadLocation(o.Timezone)
	if err != nil || o.Timezone == "" {
		return time.UTC
	}
	return loc
}

// Status reports whether the schedule is currently open, and how long until
// the next transition (close time if open, next open time if closed).
type Status struct {
	Open              bool          `json:"open"`
	Now               time.Time     `json:"now"`
	NextTransition    time.Time     `json:"next_transition"`
	TimeUntilNext     time.Duration `json:"time_until_next"`
	ActiveWindowLabel string        `json:"active_window_label,omitempty"`
}

// Evaluate computes the live Status against the wall clock, entirely from
// user-supplied Windows (no hardcoded hours anywhere in VOID).
func (o *OperatingHours) Evaluate(at time.Time) Status {
	o.mu.RLock()
	defer o.mu.RUnlock()

	loc := o.loc()
	now := at.In(loc)

	if len(o.Windows) == 0 {
		return Status{Open: false, Now: now}
	}

	// Check if `now` falls inside any window today (or a window that
	// started yesterday and spans midnight).
	for _, w := range o.Windows {
		start, end := windowBounds(now, w)
		if !end.After(start) {
			end = end.Add(24 * time.Hour) // overnight window
		}
		if !now.Before(start) && now.Before(end) {
			return Status{
				Open: true, Now: now, NextTransition: end,
				TimeUntilNext:     end.Sub(now),
				ActiveWindowLabel: w.Weekday.String(),
			}
		}
	}

	// Not open now: find the soonest upcoming window start within the next 8 days.
	var next time.Time
	for d := 0; d <= 8; d++ {
		day := now.AddDate(0, 0, d)
		for _, w := range o.Windows {
			start, _ := windowBounds(day, w)
			if start.After(now) && (next.IsZero() || start.Before(next)) {
				next = start
			}
		}
		if !next.IsZero() {
			break
		}
	}
	if next.IsZero() {
		return Status{Open: false, Now: now}
	}
	return Status{Open: false, Now: now, NextTransition: next, TimeUntilNext: next.Sub(now)}
}

func windowBounds(reference time.Time, w Window) (time.Time, time.Time) {
	// Align `reference`'s date to the window's target weekday within the
	// same week (offset -6..+6 days), then apply the configured start/end time.
	offset := int(w.Weekday) - int(reference.Weekday())
	day := reference.AddDate(0, 0, offset)
	loc := reference.Location()
	start := time.Date(day.Year(), day.Month(), day.Day(), w.StartHour, w.StartMin, 0, 0, loc)
	end := time.Date(day.Year(), day.Month(), day.Day(), w.EndHour, w.EndMin, 0, 0, loc)
	return start, end
}
