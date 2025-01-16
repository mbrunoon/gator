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

func printUser(user database.User) {
	fmt.Printf(" * ID:		%v\n", user.ID)
	fmt.Printf(" * Name: 	%v\n", user.Name)
}
