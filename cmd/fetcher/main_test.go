package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/paranoidupdate/rssreader/internal/core"
)

func ptr[T any](v T) *T { return &v }

func TestGetGUID(t *testing.T) {
	tests := []struct {
		name     string
		guid     string
		link     string
		wantGuid string
		wantOk   bool
	}{
		{"has guid, no link", "awesome guid", "", "awesome guid", true},
		{"no guid, has link", "", "amazing link", "amazing link", true},
		{"has guid, has link", "awesome guid", "amazing link", "awesome guid", true},
		{"no guid, no link", "", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := gofeed.Item{GUID: tt.guid, Link: tt.link}
			gotGuid, gotOk := getGUID(&item)

			if tt.wantOk != gotOk || tt.wantGuid != gotGuid {
				t.Fatalf("got (%t) %q, want (%t) %q", gotOk, gotGuid, tt.wantOk, tt.wantGuid)
			}
		})
	}
}

func TestGetDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        *string
	}{
		{"has description", "awesome feed", ptr("awesome feed")},
		{"empty description", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := gofeed.Item{Description: tt.description}
			got := getDescription(&item)

			if (got == nil) != (tt.want == nil) {
				t.Fatalf("got %p, want %p", got, tt.want)
			}
			if got == nil && tt.want == nil {
				return
			}
			if *got != *tt.want {
				t.Fatalf("got %q, want %q", *got, *tt.want)
			}
		})
	}
}

func TestParseItems(t *testing.T) {
	testTime := time.Date(2026, time.August, 11, 20, 1, 30, 50, time.UTC)

	feed := core.Feed{ID: 123}
	items := []*gofeed.Item{
		{
			// Normal case
			Title:           "Test",
			GUID:            "unique guid",
			Description:     "great reading",
			Link:            "https://example.com",
			PublishedParsed: &testTime,
			UpdatedParsed:   &testTime,
		},
		{
			// No GUID
			Title:           "Test2",
			GUID:            "",
			Description:     "great reading",
			Link:            "https://example.com",
			PublishedParsed: &testTime,
			UpdatedParsed:   &testTime,
		},
		{
			// No GUID, no link. Should be skipped
			Title:           "Test3",
			GUID:            "",
			Description:     "great reading",
			Link:            "",
			PublishedParsed: &testTime,
			UpdatedParsed:   &testTime,
		},
		{
			// No description
			Title:           "Test4",
			GUID:            "unique guid",
			Description:     "",
			Link:            "https://example.com",
			PublishedParsed: &testTime,
			UpdatedParsed:   &testTime,
		},
	}
	want := []core.Item{
		{
			FeedID:      feed.ID,
			Title:       "Test",
			GUID:        "unique guid",
			Description: ptr("great reading"),
			Link:        "https://example.com",
			CreatedAt:   time.Time{},
			PublishedAt: &testTime,
			UpdatedAt:   &testTime,
		},
		{
			FeedID:      feed.ID,
			Title:       "Test2",
			GUID:        "https://example.com",
			Description: ptr("great reading"),
			Link:        "https://example.com",
			CreatedAt:   time.Time{},
			PublishedAt: &testTime,
			UpdatedAt:   &testTime,
		},
		{
			FeedID:      feed.ID,
			Title:       "Test4",
			GUID:        "unique guid",
			Description: nil,
			Link:        "https://example.com",
			CreatedAt:   time.Time{},
			PublishedAt: &testTime,
			UpdatedAt:   &testTime,
		},
	}

	gotItems := parseItems(&feed, items)
	if len(gotItems) != len(want) {
		t.Fatalf("Got %d items , want %d items", len(gotItems), len(want))
	}
	for i := 0; i < len(want); i++ {
		if !reflect.DeepEqual(gotItems[i], want[i]) {
			t.Errorf("iter %d: got %+v, want %+v", i, gotItems[i], want[i])
		}
	}
}
