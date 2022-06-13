package gincache

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const HTTPHeaderContentType = "Content-Type"

type CachedResponse struct {
	Header http.Header
	Data   []byte
	Status int
}

func (cr *CachedResponse) Send(c *gin.Context) {
	for key, values := range cr.Header {
		if strings.EqualFold(key, HTTPHeaderContentType) {
			continue
		}

		for _, val := range values {
			c.Header(key, val)
		}
	}

	c.Data(cr.Status, cr.Header.Get(HTTPHeaderContentType), cr.Data)

	c.Abort()
}
