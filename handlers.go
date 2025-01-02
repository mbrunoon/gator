package main

import (
	"errors"
	"fmt"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return errors.New("username is required")
	}

	username := cmd.Args[0]

	err := s.config.SetUser(username)
	if err != nil {
		return fmt.Errorf("couldn't set user: %w", err)
	}

	fmt.Printf("New current username: %s\n", s.config.CurrentUserName)

	return nil
}
