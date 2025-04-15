package main

import (
	"fmt"
	"os"

	"github.com/s42yt/thighpads/pkg/app"
)

const (
	appVersion = "1.0.0"
)

func main() {

	fmt.Printf("ThighPads v%s - A Snippet Manager\n", appVersion)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
