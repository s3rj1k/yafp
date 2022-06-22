package tgscrapper

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/araddon/dateparse"
)

func GetMessage(selection *goquery.Selection) (title, body string, err error) {
	item := selection.Find("div.tgme_widget_message_text").First()
	if item.Length() == 0 {
		return "", "", ErrNotFound
	}

	body, err = item.Html()
	if err != nil {
		return "", "", fmt.Errorf("message body error: %w", ErrNoData)
	}

	body = strings.TrimSpace(body)

	title = Ellipsize(Textify(item))
	if title == "" {
		return "", "", fmt.Errorf("message title error: %w", ErrNoData)
	}

	if strings.Contains(title, "http://") || strings.Contains(title, "https://") {
		if val := selection.Find("div.link_preview_title").First().Text(); val != "" {
			return val, body, nil
		}
	}

	return title, body, nil
}

func GetMessageAuthor(selection *goquery.Selection) (name string, err error) {
	item := selection.Find("a.tgme_widget_message_owner_name").First()
	if item.Length() == 0 {
		return "", ErrNotFound
	}

	name = item.Text()
	if name == "" {
		return "", fmt.Errorf("author name error: %w", ErrNoData)
	}

	return name, nil
}

func GetMessageDateTime(selection *goquery.Selection) (time.Time, error) {
	val, exists := selection.Find("a.tgme_widget_message_date").First().Find("time").Attr("datetime")
	if !exists {
		return time.Time{}, ErrNotFound
	}

	if val == "" {
		return time.Time{}, ErrNoData
	}

	tt, err := dateparse.ParseAny(val)
	if err != nil {
		return time.Time{}, ErrInvalidData
	}

	return tt.UTC().Round(time.Second), nil
}

func GetMessageLink(selection *goquery.Selection) (string, error) {
	val, exists := selection.Find("a.tgme_widget_message_date").First().Attr("href")
	if !exists {
		return "", ErrNotFound
	}

	if val == "" {
		return "", ErrNoData
	}

	return strings.Replace(val, "://t.me/", "://t.me/s/", 1), nil
}

func GetChannelTitle(selection *goquery.Selection) (string, error) {
	val, exists := selection.Find("meta[property='og:title']").First().Attr("content")
	if !exists {
		return "", ErrNotFound
	}

	if val == "" {
		return "", ErrNoData
	}

	return val, nil
}

func GetChannelDescription(selection *goquery.Selection) (string, error) {
	val, exists := selection.Find("meta[property='og:description']").First().Attr("content")
	if !exists {
		return "", ErrNotFound
	}

	if val == "" {
		return "", ErrNoData
	}

	return strings.ReplaceAll(val, "\n", "<br/>"), nil
}

func GetChannelLink(selection *goquery.Selection) (string, error) {
	val, exists := selection.Find("meta[property='al:android:url']").First().Attr("content")
	if !exists {
		return "", ErrNotFound
	}

	if val == "" {
		return "", ErrNoData
	}

	return val, nil
}
