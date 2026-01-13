package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/WagnerJust/go-gator/internal/config"
	"github.com/WagnerJust/go-gator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	Config *config.Config
	Db *database.Queries
}
type commands struct {
	CmdRegister map[string]func(*state, Command) error
}

func (c *commands) run (s *state, cmd Command) error {
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

func (c *commands) register (name string, f func(*state, Command) error) {
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
		CmdRegister: make(map[string]func(*state, Command) error),
	}

	commands.register("login", handlerLogin)
	commands.register("print", handlerPrintConfig)
	commands.register("register", handlerRegisterUser)
	commands.register("reset", handlerResetDatabase)
	commands.register("users", handlerGetAllUsers)
	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("feeds", handlerGetAllFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollowFeed))
	commands.register("following", middlewareLoggedIn(handlerGetFollowing))
	commands.register("unfollow", middlewareLoggedIn(handlerUnfollowFeed))
	commands.register("agg", handlerScrapeFeeds)
	commands.register("aggone", handlerAggOne)
	commands.register("browse", middlewareLoggedIn(handlerBrowsePosts))

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Error: you must provide a command")
		os.Exit(1)
	}
	userCommand := Command{
		Name: args[1],
		Args: args[2:],
	}

	err = commands.run(appState, userCommand)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
