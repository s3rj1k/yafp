package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jlelse/feeds"
	"github.com/s3rj1k/yafp/pkg/capitalise"
	"github.com/s3rj1k/yafp/pkg/feedhlp"
	"github.com/s3rj1k/yafp/pkg/tgscrapper"
)

type TG struct {
	Name string `uri:"name" binding:"required,tg"`
}

func handleTG(c *gin.Context) {
	cfg := new(TG)

	if err := c.BindUri(&cfg); err != nil {
		c.String(http.StatusBadRequest, "invalid telegram channel name\n")

		return
	}

	data := tgscrapper.NewMessages(cfg.Name)

	ctx, cancel := context.WithTimeout(context.Background(), feedFetchTimeoutSeconds*time.Second)
	defer cancel()

	err := tgscrapper.Worker(ctx, data, c.Request.UserAgent())
	if err != nil {
		c.String(http.StatusServiceUnavailable,
			"%d %s", http.StatusServiceUnavailable, capitalise.First(err.Error()))

		return
	}

	contentType := feedhlp.ContentTypeRSS

	feedOut := &feeds.Feed{
		Title: data.ChannelTitle,
		Link: &feeds.Link{
			Href: data.ChannelLink,
		},
		Description: data.ChannelDescription,
		Updated:     data.GenerationTime,
		Created:     data.OldestMessageDate,
	}

	feedOut.Items = make([]*feeds.Item, 0, len(data.Items))

	for _, el := range data.Items {
		if el == nil {
			continue
		}

		item := new(feeds.Item)

		item.Id = strconv.Itoa(el.ID)
		item.Title = el.Title
		item.Description = el.Body
		item.Link = &feeds.Link{
			Href: el.Link,
		}

		if !el.DateTime.IsZero() {
			item.Created = el.DateTime
			item.Updated = el.DateTime
		}

		item.Author = &feeds.Author{
			Name: el.Author,
		}

		feedOut.Items = append(feedOut.Items, item)
	}

	out, err := feedhlp.RenderFeedBasedOnProvidedContentType(feedOut, contentType)
	if err != nil {
		c.String(http.StatusServiceUnavailable,
			"%d Unable to build feed", http.StatusServiceUnavailable)

		return
	}

	c.Data(http.StatusOK, contentType, []byte(out))
}
