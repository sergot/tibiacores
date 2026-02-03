package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// RateLimiterConfig defines the configuration for rate limiting
type RateLimiterConfig struct {
	// RequestsPerSecond is the number of requests allowed per second
	RequestsPerSecond float64
	// Burst is the maximum burst size
	Burst int
}

// IPRateLimiter manages rate limiters per IP address
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      float64
	burst    int
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(rps float64, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
	}
}

// GetLimiter returns the rate limiter for a given IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(i.rps), i.burst)
		i.limiters[ip] = limiter
	}

	return limiter
}

// RateLimiterMiddleware creates a rate limiting middleware
func RateLimiterMiddleware(limiter *IPRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			if ip == "" {
				ip = c.Request().RemoteAddr
			}

			l := limiter.GetLimiter(ip)
			if !l.Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded")
			}

			return next(c)
		}
	}
}

// CleanupOldLimiters periodically cleans up old rate limiters to prevent memory leaks
func (i *IPRateLimiter) CleanupOldLimiters(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		// In a production environment, you would track last access times
		// and remove limiters that haven't been used recently
		// For now, we clear all limiters periodically
		if len(i.limiters) > 10000 {
			i.limiters = make(map[string]*rate.Limiter)
		}
		i.mu.Unlock()
	}
}
