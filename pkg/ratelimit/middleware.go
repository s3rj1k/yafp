package ratelimit

import (
	"net/http"

	"github.com/didip/tollbooth"
	"github.com/gin-gonic/gin"
)

func HandlerFunc(max float64) gin.HandlerFunc {
	lmt := tollbooth.NewLimiter(max, nil)

	return func(c *gin.Context) {
		if err := tollbooth.LimitByRequest(lmt, c.Writer, c.Request); err != nil {
			c.String(http.StatusTooManyRequests, "%d Too Many Requests\n", http.StatusTooManyRequests)
			c.Abort()

			return
		}

		c.Next()
	}
}
