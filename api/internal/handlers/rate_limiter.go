package handlers

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests     map[string][]time.Time
	mutex        sync.Mutex
	maxRequests  int
	windowLength time.Duration
}

func NewRateLimiter(maxRequests int, windowLength time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:     make(map[string][]time.Time),
		maxRequests:  maxRequests,
		windowLength: windowLength,
	}
}

func (rl *RateLimiter) IsRateLimited(clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.windowLength)

	if reqs, exists := rl.requests[clientIP]; exists {
		newReqs := []time.Time{}
		for _, req := range reqs {
			if req.After(cutoff) {
				newReqs = append(newReqs, req)
			}
		}
		rl.requests[clientIP] = newReqs
	}
	rl.requests[clientIP] = append(rl.requests[clientIP], now)
	return len(rl.requests[clientIP]) > rl.maxRequests
}

func MiddlewareRateLimit(rl *RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr

			// You can use X-Forwarded-For if behind a load balancer
			forwardedFor := r.Header.Get("X-Forwarded-For")
			if forwardedFor != "" {
				clientIP = forwardedFor
			}

			if rl.IsRateLimited(clientIP) {
				w.Header().Set("Retry-After", "60")
				RespondWithError(w, http.StatusTooManyRequests, "Rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
