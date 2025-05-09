package middleware

import (
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
)

type rateLimiter struct {
	sync.RWMutex
	requests map[string][]time.Time
	window   time.Duration
	limit    int
}

func newRateLimiter(window time.Duration, limit int) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		window:   window,
		limit:    limit,
	}

	// Start cleanup goroutine
	go func() {
		for {
			time.Sleep(window)
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *rateLimiter) cleanup() {
	rl.Lock()
	defer rl.Unlock()

	now := time.Now()
	for ip, times := range rl.requests {
		var valid []time.Time
		for _, t := range times {
			if now.Sub(t) < rl.window {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = valid
		}
	}
}

func (rl *rateLimiter) isAllowed(ip string) bool {
	rl.Lock()
	defer rl.Unlock()

	now := time.Now()
	times := rl.requests[ip]

	// Remove old requests outside the window
	var valid []time.Time
	for _, t := range times {
		if now.Sub(t) < rl.window {
			valid = append(valid, t)
		}
	}

	// Update requests
	valid = append(valid, now)
	rl.requests[ip] = valid

	return len(valid) <= rl.limit
}

// RateLimitAuth creates a middleware that limits requests to auth endpoints
func RateLimitAuth() echo.MiddlewareFunc {
	limiter := newRateLimiter(5*time.Minute, 30) // 30 requests per 5 minutes

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()

			if !limiter.isAllowed(ip) {
				return apperror.ValidationError("Too many requests", nil).
					WithDetails(&apperror.ValidationErrorDetails{
						Field:  "request",
						Reason: "rate_limit_exceeded",
					})
			}

			return next(c)
		}
	}
}
