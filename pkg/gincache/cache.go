package gincache

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jellydator/ttlcache/v3"
	"golang.org/x/sync/singleflight"
)

func getCacheKey(c *gin.Context) (string, error) {
	if c == nil {
		return "", fmt.Errorf("undefined server context")
	}

	if c.Request == nil {
		return "", fmt.Errorf("undefined request object")
	}

	if c.Request.URL == nil {
		return "", fmt.Errorf("undefined request URL object")
	}

	return c.Request.URL.String(), nil
}

func isCacheble(c *gin.Context, rcw *ResponseCacheWriter) bool {
	return c.Request.Method == http.MethodGet &&
		!c.IsAborted() &&
		rcw.Status() >= 200 &&
		rcw.Status() < 300 // only cache 2xx response to GET
}

func getDynamicTTLValue(item *ttlcache.Item[string, any]) string {
	if item.IsExpired() {
		return "0s"
	}

	delta := item.ExpiresAt().UTC().Unix() - time.Now().UTC().Unix()
	if delta < 0 {
		return "0s"
	}

	duration, err := time.ParseDuration(fmt.Sprintf("%ds", delta))
	if err != nil {
		panic(err)
	}

	return duration.String()
}

// Original code by: https://github.com/chenyahui/gin-cache

func Cache(
	cache *ttlcache.Cache[string, any],
	recordTTL, singleFlightForgetTimerDuration time.Duration,
) gin.HandlerFunc {
	sfg := new(singleflight.Group)

	return func(c *gin.Context) {
		cacheKey, err := getCacheKey(c)
		if err != nil {
			panic(err)
		}

		item := cache.Get(cacheKey, ttlcache.WithDisableTouchOnHit[string, any]())
		if item != nil {
			if cachedResponse, ok := item.Value().(*CachedResponse); !item.IsExpired() && ok {
				cachedResponse.Send(c,
					HTTPHeader{
						Key:   "X-Cache",
						Value: "HIT",
					},
					HTTPHeader{
						Key:   "X-Cache-TTL",
						Value: getDynamicTTLValue(item),
					},
				)

				return
			}
		}

		// use ResponseCacheWriter in order to record the response
		// https://github.com/gin-gonic/gin/issues/1363#issuecomment-577722498
		rcw := &ResponseCacheWriter{
			body:           &bytes.Buffer{},
			ResponseWriter: c.Writer,
		}
		c.Writer = rcw

		var inFlight bool

		cachedResponseObj, err, _ := sfg.Do(cacheKey, func() (any, error) {
			if singleFlightForgetTimerDuration.Seconds() > 0 {
				forgetTimer := time.AfterFunc(singleFlightForgetTimerDuration, func() {
					sfg.Forget(cacheKey)
				})
				defer forgetTimer.Stop()
			}

			c.Next()

			inFlight = true

			cachedResponse := &CachedResponse{
				Status: rcw.Status(),
				Data:   rcw.body.Bytes(),
				Header: rcw.Header().Clone(),
			}

			if isCacheble(c, rcw) {
				cache.Set(cacheKey, cachedResponse, recordTTL)
			}

			return cachedResponse, nil
		})
		if err != nil {
			panic(err)
		}

		if !inFlight {
			cachedResponse, ok := cachedResponseObj.(*CachedResponse)
			if !ok {
				panic("cached object type mismatch")
			}

			cachedResponse.Send(c)
		}
	}
}
