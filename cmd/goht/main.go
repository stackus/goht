package main

import (
	"os"

	"github.com/stackus/goht/cmd/goht/cmd"
)

func main() {
	run(cmd.Execute, os.Exit)
}

func run(run func() error, exit func(int)) {
	if err := run(); err != nil {
		exit(1)
	}
}
