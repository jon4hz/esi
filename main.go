package main

import (
	"io"
	"os"

	"github.com/jon4hz/esi/cmd"

	stdlog "log"
)

func main() {
	stdlog.Default().SetOutput(io.Discard)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
