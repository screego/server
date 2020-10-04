package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

func Run(version, commitHash string) {
	app := cli.App{
		Name:    "screego",
		Version: fmt.Sprintf("%s; screego/server@%s", version, commitHash),
		Commands: []cli.Command{
			serveCmd(version),
			hashCmd,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("app error")
	}
}
