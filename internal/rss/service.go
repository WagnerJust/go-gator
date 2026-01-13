package rss

import (
	"context"
	"encoding/xml"
	"html"
	"io"
	"net/http"
)
func (f *RSSFeed) cleanupUnescapedEntities () {
	f.Channel.Title = html.UnescapeString(f.Channel.Title)
	if f.Channel.Description != nil {
		unescaped := html.UnescapeString(*f.Channel.Description)
		f.Channel.Description = &unescaped
	}
	for index, item := range f.Channel.Item {
		f.Channel.Item[index].Title = html.UnescapeString(item.Title)
		if item.Description != nil {
			unescaped := html.UnescapeString(*item.Description)
			f.Channel.Item[index].Description = &unescaped
		}
	}
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}
	req.Header.Set("user-agent", "go-gator")

	var client http.Client
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}
	defer res.Body.Close()
	var feed RSSFeed

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, err
	}

	err = xml.Unmarshal(data, &feed)
	if err != nil {
		return &RSSFeed{}, err
	}

	feed.cleanupUnescapedEntities()

	return &feed, nil
}
