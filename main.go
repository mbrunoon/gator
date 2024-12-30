package main

import (
	"fmt"
	"log"

	"github.com/mbrunoon/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.SetUser("brn")
	if err != nil {
		log.Fatal(err)
	}

	cfg, err = config.Read()
	if err != nil {
		log.Fatal(cfg)
	}

	fmt.Println(cfg)
}
