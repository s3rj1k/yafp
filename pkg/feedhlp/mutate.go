package feedhlp

import (
	"time"

	"github.com/jlelse/feeds"
	"github.com/mmcdole/gofeed"
)

func MutateFeed(feedIn *gofeed.Feed, mutateFeedItemFunc func(item *feeds.Item) *feeds.Item) *feeds.Feed {
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

		item := new(feeds.Item)

		item.Id = el.GUID
		item.Title = el.Title
		item.Description = el.Description
		item.Link = &feeds.Link{
			Href: el.Link,
		}

		var created, updated time.Time

		if el.PublishedParsed != nil {
			created = *el.PublishedParsed
		}

		if el.UpdatedParsed != nil {
			updated = *el.UpdatedParsed
		}

		if !created.IsZero() {
			item.Created = created
		}

		if !updated.IsZero() {
			item.Updated = updated
		}

		author := new(feeds.Author)

		if el.Author != nil {
			author.Name = el.Author.Name
			author.Email = el.Author.Email
		}

		item = mutateFeedItemFunc(item)
		if item == nil {
			continue
		}

		feedOut.Items = append(feedOut.Items, item)
	}

	return feedOut
}
