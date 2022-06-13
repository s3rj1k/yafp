package ratelimit_test

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jellydator/ttlcache/v3"
	"github.com/s3rj1k/yafp/pkg/ratelimit"
)

func getFreePort() int {
	for {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil || addr == nil {
			continue
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil || l == nil {
			continue
		}

		var port int

		if val, ok := l.Addr().(*net.TCPAddr); ok {
			port = val.Port
		}

		_ = l.Close()

		if port != 0 {
			return port
		}
	}
}

func TestLimitByKey(t *testing.T) {
	t.Parallel()

	bindHost := net.JoinHostPort("localhost", strconv.Itoa(getFreePort()))
	endpoint := url.URL{
		Scheme: "http",
		Host:   bindHost,
	}

	cache := ttlcache.New[string, any](
		ttlcache.WithTTL[string, any](time.Minute),
	)

	r := gin.Default()

	r.Use(
		ratelimit.NewRateLimiter(
			cache,
			ratelimit.DefaultKeyFunc,
			ratelimit.DefaultLimiterFunc,
			ratelimit.DefaultAbortFunc,
		),
	)

	r.GET("/", func(c *gin.Context) {})

	go func() {
		if err := r.Run(bindHost); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				t.Errorf("Error run http server: %s", err.Error())
			}
		}
	}()

	for i := 0; i < ratelimit.DefaultBurst*2; i++ {
		resp, err := http.DefaultClient.Get(endpoint.String())
		if err != nil {
			t.Errorf("Error during requests: %s", err.Error())

			return
		}

		switch {
		case i < ratelimit.DefaultBurst:
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Unexpected status code on %d request: %d ", i, resp.StatusCode)
			}
		case i >= ratelimit.DefaultBurst:
			if resp.StatusCode != http.StatusTooManyRequests {
				t.Errorf("Threashold break not detected on %d request", i)
			}
		}
	}

	cache.DeleteAll()

	for i := 0; i < ratelimit.DefaultBurst/2; i++ {
		resp, err := http.DefaultClient.Get(endpoint.String())
		if err != nil {
			t.Fatalf("Error during requests: %s", err.Error())

			return
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Unexpected status code: %d", resp.StatusCode)
		}
	}
}
