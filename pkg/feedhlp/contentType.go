package feedhlp

import (
	"errors"

	"github.com/jlelse/feeds"
	"github.com/mmcdole/gofeed"
)

const (
	ContentTypeRSS  = "application/rss+xml"
	ContentTypeAtom = "application/atom+xml"
	ContentTypeJSON = "application/feed+json"
)

var ErrUnknownContentType = errors.New("unknown Content-Type")

func GetContentTypeFromFeed(feed *gofeed.Feed) string {
	var contentType string

	if feed == nil {
		return ""
	}

	switch feed.FeedType {
	case "rss":
		contentType = ContentTypeRSS
	case "atom":
		contentType = ContentTypeAtom
	case "json":
		contentType = ContentTypeJSON
	}

	return contentType
}

func RenderFeedBasedOnProvidedContentType(feed *feeds.Feed, contentType string) (string, error) {
	var (
		out string
		err error
	)

	switch contentType {
	case ContentTypeRSS:
		out, err = feed.ToRss()
	case ContentTypeAtom:
		out, err = feed.ToAtom()
	case ContentTypeJSON:
		out, err = feed.ToJSON()
	default:
		err = ErrUnknownContentType
	}

	return out, err
}
