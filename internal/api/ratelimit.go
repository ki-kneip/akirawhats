package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	capacity float64
	refill   float64 // tokens per second
}

func newRateLimiter(capacity int, perMinute int) *rateLimiter {
	return &rateLimiter{
		buckets:  make(map[string]*bucket),
		capacity: float64(capacity),
		refill:   float64(perMinute) / 60.0,
	}
}

func (r *rateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	b, ok := r.buckets[key]
	if !ok {
		r.buckets[key] = &bucket{tokens: r.capacity - 1, lastRefill: now}
		return true
	}
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * r.refill
	if b.tokens > r.capacity {
		b.tokens = r.capacity
	}
	b.lastRefill = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// gc removes buckets that have refilled to capacity (idle).
func (r *rateLimiter) gc() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for k, b := range r.buckets {
		elapsed := now.Sub(b.lastRefill).Seconds()
		if b.tokens+elapsed*r.refill >= r.capacity {
			delete(r.buckets, k)
		}
	}
}

func rateLimitMiddleware(rl *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}

func startRateLimitGC(rls ...*rateLimiter) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			for _, rl := range rls {
				rl.gc()
			}
		}
	}()
}
