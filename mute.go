package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jlelse/feeds"
	"github.com/mmcdole/gofeed"
	"github.com/s3rj1k/yafp/pkg/cachedregexp"
	"github.com/s3rj1k/yafp/pkg/feedhlp"
	"github.com/s3rj1k/yafp/pkg/validation"
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

func handleMuteFeed(c *gin.Context) {
	cfg := new(Mute)

	if err := c.BindQuery(cfg); err != nil {
		c.String(http.StatusBadRequest, "%s\n",
			validation.ErrorResponse(err, muteProperURLQueryParamsName()),
		)

		return
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

	c.Data(http.StatusOK, contentType, []byte(out))
}
