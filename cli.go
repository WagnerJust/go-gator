package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/WagnerJust/go-gator/internal/config"
	"github.com/WagnerJust/go-gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	Config *config.Config
	Db *database.Queries
}

type command struct {
	Name string
	Args []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("username is required. username must be the only argument")
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

func handlerPrintConfig(s *state, cmd command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("no arguments expected")
	}
	fmt.Println(s.Config.String())
	return nil
}

func handlerRegisterUser (s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("username is required. username must be the only argument")
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
	fmt.Printf("User Data: ID=%v, Name=%s, CreatedAt=%v, UpdatedAt=%v\n",
		user.ID, user.Name, user.CreatedAt, user.UpdatedAt)
	return nil
}

func handleResetDatabase(s *state, cmd command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("reset must be run without arguments")
	}

	err := s.Db.DeleteAllUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("All users have been cleared from the database")
	return nil
}

func handleGetAllUsers(s *state, cmd command) error {

	if len(cmd.Args) != 0 {
		return fmt.Errorf("`users` must be run without arguments")
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

func handleAddFeed (s *state, cmd command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("`addFeed` should be called like: `addfeed {name} {url}`")
	}

	currentUser, err := s.Db.GetUserByName(context.Background(), s.Config.CurrentUserName)
	if err != nil {
		return err
	}

	feedParams := database.CreateFeedParams{
		ID: uuid.New(),
		Name: cmd.Args[0],
		Url: cmd.Args[1],
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: currentUser.ID,
	}
	entry, err := s.Db.CreateFeed(context.Background(),feedParams)
	if err != nil {
		return err
	}

	fmt.Println(entry)

	return nil
}

func handleGetAllFeeds (s *state, cmd command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("`feeds` does not take any arguments")
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


type commands struct {
	CmdRegister map[string]func(*state, command) error
}

func (c *commands) run (s *state, cmd command) error {
	handler, ok := c.CmdRegister[cmd.Name]
	if !ok {
		return fmt.Errorf("command not found: %s", cmd.Name)
	}
	err := handler(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *commands) register (name string, f func(*state, command) error) {
	c.CmdRegister[name] = f
}

func CliLoop () {
	appState := &state{
		Config: config.NewConfig(),
	}
	err := appState.Config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
	}

	db, err := sql.Open("postgres", appState.Config.DbUrl)
	if err != nil {
		fmt.Println("Error opening database:", err)
		os.Exit(1)
	}
	defer db.Close()
	appState.Db = database.New(db)

	commands := commands{
		CmdRegister: make(map[string]func(*state, command) error),
	}

	commands.register("login", handlerLogin)
	commands.register("print", handlerPrintConfig)
	commands.register("register", handlerRegisterUser)
	commands.register("reset", handleResetDatabase)
	commands.register("users", handleGetAllUsers)
	commands.register("addfeed", handleAddFeed)
	commands.register("feeds", handleGetAllFeeds)

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Error: you must provide a command")
		os.Exit(1)
	}
	userCommand := command{
		Name: args[1],
		Args: args[2:],
	}

	err = commands.run(appState, userCommand)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
