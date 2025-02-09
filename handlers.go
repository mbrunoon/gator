package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mbrunoon/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {

	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
		if err != nil {
			return fmt.Errorf("error: %w", err)
		}

		return handler(s, cmd, user)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return errors.New("username is required")
	}

	username := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return errors.New("user not found")
	}

	err = s.config.SetUser(username)
	if err != nil {
		return errors.New("could not set current user")
	}

	fmt.Printf("New current username: %s\n", s.config.CurrentUserName)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return errors.New("invalid command")
	}

	newUserName := cmd.Args[0]

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		Name:      newUserName,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	if err != nil {
		return fmt.Errorf("could not create user %s: %w", newUserName, err)
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("Could not set user %s", user.Name)
	}

	fmt.Println("User created:")
	printUser(user)

	return nil
}

func handlerReset(s *state, _ command) error {

	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not clean user table: %w", err)
	}

	fmt.Println("User table is clean!")
	return nil
}

func handlerUsers(s *state, _ command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not select users: %w", err)
	}

	for _, u := range users {

		if s.config.CurrentUserName == u.Name {
			fmt.Println("-", u.Name, "(current)")
		} else {
			fmt.Println("-", u.Name)
		}

	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) < 1 || len(cmd.Args) > 2 {
		return fmt.Errorf("wrongs params")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("error, %w", err)
	}

	log.Printf("Collecting feeds every %s...", timeBetweenRequests)

	ticket := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticket.C {
		scrapFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		return errors.New("invalid args")
	}

	name := cmd.Args[0]
	url := cmd.Args[1]

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:     uuid.New(),
		UserID: user.ID,
		Name:   name,
		Url:    url,
	})

	if err != nil {
		return err
	}

	fmt.Println("Feed created:")
	fmt.Println("ID:", feed.ID)
	fmt.Println("UserID:", feed.UserID)
	fmt.Println("Name:", feed.Name)
	fmt.Println("Url:", feed.Url)

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})

	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	fmt.Println("Feed followed:")
	fmt.Println(feedFollow.UserName)
	fmt.Println(feedFollow.FeedName)

	return nil
}

func handlerFeeds(s *state, cmd command) error {

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error:", err)
	}

	if len(feeds) == 0 {
		fmt.Println("no feeds found")
		return nil
	}

	fmt.Println("Total feeds found", len(feeds))
	for _, feed := range feeds {
		user, err := s.db.GetUserById(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("error:", err)
		}

		printFeed(feed, user)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("insuficient args")
	}

	url := cmd.Args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("feed not found")
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})

	if err != nil {
		return fmt.Errorf("error on create feed follow")
	}

	fmt.Printf("User %s is now following %s \n", user.Name, feed.Url)
	fmt.Println(feedFollow)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("no feeds found")
		return nil
	}

	for _, feed := range feeds {
		fmt.Printf("- '%s' \n", feed.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("wrong args number")
	}

	url := cmd.Args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	feedFollow, err := s.db.GetFeedFollowByUser(context.Background(), database.GetFeedFollowByUserParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	err = s.db.DeleteFeedFollow(context.Background(), feedFollow.ID)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	return nil
}

// privates

func printUser(user database.User) {
	fmt.Printf(" * ID:		%v\n", user.ID)
	fmt.Printf(" * Name: 	%v\n", user.Name)
}

func printFeed(feed database.Feed, user database.User) {
	fmt.Println("ID:", feed.ID)
	fmt.Println("Name:", feed.Name)
	fmt.Println("URL:", feed.Url)
	fmt.Println("User:", user.Name)
}

func scrapFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Errorf("Couldn't no fetch feed")
		return
	}

	scrapFeed(s.db, feed)
}

func scrapFeed(db *database.Queries, feed database.Feed) {
	feedData, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Errorf("error on fetch fedd: %w", err)
		return
	}

	_, err = db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		fmt.Errorf("couldn't update feed: %w", err)
	}

	for _, item := range feedData.Channel.Item {
		fmt.Printf("Found post: %s\n", item.Title)
	}

	fmt.Printf("%v news post found at %s", len(feedData.Channel.Item), feed.Name)
}
