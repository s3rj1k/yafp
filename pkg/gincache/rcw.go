package gincache

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

type ResponseCacheWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (rcw *ResponseCacheWriter) Write(b []byte) (int, error) {
	n, err := rcw.body.Write(b)
	if err != nil {
		return n, err //nolint:wrapcheck // pass bytes.Buffer error unwrapped
	}

	return rcw.ResponseWriter.Write(b) //nolint:wrapcheck // pass http.ResponseWriter error unwrapped
}
