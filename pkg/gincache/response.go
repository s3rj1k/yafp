package gincache

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const HTTPHeaderContentType = "Content-Type"

type HTTPHeader struct {
	Key   string
	Value string
}

type CachedResponse struct {
	Header http.Header
	Data   []byte
	Status int
}

func (cr *CachedResponse) Send(c *gin.Context, headers ...HTTPHeader) {
	for key, values := range cr.Header {
		if strings.EqualFold(key, HTTPHeaderContentType) {
			continue
		}

		for _, val := range values {
			c.Header(key, val)
		}
	}

	for _, el := range headers {
		if strings.EqualFold(el.Key, HTTPHeaderContentType) {
			continue
		}

		c.Header(el.Key, el.Value)
	}

	c.Data(cr.Status, cr.Header.Get(HTTPHeaderContentType), cr.Data)

	c.Abort()
}
