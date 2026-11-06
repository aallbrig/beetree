package api

import (
	"strconv"
	"sync"
	"time"
)

// Default per-user write limit: DefaultWriteLimit writes per DefaultWriteWindow.
const (
	DefaultWriteLimit  = 60
	DefaultWriteWindow = time.Minute
)

// writeLimiter is a fixed-window counter per caller. State is per process, so
// on a scaled-out Lambda the limit is best effort per instance; the API Gateway
// throttle remains the global backstop.
type writeLimiter struct {
	mu      sync.Mutex
	windows map[string]*window
}

type window struct {
	start time.Time
	count int
}

// allow records one write for key. When the limit is spent it returns false and
// the whole seconds until the window resets.
func (l *writeLimiter) allow(key string, now time.Time, limit int, span time.Duration) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.windows == nil {
		l.windows = map[string]*window{}
	}
	w := l.windows[key]
	if w == nil || !now.Before(w.start.Add(span)) {
		if len(l.windows) > 10000 {
			for k, v := range l.windows {
				if !now.Before(v.start.Add(span)) {
					delete(l.windows, k)
				}
			}
		}
		w = &window{start: now}
		l.windows[key] = w
	}
	if w.count >= limit {
		wait := int((w.start.Add(span).Sub(now) + time.Second - 1) / time.Second)
		return false, wait
	}
	w.count++
	return true, 0
}

func retryAfter(secs int) string { return strconv.Itoa(secs) }
