package main
import (
	"context"
	"fmt"
	"strings"
	"time"
	"github.com/WagnerJust/go-gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Command struct {
	Name string
	Args []string
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
