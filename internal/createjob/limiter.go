package createjob

import (
	"strings"
	"sync"
	"time"
)

// Limiter tracks successful alias creations per account per local hour.
type Limiter struct {
	mu      sync.Mutex
	limit   int
	buckets map[string]int
}

type QuotaState struct {
	Buckets map[string]int `json:"buckets,omitempty"`
}

func NewLimiter(limit int) *Limiter {
	if limit <= 0 {
		limit = 5
	}
	return &Limiter{
		limit:   limit,
		buckets: make(map[string]int),
	}
}

func (l *Limiter) Restore(state QuotaState, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.buckets = make(map[string]int, len(state.Buckets))
	for key, used := range state.Buckets {
		if used <= 0 {
			continue
		}
		hour, ok := hourFromLimitKey(key)
		if ok && hour.Before(now.Local().Truncate(time.Hour)) {
			continue
		}
		if used > l.limit {
			used = l.limit
		}
		l.buckets[key] = used
	}
}

func (l *Limiter) Snapshot(now time.Time) QuotaState {
	l.mu.Lock()
	defer l.mu.Unlock()

	currentHour := now.Local().Truncate(time.Hour)
	buckets := make(map[string]int, len(l.buckets))
	for key, used := range l.buckets {
		if used <= 0 {
			continue
		}
		hour, ok := hourFromLimitKey(key)
		if ok && hour.Before(currentHour) {
			continue
		}
		buckets[key] = used
	}
	return QuotaState{Buckets: buckets}
}

func (l *Limiter) Remaining(accountID string, at time.Time) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	used := l.buckets[limitKey(accountID, at)]
	remaining := l.limit - used
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (l *Limiter) TryReserve(accountID string, at time.Time, count int) bool {
	if count <= 0 {
		return true
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	key := limitKey(accountID, at)
	if l.buckets[key]+count > l.limit {
		return false
	}
	l.buckets[key] += count
	return true
}

func (l *Limiter) RecordSuccess(accountID string, at time.Time) bool {
	return l.TryReserve(accountID, at, 1)
}

func (l *Limiter) Release(accountID string, at time.Time, count int) {
	if count <= 0 {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	key := limitKey(accountID, at)
	l.buckets[key] -= count
	if l.buckets[key] <= 0 {
		delete(l.buckets, key)
	}
}

func limitKey(accountID string, at time.Time) string {
	return accountID + "|" + at.Local().Truncate(time.Hour).Format(time.RFC3339)
}

func hourFromLimitKey(key string) (time.Time, bool) {
	idx := strings.LastIndex(key, "|")
	if idx < 0 || idx == len(key)-1 {
		return time.Time{}, false
	}
	hour, err := time.Parse(time.RFC3339, key[idx+1:])
	if err != nil {
		return time.Time{}, false
	}
	return hour.Local().Truncate(time.Hour), true
}
