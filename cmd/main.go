package main

import (
	"fmt"
	"os"

	"github.com/rhythin/sever-management/internal"
)

func main() {
	cfg, err := internal.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded config: %+v\n", cfg)
	// TODO: Initialize DI, DB, HTTP server, etc.
}
