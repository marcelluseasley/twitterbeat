package main

import (
	"os"

	"github.com/marcelluseasley/twitterbeat/cmd"

	_ "github.com/marcelluseasley/twitterbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
