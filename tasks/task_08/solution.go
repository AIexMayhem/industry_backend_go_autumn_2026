package main

import (
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu          sync.Mutex
	clock       Clock
	rate        float64
	burst       int
	spent       int
	last        time.Time
	refillStart time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	limiter := &Limiter{
		clock: clock,
		rate:  ratePerSec,
		burst: burst,
	}
	if clock != nil && burst > 0 {
		limiter.last = clock.Now()
		limiter.refillStart = limiter.last
	}
	return limiter
}

func (l *Limiter) AllowN(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.clock == nil || l.burst <= 0 || n <= 0 || n > l.burst {
		return false
	}
	now := l.clock.Now()
	if now.After(l.last) {
		l.last = now
	}
	tokens := float64(l.burst - l.spent)
	if l.rate > 0 {
		tokens += float64(l.last.Sub(l.refillStart)) * l.rate / float64(time.Second)
	}
	if tokens >= float64(l.burst) {
		tokens = float64(l.burst)
		l.spent = 0
		l.refillStart = l.last
	}
	if tokens < float64(n) {
		return false
	}
	l.spent += n
	return true
}
