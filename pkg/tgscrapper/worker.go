package tgscrapper

import (
	"context"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func Worker(ctx context.Context, data *TGMessages, userAgent string) error {
	for i := 0; i < maxNumberOfPages; i++ {
		doc, err := Get(ctx, data, userAgent)
		if err != nil {
			return err
		}

		if doc == nil {
			continue
		}

		data = Parse(doc, data)

		if val := data.GenerationTime.Sub(data.OldestMessageDate); val.Hours() > maxNumberOfHoursAgoForOldestMessage {
			break
		}

		if len(data.Items) > maxNumberOfMessages {
			break
		}
	}

	data.Sort()

	return nil
}

func Get(ctx context.Context, data *TGMessages, userAgent string) (*goquery.Document, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, data.MustPaginationURL(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("prepare request error: %w", err)
	}

	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("run request error: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %w", err)
	}

	return doc, nil
}

func Parse(doc *goquery.Document, data *TGMessages) *TGMessages {
	var err error

	if data.ChannelTitle == "" {
		data.ChannelTitle, err = GetChannelTitle(doc.Contents())
		if err != nil {
			data.ChannelTitle = data.ChannelName
		}
	}

	if data.ChannelDescription == "" {
		data.ChannelDescription, err = GetChannelDescription(doc.Contents())
		if err != nil {
			data.ChannelDescription = ""
		}
	}

	if data.ChannelLink == "" {
		data.ChannelLink, err = GetChannelLink(doc.Contents())
		if err != nil {
			data.ChannelLink = "@" + data.ChannelName
		}
	}

	doc.Find("div.tgme_widget_message_bubble").Each(func(i int, selection *goquery.Selection) {
		var err error

		tgm := new(TGMessage)

		tgm.Title, tgm.Body, err = GetMessage(selection)
		if err != nil {
			return
		}

		tgm.DateTime, err = GetMessageDateTime(selection)
		if err != nil {
			return
		}

		tgm.Link, err = GetMessageLink(selection)
		if err != nil {
			return
		}

		err = tgm.PopulateMessageID()
		if err != nil {
			return
		}

		tgm.Author, err = GetMessageAuthor(selection)
		if err != nil {
			return
		}

		data.Store(tgm)
	})

	return data
}
