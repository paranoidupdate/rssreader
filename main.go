package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
)

type config struct {
	dbAddress    string
	pollInterval int
}

func initConfig() (*config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}
	dbAddress := os.Getenv("DB_ADDRESS")
	if dbAddress == "" {
		return nil, errors.New("DB address is not defined")
	}
	pollIntervalStr := os.Getenv("POLL_INTERVAL")
	pollInterval, err := strconv.Atoi(pollIntervalStr) // TODO: use time.Duration
	if err != nil {
		return nil, fmt.Errorf("poll interval is not int: %w", err)
	}

	return &config{dbAddress, pollInterval}, nil
}

func main() {
	conf, err := initConfig()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}
	fmt.Println("Config", conf)
	feedLink := "https://openai.com/news/rss.xml"
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedLink) // TODO: configure timeout (with context)
	if err != nil {
		slog.Error("can't parse feed", "feedLink", feedLink, "error_msg", err)
	} else {
		fmt.Printf("Feed type: %v, feed title: %v\n", feed.FeedType, feed.Title)
		if feed.Len() > 0 {
			fmt.Println(feed.Items[0].Title)
			fmt.Println(feed.Items[0].Description)
		}
	}
}
