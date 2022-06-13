package feedhlp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/mmcdole/gofeed"
)

func HTTPErrorResponse(err error) (statusCode int, status string) {
	if err == nil {
		return http.StatusOK, fmt.Sprintf("%d OK", http.StatusOK)
	}

	if errors.Is(err, context.Canceled) {
		return http.StatusServiceUnavailable, fmt.Sprintf("%d Service Unavailable", http.StatusServiceUnavailable)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout, fmt.Sprintf("%d Gateway Timeout", http.StatusGatewayTimeout)
	}

	if errors.Is(err, gofeed.ErrFeedTypeNotDetected) {
		return http.StatusServiceUnavailable, fmt.Sprintf("%d Failed to detect feed type", http.StatusServiceUnavailable)
	}

	var httpError gofeed.HTTPError

	if errors.As(err, &httpError) {
		return httpError.StatusCode, httpError.Status
	}

	return http.StatusServiceUnavailable, fmt.Sprintf("%d Unexpected Error", http.StatusServiceUnavailable)
}
