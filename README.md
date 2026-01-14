### GO-GATOR Project

## Requirements
  - Postgres
  - Golang

## Installation

# Method: From Source Code
  1. Clone repository
  2. Run `go build` from project root directory
  3. Add `go-gator` to your path

# Method: Go Install
  1. Run `go install github.com/WagnerJust/go-gator@latest`

## Post-Installation Instructions
  1. Create a `.gatorconfig.json` file in your *HOME DIRECTORY*
  2. Create an empty postgres database. You may use any port and any database name.
  3. Fill the `.gatorconfig.json` with contents matching this template:
  ```json
  {
    "db_url": "postgres://justin@localhost:5432/gogator?sslmode=disable",
  }
  ```
  4. Register a user with the command `go-gator register <username>`
  
## Commands Available
1. `login <username>` - Login as an existing user
2. `print` - Print the current configuration
3. `register <username>` - Register a new user
4. `reset` - Reset the database
5. `users` - Get all users
6. `addfeed <name> <url>` - Add a new feed (requires login)
7. `feeds` - Get all feeds
8. `follow <feed_url>` - Follow a feed (requires login)
9. `following` - Get feeds you are following (requires login)
10. `unfollow <feed_url>` - Unfollow a feed (requires login)
11. `agg <time_between_reqs>` - Scrape feeds continuously
12. `aggone` - Scrape feeds once
13. `browse [limit]` - Browse posts from followed feeds (requires login)
