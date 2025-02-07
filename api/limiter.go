package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	rate   int
	bucket int
	tokens int
	mutex  sync.Mutex
	ticker *time.Ticker
}

func NewRateLimiter(rate, bucket int) *RateLimiter {
	limiter := &RateLimiter{
		rate:   rate,
		bucket: bucket,
		tokens: bucket,
		ticker: time.NewTicker(time.Second / time.Duration(rate)),
	}

	go limiter.refillTokens()

	return limiter
}

func (r *RateLimiter) refillTokens() {
	for range r.ticker.C {
		r.mutex.Lock()

		if r.tokens < r.bucket {
			r.tokens++
		}

		r.mutex.Unlock()
	}
}

func (r *RateLimiter) Allow() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.tokens > 0 {
		r.tokens--
		return true
	}

	return false
}

func RateLimiterMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many request",
			})
			return
		}
		c.Next()
	}
}
