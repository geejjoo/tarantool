package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"

	"kv-storage/internal/interfaces"
)

type RateLimiter struct {
	requests map[string]*TokenBucket
	mu       sync.RWMutex
	rate     int
	burst    int
	logger   interfaces.Logger
}

type TokenBucket struct {
	tokens     int
	lastRefill time.Time
	rate       int
	burst      int
	mu         sync.Mutex
}

func NewRateLimiter(rate, burst int, logger interfaces.Logger) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*TokenBucket),
		rate:     rate,
		burst:    burst,
		logger:   logger,
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := getClientIP(c.Request)

		if !rl.allowRequest(clientIP) {
			rl.logger.Warn("Rate limit exceeded", "client_ip", clientIP)
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) allowRequest(clientIP string) bool {
	rl.mu.Lock()
	bucket, exists := rl.requests[clientIP]
	if !exists {
		bucket = &TokenBucket{
			tokens:     rl.burst,
			lastRefill: time.Now(),
			rate:       rl.rate,
			burst:      rl.burst,
		}
		rl.requests[clientIP] = bucket
	}
	rl.mu.Unlock()

	return bucket.takeToken()
}

func (tb *TokenBucket) takeToken() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed.Seconds() * float64(tb.rate))

	if tokensToAdd > 0 {
		tb.tokens = min(tb.burst, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
