package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paranoidupdate/rssreader/internal/core"
)

type StoreResult struct {
	Saved, Skipped int
}

type Storage struct {
	dbpool *pgxpool.Pool
}

func (s *Storage) AddFeed(feedLink string) error {
	return nil
}

// Store items fetched from a feed in a database and update feed's fetched_at time.
// Don't update fetched_at if item storing failed (batch).
func (s *Storage) StoreFeedItems(ctx context.Context, feed *core.Feed, items []core.Item) (*StoreResult, error) {
	batch := &pgx.Batch{}
	query := `INSERT INTO items (feed_id, guid, title, description, link, published_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (feed_id, guid) DO NOTHING`
	for _, item := range items {
		batch.Queue(query, item.FeedID, item.GUID, item.Title, item.Description, item.Link, item.PublishedAt, item.UpdatedAt)
	}
	batch.Queue("UPDATE feeds SET fetched_at = now() WHERE id = $1", feed.ID)
	batchResult := s.dbpool.SendBatch(ctx, batch)
	defer batchResult.Close()

	var storedItems, skippedItems int
	for i := 0; i < len(items); i++ {
		commandTag, err := batchResult.Exec()
		if err != nil {
			return nil, fmt.Errorf("failed saving batch item %d: %w", i, err)
		}
		if commandTag.RowsAffected() == 0 {
			skippedItems++
			// TODO: Add metrics for saved and skipped items
		} else {
			storedItems++
		}
	}

	// Check UPDATE result
	commandTag, err := batchResult.Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to update feed ID %d: %w", feed.ID, err)
	}
	if commandTag.RowsAffected() != 1 {
		// 0 rows = feed deleted mid-cycle.
		// Can't happen until feed deletion exists (TBD)
		// FK doesn't protect this (no lock taken when all items conflict or batch is empty.
		return nil, fmt.Errorf("feed %d missing after item insert (items committed, fetched_at not advanced)", feed.ID)
	}

	return &StoreResult{storedItems, skippedItems}, nil
}

func (s *Storage) GetItems(feedLink string) ([]core.Item, error) {
	return nil, nil
}

func (s *Storage) GetFeeds(ctx context.Context) ([]core.Feed, error) {
	rows, err := s.dbpool.Query(ctx, `SELECT id, feed_link, page_link, title, created_at, fetched_at, updated_at FROM feeds`)
	if err != nil {
		return nil, fmt.Errorf("querying feeds: %w", err)
	}

	feeds, err := pgx.CollectRows(rows, pgx.RowToStructByName[core.Feed])
	if err != nil {
		return nil, fmt.Errorf("parsing feed rows: %w", err)
	}

	return feeds, nil
}

func (s *Storage) Close() {
	s.dbpool.Close()
}

func NewStorage(ctx context.Context, dbAddress string) (*Storage, error) {
	dbpool, err := pgxpool.New(ctx, dbAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	return &Storage{dbpool}, nil
}
