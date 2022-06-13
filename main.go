package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jellydator/ttlcache/v3"
	"github.com/s3rj1k/yafp/pkg/ratelimit"
)

const (
	defaultCacheRecordTTL = 15 * time.Minute
)

//nolint:gochecknoglobals // global cache
var (
	cache *ttlcache.Cache[string, any]
)

func main() {
	if err := parseInputConfiguration(); err != nil {
		panic(err)
	}

	printInfo()

	if flagVersion {
		return
	}

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("regexp", validateRegularExpression); err != nil {
			panic(err)
		}
	}

	router.RemoveExtraSlash = true

	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	_ = router.Use(
		ratelimit.NewRateLimiter(
			cache,
			ratelimit.DefaultKeyFunc,
			ratelimit.DefaultLimiterFunc,
			ratelimit.DefaultAbortFunc,
		),
	)

	router.HandleMethodNotAllowed = true
	router.NoMethod(func(c *gin.Context) {
		c.Header("Allow", http.MethodGet+", "+http.MethodHead)
		c.String(http.StatusMethodNotAllowed, "%d Method Not Allowed\n", http.StatusMethodNotAllowed)
	})

	router.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "%d Not Found\n", http.StatusNotFound)
	})

	_ = router.HEAD("/mute", func(c *gin.Context) {
		c.String(http.StatusNoContent, "")
	})

	cache = ttlcache.New[string, any](
		ttlcache.WithTTL[string, any](defaultCacheRecordTTL),
	)

	_ = router.GET("/mute", handleMuteFeed)

	go cache.Start()
	defer cache.Stop()

	if err := router.Run(flagBindAddress); err != nil {
		panic(err)
	}
}
