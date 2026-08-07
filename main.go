package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/paranoidupdate/rssreader/internal/config"
	"github.com/paranoidupdate/rssreader/internal/core"
	"github.com/paranoidupdate/rssreader/internal/storage"
)

const maxConcurrency = 5 // TODO: Put into config

func getGUID(feedItem *gofeed.Item) (string, bool) {
	guid := feedItem.GUID
	if guid == "" {
		guid = feedItem.Link
	}

	if guid == "" {
		return guid, false
	}
	return guid, true
}

func getDescription(feedItem *gofeed.Item) *string {
	description := feedItem.Description
	if description == "" {
		return nil
	}
	return &description
}

func parseItems(feed *core.Feed, feedItems []*gofeed.Item) []core.Item {
	items := make([]core.Item, 0)
	for _, feedItem := range feedItems {
		if guid, ok := getGUID(feedItem); ok {
			item := core.Item{
				FeedID:      feed.ID,
				GUID:        guid,
				Title:       feedItem.Title,
				Description: getDescription(feedItem),
				Link:        feedItem.Link,
				CreatedAt:   time.Time{}, // Can be omitted, keep for clarity
				PublishedAt: feedItem.PublishedParsed,
				UpdatedAt:   feedItem.UpdatedParsed, // TODO: Updates are not supported yet
			}
			items = append(items, item)
		} else {
			slog.Info("Item doesn't have GUID", "feed_link", feed.FeedLink, "item_title", feedItem.Title)
			// TODO: Add metric for items without GUID
		}
	}
	return items
}

func run() error {
	conf, err := config.InitConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	slog.Info("Config loaded")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, err := storage.NewStorage(ctx, conf.DbAddress)
	if err != nil {
		return fmt.Errorf("couldn't connect to the database: %w", err)
	}
	defer db.Close()

	feeds, err := db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get feed info: %w", err)
	}

	fp := gofeed.NewParser()

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency) // chan as a semaphore

	for _, feed := range feeds {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			fctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			feedData, err := fp.ParseURLWithContext(feed.FeedLink, fctx)
			if err != nil {
				slog.Error("can't parse feed", "feed_link", feed.FeedLink, "error", err)
			} else {
				slog.Info("reading feed", "feed_title", feedData.Title)
				items := parseItems(&feed, feedData.Items)
				sctx, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()
				if storeResult, err := db.StoreFeedItems(sctx, &feed, items); err != nil {
					slog.Error("feed processing failed", "feed_link", feed.FeedLink, "error", err)
				} else {
					slog.Info("items stored", "feed_link", feed.FeedLink, "saved_items", storeResult.Saved, "skipped_items", storeResult.Skipped)
				}
			}
		})
	}
	wg.Wait()
	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("Fetcher failed", "error", err)
		os.Exit(1)
	}
}
