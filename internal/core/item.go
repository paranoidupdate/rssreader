package core

import "time"

type Item struct {
	ID          int64      `db:"id"`
	FeedID      int64      `db:"feed_id"`
	GUID        string     `db:"guid"`
	Title       string     `db:"title"`
	Description *string    `db:"description"`
	Link        string     `db:"link"`
	CreatedAt   time.Time  `db:"created_at"`
	PublishedAt *time.Time `db:"published_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
}
