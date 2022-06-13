package gincache_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jellydator/ttlcache/v3"
	"github.com/s3rj1k/yafp/pkg/gincache"
	"github.com/stretchr/testify/assert"
)

const (
	defaultSingleFlightForgetTimerDuration = 90 * time.Second
)

func init() { //nolint:gochecknoinits // GIN test initialization
	gin.SetMode(gin.TestMode)
}

//revive:disable:flag-parameter // ignore this for unit tests

func mockHTTPRequest(middleware gin.HandlerFunc, url string, withRand bool) *httptest.ResponseRecorder {
	testWriter := httptest.NewRecorder()

	_, engine := gin.CreateTestContext(testWriter)
	engine.Use(middleware)
	engine.GET("/cache", func(c *gin.Context) {
		body := "uid:" + c.Query("uid")
		if withRand {
			body += fmt.Sprintf(",rand:%d", rand.Int()) //nolint:gosec // no need for secure random number generator
		}

		c.String(http.StatusOK, body)
	})

	testRequest := httptest.NewRequest(http.MethodGet, url, nil)

	engine.ServeHTTP(testWriter, testRequest)

	return testWriter
}

//revive:enable:flag-parameter

func TestCache(t *testing.T) {
	t.Parallel()

	cache := ttlcache.New[string, any](
		ttlcache.WithTTL[string, any](time.Minute),
	)

	cacheMiddleware := gincache.Cache(cache, 3*time.Second, defaultSingleFlightForgetTimerDuration)

	w1 := mockHTTPRequest(cacheMiddleware, "/cache?uid=u1", true)
	w2 := mockHTTPRequest(cacheMiddleware, "/cache?uid=u1", true)
	w3 := mockHTTPRequest(cacheMiddleware, "/cache?uid=u2", true)

	assert.Equal(t, w1.Body, w2.Body)
	assert.Equal(t, w1.Code, w2.Code)

	assert.NotEqual(t, w2.Body, w3.Body)

	w4 := mockHTTPRequest(cacheMiddleware, "/cache?uid=u4", false)
	assert.Equal(t, "uid:u4", w4.Body.String())
}

func TestCacheDuration(t *testing.T) {
	t.Parallel()

	cache := ttlcache.New[string, any](
		ttlcache.WithTTL[string, any](time.Minute),
	)

	cacheMiddleware := gincache.Cache(cache, 3*time.Second, defaultSingleFlightForgetTimerDuration)

	w1 := mockHTTPRequest(cacheMiddleware, "/cache?uid=u1", true)

	time.Sleep(1 * time.Second)

	w2 := mockHTTPRequest(cacheMiddleware, "/cache?uid=u1", true)
	assert.Equal(t, w1.Body, w2.Body)
	assert.Equal(t, w1.Code, w2.Code)
	time.Sleep(2 * time.Second)

	w3 := mockHTTPRequest(cacheMiddleware, "/cache?uid=u1", true)
	assert.NotEqual(t, w1.Body, w3.Body)
}

func TestHeader(t *testing.T) {
	t.Parallel()

	testWriter := httptest.NewRecorder()

	_, engine := gin.CreateTestContext(testWriter)

	cache := ttlcache.New[string, any](
		ttlcache.WithTTL[string, any](time.Minute),
	)

	cacheMiddleware := gincache.Cache(cache, 3*time.Second, defaultSingleFlightForgetTimerDuration)

	engine.Use(cacheMiddleware)

	engine.Use(func(c *gin.Context) {
		c.Header("test_header_key", "test_header_value")
	})

	engine.GET("/cache", func(c *gin.Context) {
		c.Header("test_header_key", "test_header_value2")
		c.String(http.StatusOK, "value")
	})

	testRequest := httptest.NewRequest(http.MethodGet, "/cache", nil)

	engine.ServeHTTP(testWriter, testRequest)
	values := testWriter.Header().Values("test_header_key")
	assert.Equal(t, 1, len(values))
	assert.Equal(t, "test_header_value2", values[0])
}

func TestConcurrentRequest(t *testing.T) {
	t.Parallel()

	cache := ttlcache.New[string, any](
		ttlcache.WithTTL[string, any](time.Minute),
	)

	cacheMiddleware := gincache.Cache(cache, 1*time.Second, defaultSingleFlightForgetTimerDuration)

	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			uid := rand.Intn(5) //nolint:gosec // no need for secure random number generator
			url := fmt.Sprintf("/cache?uid=%d", uid)
			expect := fmt.Sprintf("uid:%d", uid)

			writer := mockHTTPRequest(cacheMiddleware, url, false)
			assert.Equal(t, expect, writer.Body.String())
		}()
	}

	wg.Wait()
}

func TestWriteHeader(t *testing.T) {
	t.Parallel()

	cache := ttlcache.New[string, any](
		ttlcache.WithTTL[string, any](time.Minute),
	)

	cacheMiddleware := gincache.Cache(cache, 1*time.Second, defaultSingleFlightForgetTimerDuration)

	testWriter := httptest.NewRecorder()

	_, engine := gin.CreateTestContext(testWriter)
	engine.Use(cacheMiddleware)
	engine.GET("/cache", func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Set("hello", "world")
	})

	testRequest := httptest.NewRequest(http.MethodGet, "/cache", nil)
	engine.ServeHTTP(testWriter, testRequest)
	assert.Equal(t, "world", testWriter.Header().Get("hello"))
}
