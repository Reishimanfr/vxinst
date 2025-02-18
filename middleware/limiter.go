/*
VxInstagram - Blazing fast embedder for instagram posts
Copyright (C) 2025 Bash06

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package middleware

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
