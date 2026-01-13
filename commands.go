package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/WagnerJust/go-gator/internal/database"
	"github.com/WagnerJust/go-gator/internal/rss"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Command struct {
	Name string
	Args []string
}

func stringPtrToNullString(s *string) sql.NullString {
    if s == nil {
        return sql.NullString{Valid: false}
    }
    return sql.NullString{String: *s, Valid: true}
}


func middlewareLoggedIn(handler func(s *state, cmd Command, user database.User) error) func(*state, Command) error {
	return func(s *state, cmd Command) error {
		user, err := s.Db.GetUserByName(context.Background(), s.Config.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func handlerLogin(s *state, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: login <username>")
	}
	_, err := s.Db.GetUserByName(context.Background(),cmd.Args[0])
	if err != nil {
		return err
	}
	err = s.Config.SetUser(cmd.Args[0])
	if err != nil {
		return err
	}
	fmt.Println("User has been set")
	return nil
}

func handlerPrintConfig(s *state, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: config")
	}
	fmt.Println(s.Config.String())
	return nil
}

func handlerRegisterUser (s *state, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: register <username>")
	}

	userParams := database.CreateUserParams{
		ID: uuid.New(),
		Name: cmd.Args[0],
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user, err := s.Db.CreateUser(context.Background(), userParams )
	if err != nil {
		return err
	}
	s.Config.SetUser(user.Name)
	fmt.Println("User created successfully!")
	fmt.Printf("User: %+v\n", user)
	return nil
}

func handlerResetDatabase(s *state, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: reset")
	}

	err := s.Db.DeleteAllUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("All users have been cleared from the database")
	return nil
}

func handlerGetAllUsers(s *state, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: users")
	}

	users, err := s.Db.GetAllUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		phrase := "* " + strings.ToLower(user.Name)
		if strings.EqualFold(user.Name, s.Config.CurrentUserName) {
			phrase += " (current)"
		}
		fmt.Println(phrase)
	}
	return nil
}

func handlerAddFeed (s *state, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}

	feedParams := database.CreateFeedParams{
		ID: uuid.New(),
		Name: cmd.Args[0],
		Url: cmd.Args[1],
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
	}
	feed, err := s.Db.CreateFeed(context.Background(),feedParams)
	if err != nil {
		return err
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	}

	feedFollow, err := s.Db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return err
	}

	fmt.Printf("Feed Follow: %+v\n", feedFollow)
	return nil
}

func handlerGetAllFeeds (s *state, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: feeds")
	}

	feeds, err := s.Db.GetAllFeedsWithUsers(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("name: %s\n\turl: %s\n\tuser: %s\n", feed.Name, feed.Url, feed.UserName)
	}
	return nil
}

func handlerFollowFeed (s *state, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: follow <url>")
	}

	feed, err := s.Db.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	}

	feedFollow, err := s.Db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return err
	}

	fmt.Printf("Feed Follow: %+v\n", feedFollow)
	return nil
}

func handlerGetFollowing (s *state, cmd Command, user database.User) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: following")
	}

	feedFollows, err := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}
	if len(feedFollows) == 0 {
		fmt.Println("You are not following any feeds")
		return nil
	}
	fmt.Println("You are following these feeds:")
	for _, feed := range feedFollows {
		fmt.Printf("\t- %s\n", feed.FeedName)
	}
	return nil
}

func handlerUnfollowFeed(s *state, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return  fmt.Errorf("usage: unfollow <url>")
	}
	feed, err := s.Db.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}
	params := database.DeleteFeedFollowByUserParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.Db.DeleteFeedFollowByUser(context.Background(),params)
	if err != nil {
		return err
	}
	return nil
}


func handlerScrapeFeeds ( s *state, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: agg <duration_string>");
	}
	timeBetweenReqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}
	fmt.Println("Collecting feeds every ", timeBetweenReqs.String())
	ticker := time.NewTicker(timeBetweenReqs)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		fmt.Println("Checking...")
		feed, err := s.Db.GetNextFeedToFetch(context.Background())
		if err != nil {
			return err
		}
		err = s.Db.MarkFeedFetched(context.Background(), feed.ID)
		if err != nil {
			return err
		}

		fetchedFeed, err := rss.FetchFeed(context.Background(), feed.Url)
		if err != nil {
			return err
		}
		postsCreated := 0
		for _, item := range fetchedFeed.Channel.Item {

			pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
			if err != nil {
				fmt.Printf("Could not parse publication date for '%s': %v\n", item.Title, err)
				continue
			}
			postParams := database.CreatePostParams{
				ID: uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				PublishedAt: pubDate,
				Title: item.Title,
				Url: item.Link,
				Description: stringPtrToNullString(item.Description),
				FeedID: feed.ID,
			}
			_, err = s.Db.CreatePost(context.Background(), postParams)
			if err != nil {
				if strings.Contains(err.Error(), "duplicate key value") {
					continue
				}
				fmt.Printf("Error creating post: %v\n", err)
				return err
			}
			postsCreated++
		}
		fmt.Printf("Feed %s collected, %v new posts found\n", feed.Name, postsCreated)
	}
}


func handlerBrowsePosts( s *state, cmd Command, user database.User) error {
	if len(cmd.Args) > 1 {
		return fmt.Errorf("usage: browse <optional_limit_integer>")
	}
	limit := 2
	var err error
	if len(cmd.Args) == 1 {
		limit, err = strconv.Atoi(cmd.Args[0])
		if err != nil {
			return err
		}
	}

	params := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: int32(limit),
	}
	posts, err := s.Db.GetPostsForUser(context.Background(), params)
	if err != nil {
		return err
	}
	for _, post := range posts {
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("URL: %s\n", post.Url)
		if post.Description.Valid {
			fmt.Printf("Description: %s\n", post.Description.String)
		}
		fmt.Printf("Published: %s\n", post.PublishedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("================================================================================")
	}
	return nil
}

func handlerAggOne(s *state, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: aggone <feed_url>")
	}

	feed, err := s.Db.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Fetching feed: %s\n", feed.Name)
	fmt.Printf("URL: %s\n", feed.Url)
	fmt.Printf("Feed ID: %s\n", feed.ID)

	err = s.Db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return err
	}

	fetchedFeed, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	fmt.Printf("\nFetched RSS Feed: %s\n", fetchedFeed.Channel.Title)
	fmt.Printf("Number of items: %d\n\n", len(fetchedFeed.Channel.Item))

	postsCreated := 0
	duplicates := 0
	parseErrors := 0

	for i, item := range fetchedFeed.Channel.Item {
		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			// Try alternative date formats
			pubDate, err = time.Parse(time.RFC1123, item.PubDate)
			if err != nil {
				// Try RFC3339
				pubDate, err = time.Parse(time.RFC3339, item.PubDate)
				if err != nil {
					fmt.Printf("[%d] Could not parse date '%s' for: %s\n", i+1, item.PubDate, item.Title)
					parseErrors++
					continue
				}
			}
		}

		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			PublishedAt: pubDate,
			Title:       item.Title,
			Url: item.Link,
			Description: stringPtrToNullString(item.Description),
			FeedID:      feed.ID,
		}
		_, err = s.Db.CreatePost(context.Background(), postParams)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value") {
				duplicates++
				continue
			}
			fmt.Printf("Error creating post: %v\n", err)
			return err
		}
		postsCreated++
		if postsCreated <= 3 {
			fmt.Printf("✓ Created: %s (published: %s)\n", item.Title, pubDate.Format("2006-01-02"))
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Feed: %s\n", feed.Name)
	fmt.Printf("Total items: %d\n", len(fetchedFeed.Channel.Item))
	fmt.Printf("New posts: %d\n", postsCreated)
	fmt.Printf("Duplicates: %d\n", duplicates)
	fmt.Printf("Parse errors: %d\n", parseErrors)

	return nil
}
