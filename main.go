package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

const (
	defaultRateLimitMaxValue = 20
	defaultAbbRevisionNum    = 8
)

//nolint:gochecknoglobals // CLI configuration flags
var (
	flagBindAddress string

	flagVersion bool
)

func printInfo() {
	vcsRev, vcsDate := getVCSInfo(defaultAbbRevisionNum)
	if vcsDate.Unix() == 0 {
		fmt.Printf("[GIN] %s | VCS_REV: %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"), // https://go.dev/src/time/format.go
			vcsRev,
		)
	} else {
		fmt.Printf("[GIN] %s | VCS_REV: %s | VCS_DATE: %v\n",
			time.Now().Format("2006/01/02 - 15:04:05"), // https://go.dev/src/time/format.go
			vcsRev, vcsDate,
		)
	}
}

func parseInputConfiguration() error {
	flag.BoolVar(&flagVersion, "version", false, "Show build information and exit")
	flag.StringVar(&flagBindAddress, "bind-address", ":8080", "Address for HTTP server bind")

	flag.Parse()

	return nil
}

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
		if err := v.RegisterValidation("regexp", validRegularExpression); err != nil {
			panic(err)
		}
	}

	router.RemoveExtraSlash = true

	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	router.Use(rateLimit(defaultRateLimitMaxValue))

	router.HandleMethodNotAllowed = true
	router.NoMethod(func(c *gin.Context) {
		c.Header("Allow", http.MethodGet+", "+http.MethodHead)
		c.String(http.StatusMethodNotAllowed, "%d Method Not Allowed\n", http.StatusMethodNotAllowed)
	})

	router.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "%d Not Found\n", http.StatusNotFound)
	})

	router.HEAD("/mute", func(c *gin.Context) {
		c.String(http.StatusNoContent, "")
	})

	router.GET("/mute", handleMuteFeed)

	if err := router.Run(flagBindAddress); err != nil {
		panic(err)
	}
}
