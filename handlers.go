package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mbrunoon/gator/internal/database"
)

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

func handlerAgg(_ *state, _ command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}

	fmt.Printf("Feed: \n", feed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.Args) != 2 {
		return errors.New("invalid args")
	}

	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
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

func handlerFollow(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("insuficient args")
	}

	url := cmd.Args[0]

	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("error on get user")
	}

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

func handlerFollowing(s *state, cmd command) error {
	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("no feeds found")
		return nil
	}

	for _, feed := range feeds {
		fmt.Println(feed.FeedName)
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
