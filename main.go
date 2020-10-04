package main

import (
	"github.com/screego/server/cmd"
	pmode "github.com/screego/server/config/mode"
)

var (
	version    = "unknown"
	commitHash = "unknown"
	mode       = pmode.Dev
)

func main() {
	pmode.Set(mode)
	cmd.Run(version, commitHash)
}
