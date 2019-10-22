package main

import (
	"github.com/kui-shell/kask/kui"
)

// version information that will come from goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	kui.Start(version, commit, date)
}
