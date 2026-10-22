package main

import (
	"os"

	"github.com/kakeetopius/sesame/cmd"
)

func main() {
	err := cmd.Command().Execute()
	if err != nil {
		os.Exit(1)
	}
}
