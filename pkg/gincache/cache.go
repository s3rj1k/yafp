package gincache

import (
	"bytes"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/jellydator/ttlcache/v3"
	"golang.org/x/sync/singleflight"
)

// Original code by: https://github.com/chenyahui/gin-cache

func Cache(
	cache *ttlcache.Cache[string, any],
	recordTTL, singleFlightForgetTimerDuration time.Duration,
) gin.HandlerFunc {
	sfg := new(singleflight.Group)

	return func(c *gin.Context) {
		cacheKey := c.Request.URL.RawQuery

		// FIXME:
		spew.Dump(cacheKey)

		item := cache.Get(cacheKey, ttlcache.WithDisableTouchOnHit[string, any]())
		if item != nil {
			if cachedResponse, ok := item.Value().(*CachedResponse); ok {
				cachedResponse.Send(c)

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

			if !c.IsAborted() && rcw.Status() >= 200 && rcw.Status() < 300 {
				cache.Set(cacheKey, cachedResponse, recordTTL) // only cache 2xx response
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
