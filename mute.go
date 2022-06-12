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
	"github.com/jlelse/feeds"
	"github.com/mmcdole/gofeed"
	"github.com/s3rj1k/yafp/pkg/cachedregexp"
)

const (
	feedFetchTimeoutSeconds = 90

	contentTypeRSS  = "application/rss+xml"
	contentTypeAtom = "application/atom+xml"
	contentTypeJSON = "application/feed+json"
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

func feedFetchErrorResponse(c *gin.Context, err error) {
	if errors.Is(err, context.Canceled) {
		c.String(http.StatusServiceUnavailable, "%d Service Unavailable", http.StatusServiceUnavailable)

		return
	}

	if errors.Is(err, context.DeadlineExceeded) {
		c.String(http.StatusGatewayTimeout, "%d Gateway Timeout", http.StatusGatewayTimeout)

		return
	}

	if errors.Is(err, gofeed.ErrFeedTypeNotDetected) {
		c.String(http.StatusServiceUnavailable,
			"%d Failed to detect feed type", http.StatusServiceUnavailable)

		return
	}

	var httpError gofeed.HTTPError

	if errors.As(err, &httpError) {
		c.String(httpError.StatusCode, httpError.Status)

		return
	}

	c.String(http.StatusServiceUnavailable,
		"%d Unexpected Error", http.StatusServiceUnavailable)
}

func handleMuteFeed(c *gin.Context) {
	cfg := new(Mute)

	err := c.BindQuery(cfg)
	if err != nil {
		validationMuteErrorResponse(c, err)

		return
	}

	bytesOut, err := cache.Get(c.Request.URL.RawQuery)
	if err == nil {
		switch gofeed.DetectFeedType(bytes.NewReader(bytesOut)) {
		case gofeed.FeedTypeRSS:
			c.Data(http.StatusOK, contentTypeRSS, bytesOut)

			return
		case gofeed.FeedTypeAtom:
			c.Data(http.StatusOK, contentTypeAtom, bytesOut)

			return
		case gofeed.FeedTypeJSON:
			c.Data(http.StatusOK, contentTypeJSON, bytesOut)

			return
		case gofeed.FeedTypeUnknown:
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
		feedFetchErrorResponse(c, err)

		return
	}

	if feedIn == nil {
		c.String(http.StatusServiceUnavailable,
			"%d Undefined input feed", http.StatusServiceUnavailable)

		return
	}

	currentTime := time.Now()

	feedOut := &feeds.Feed{
		Title: feedIn.Title,
		Link: &feeds.Link{
			Href: feedIn.Link,
		},
		Description: feedIn.Description,
		Copyright:   feedIn.Copyright,
	}

	if feedIn.PublishedParsed != nil {
		feedOut.Created = *feedIn.PublishedParsed
	}

	if feedIn.UpdatedParsed != nil {
		feedOut.Updated = currentTime
	}

	if len(feedIn.Authors) > 0 {
		author := feedIn.Authors[0]
		if author != nil {
			feedOut.Author.Name = author.Name
			feedOut.Author.Email = author.Email
		}
	}

	if feedIn.Image != nil {
		feedOut.Image = &feeds.Image{
			Title: feedIn.Image.Title,
			Link:  feedIn.Image.URL,
			Url:   feedIn.Image.URL,
		}
	}

	feedOut.Items = make([]*feeds.Item, 0, len(feedIn.Items))

	for _, el := range feedIn.Items {
		if el == nil {
			continue
		}

		var (
			updated time.Time
			created time.Time
		)

		if el.UpdatedParsed != nil {
			updated = *el.UpdatedParsed
		}

		if el.PublishedParsed != nil {
			created = *el.PublishedParsed
		}

		author := &feeds.Author{
			Name:  "",
			Email: "",
		}

		if el.Author != nil {
			author = &feeds.Author{
				Name:  el.Author.Name,
				Email: el.Author.Email,
			}
		}

		if (cfg.TitleQuery != "" && reTitle.MatchString(el.Title)) ||
			(cfg.DescriptionQuery != "" && reDescription.MatchString(el.Description)) {
			if cfg.RewriteAuthor == "" {
				// do not add item to resulting feed when
				// RegExp matched and RewriteAuthor not specified
				continue
			}

			author = &feeds.Author{
				Name:  cfg.RewriteAuthor,
				Email: "",
			}

			updated = currentTime
		}

		feedOut.Items = append(feedOut.Items, &feeds.Item{
			Title: el.Title,
			Link: &feeds.Link{
				Href: el.Link,
			},
			Author: &feeds.Author{
				Name:  author.Name,
				Email: author.Email,
			},
			Description: el.Description,
			Id:          el.GUID,
			Updated:     updated,
			Created:     created,
		})
	}

	var out, contentType string

	switch feedIn.FeedType {
	case "rss":
		contentType = contentTypeRSS
		out, err = feedOut.ToRss()
	case "atom":
		contentType = contentTypeAtom
		out, err = feedOut.ToAtom()
	case "json":
		contentType = contentTypeJSON
		out, err = feedOut.ToJSON()
	default:
		contentType = contentTypeRSS
		out, err = feedOut.ToRss()
	}

	if err != nil {
		c.String(http.StatusServiceUnavailable,
			"%d Unable to build feed", http.StatusServiceUnavailable)

		return
	}

	bytesOut = []byte(out)

	err = cache.Set(c.Request.URL.RawQuery, bytesOut)
	if err != nil {
		c.String(http.StatusServiceUnavailable,
			"%d Unable to cache feed data", http.StatusServiceUnavailable)

		return
	}

	c.Data(http.StatusOK, contentType, bytesOut)
}
