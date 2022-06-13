package ratelimit

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jellydator/ttlcache/v3"
	"golang.org/x/time/rate"
)

const (
	DefaultInterval = time.Second
	DefaultTTL      = time.Minute
	DefaultBurst    = 10

	DefaultCacheKeyPrefix = "RATELIMIT:"
)

func DefaultKeyFunc(c *gin.Context) string {
	return fmt.Sprintf("%s%s", DefaultCacheKeyPrefix, c.ClientIP())
}

func DefaultLimiterFunc(_ *gin.Context) (*rate.Limiter, time.Duration) {
	return rate.NewLimiter(rate.Every(DefaultInterval), DefaultBurst), DefaultTTL
}

func DefaultAbortFunc(c *gin.Context) {
	c.Header("Retry-After", fmt.Sprintf("%.0f", DefaultTTL.Seconds()))
	c.AbortWithStatus(http.StatusTooManyRequests)
}

// NewRateLimiter original code: https://github.com/yangxikun/gin-limit-by-key
func NewRateLimiter(cache *ttlcache.Cache[string, any], keyFunc func(*gin.Context) string,
	limiterFunc func(*gin.Context) (*rate.Limiter, time.Duration), abortFunc func(*gin.Context),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		limiter, ttl := limiterFunc(c)

		item := cache.Get(key, ttlcache.WithDisableTouchOnHit[string, any]())
		if item == nil {
			cache.Set(key, limiter, ttl)
		} else {
			if val, ok := item.Value().(*rate.Limiter); !ok {
				cache.Set(key, limiter, ttl)
			} else if val != nil {
				limiter = val
			}
		}

		if ok := limiter.Allow(); !ok {
			abortFunc(c)

			return
		}

		c.Next()
	}
}
