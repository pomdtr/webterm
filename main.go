package main

import (
	"os"

	"github.com/pomdtr/wsh/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
