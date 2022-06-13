package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jellydator/ttlcache/v3"
	"github.com/jlelse/feeds"
	"github.com/mmcdole/gofeed"
	"github.com/s3rj1k/yafp/pkg/cachedregexp"
	"github.com/s3rj1k/yafp/pkg/feedhlp"
)

const (
	feedFetchTimeoutSeconds = 90
)

type Mute struct {
	FeedURL          string `form:"feed_url" binding:"required,url"`
	TitleQuery       string `form:"title_query" binding:"required_without=DescriptionQuery,regexp"`
	DescriptionQuery string `form:"description_query" binding:"required_without=TitleQuery,regexp"`
	RewriteAuthor    string `form:"rewrite_author" binding:"ascii"`
}

func muteProperURLQueryParamsName() *strings.Replacer {
	return strings.NewReplacer(
		"FeedURL", "feed_url",
		"TitleQuery", "title_query",
		"DescriptionQuery", "description_query",
		"RewriteAuthor", "rewrite_author",
	)
}

func validationMuteErrorResponse(c *gin.Context, err error) {
	validationErrors := new(validator.ValidationErrors)
	resp := make([]string, 0)

	resp = append(resp, fmt.Sprintf("%d Bad Request", http.StatusBadRequest))

	if errors.As(err, validationErrors) {
		for _, el := range *validationErrors {
			resp = append(resp,
				fmt.Sprintf(
					"* URL query parameter validation for '%s' failed on the '%s' tag",
					muteProperURLQueryParamsName().Replace(el.Field()),
					el.Tag(),
				),
			)
		}
	}

	c.String(http.StatusBadRequest, "%s\n", strings.Join(resp, "\n"))
}

func handleMuteFeed(c *gin.Context) {
	cfg := new(Mute)

	err := c.BindQuery(cfg)
	if err != nil {
		validationMuteErrorResponse(c, err)

		return
	}

	item := cache.Get(c.Request.URL.RawQuery, ttlcache.WithDisableTouchOnHit[string, any]())
	if err != nil {
		bytesOut, ok := item.Value().([]byte)
		if ok {
			switch gofeed.DetectFeedType(bytes.NewReader(bytesOut)) {
			case gofeed.FeedTypeRSS:
				c.Data(http.StatusOK, feedhlp.ContentTypeRSS, bytesOut)

				return
			case gofeed.FeedTypeAtom:
				c.Data(http.StatusOK, feedhlp.ContentTypeAtom, bytesOut)

				return
			case gofeed.FeedTypeJSON:
				c.Data(http.StatusOK, feedhlp.ContentTypeJSON, bytesOut)

				return
			case gofeed.FeedTypeUnknown:
			}
		}
	}

	reTitle := cachedregexp.MustCompile(cache, cfg.TitleQuery)
	reDescription := cachedregexp.MustCompile(cache, cfg.DescriptionQuery)

	fp := gofeed.NewParser()

	fp.UserAgent = c.Request.UserAgent()

	ctx, cancel := context.WithTimeout(context.Background(), feedFetchTimeoutSeconds*time.Second)
	defer cancel()

	feedIn, err := fp.ParseURLWithContext(cfg.FeedURL, ctx)
	if err != nil {
		c.String(feedhlp.HTTPErrorResponse(err))

		return
	}

	if feedIn == nil {
		c.String(http.StatusServiceUnavailable,
			"%d Undefined input feed", http.StatusServiceUnavailable)

		return
	}

	currentTime := time.Now()

	feedOut := feedhlp.MutateFeed(feedIn, func(item *feeds.Item) *feeds.Item {
		if (cfg.TitleQuery != "" && reTitle.MatchString(item.Title)) ||
			(cfg.DescriptionQuery != "" && reDescription.MatchString(item.Description)) {
			if cfg.RewriteAuthor == "" {
				// do not add item to resulting feed when
				// RegExp matched and RewriteAuthor not specified
				return nil
			}

			item.Author = &feeds.Author{
				Name:  cfg.RewriteAuthor,
				Email: "",
			}

			item.Updated = currentTime
		}

		return item
	})

	contentType := feedhlp.GetContentTypeFromFeed(feedIn)
	if contentType == "" {
		contentType = feedhlp.ContentTypeRSS
	}

	out, err := feedhlp.RenderFeedBasedOnProvidedContentType(feedOut, contentType)
	if err != nil {
		c.String(http.StatusServiceUnavailable,
			"%d Unable to build feed", http.StatusServiceUnavailable)

		return
	}

	bytesOut := []byte(out)

	_ = cache.Set(c.Request.URL.RawQuery, bytesOut, ttlcache.DefaultTTL)

	c.Data(http.StatusOK, contentType, bytesOut)
}
