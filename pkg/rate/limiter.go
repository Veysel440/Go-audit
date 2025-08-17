package rate

import (
	"net"
	"sync"
	"time"
)

type bucket struct {
	tokens int
	last   time.Time
}
type Limiter struct {
	mu      sync.Mutex
	rate    int
	window  time.Duration
	buckets map[string]*bucket
}

func New(rate int, window time.Duration) *Limiter {
	return &Limiter{rate: rate, window: window, buckets: map[string]*bucket{}}
}

func (l *Limiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b, ok := l.buckets[ip]
	if !ok || now.Sub(b.last) > l.window {
		l.buckets[ip] = &bucket{tokens: l.rate - 1, last: now}
		return true
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	b.last = now
	return true
}

func IP(raddr string) string {
	host, _, err := net.SplitHostPort(raddr)
	if err != nil {
		return raddr
	}
	return host
}
