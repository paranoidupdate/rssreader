package core

import "time"

type Feed struct {
	ID        int64      `db:"id"`
	FeedLink  string     `db:"feed_link"`
	PageLink  *string    `db:"page_link"`
	Title     string     `db:"title"`
	CreatedAt time.Time  `db:"created_at"`
	FetchedAt *time.Time `db:"fetched_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}
