package main

import (
	"fmt"
	"os"

	"github.com/hossein1376/s3manager/cmd/s3manager/command"
)

func main() {
	err := command.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
