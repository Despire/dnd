package main

import (
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/Despire/dnd/restrictions"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// cache the user info.
	if _, err := user.Current(); err != nil {
		return fmt.Errorf("failed to retrieve current user: %w", err)
	}

	if err := restrictions.CreateConfigDir(); err != nil {
		return fmt.Errorf("failed to create directory for storing configuration: %w", err)
	}

	args := os.Args[1:]
	if len(args) < 1 {
		help(os.Stdout)
		return nil
	}

	switch args[0] {
	case "add":
		add(os.Stdout, os.Stdin, args[1:]...)
	case "del":
		del(os.Stdout, args[1:]...)
	case "print":
		print(os.Stdout)
	case "types":
		types(os.Stdout)
	case "commit":
		commit(os.Stdout, os.Stdin)
	default:
		help(os.Stdout)
	}

	return nil
}
