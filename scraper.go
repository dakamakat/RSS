package main

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/dakamakat/RSS/internal/database"
	"github.com/google/uuid"
)

func startScraping(db *database.Queries, concurrency int, timeBetweenRequest time.Duration) {
	log.Printf("Scraping on %v goroutines every %s duration", concurrency, timeBetweenRequest)

	ticker := time.NewTicker(timeBetweenRequest)

	for ; ; <-ticker.C {

		feeds, err := db.GetNextFeedsToFetch(context.Background(), int32(concurrency))

		if err != nil {
			log.Printf("Error fetching feeds: %v", err)
			continue
		}

		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)

			go scrapeFeed(db, wg, feed)

		}

		wg.Wait()

	}
}

func scrapeFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()

	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)

	if err != nil {
		log.Printf("Error marking feed as fetched: %v", err)
		return
	}

	rssFeed, err := urlToFeed(feed.Url)

	if err != nil {
		log.Printf("Error fetching feed: %v", err)
		return
	}

	total := 0

	for _, item := range rssFeed.Channel.Item {
		description := sql.NullString{}

		date, err := time.Parse(time.RFC1123Z, item.PubDate)

		if err != nil {
			log.Printf("Couldn't parse date %v with error %v", item.PubDate, err)
			continue
		}

		if item.Description != "" {
			description.String = item.Description
			description.Valid = true
		}

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			PublishedAt: date,
			Title:       item.Title,
			Url:         item.Link,
			FeedID:      feed.ID,
			Description: description,
		})

		if err != nil {
			log.Printf("Falied to create post: %v", err)
		} else {
			total++
		}
	}

	log.Printf("Feed %s collected, %v posts found, new posts inserted %v", feed.Name, len(rssFeed.Channel.Item), total)
}
