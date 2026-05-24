package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Lillevang/agent-init/internal/cli"
)

var (
	version   = "dev"
	commit    = "dev"
	buildDate = "unknown"
)

func main() {
	app := cli.New(os.Stdout, os.Stderr, cli.Version{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
	})
	if err := app.Run(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "agent-init: %v\n", err)
		os.Exit(1)
	}
}
