package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/screego/server/logger"
	"github.com/urfave/cli"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
)

var hashCmd = cli.Command{
	Name: "hash",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "name"},
		&cli.StringFlag{Name: "pass"},
	},
	Action: func(ctx *cli.Context) {
		logger.Init(zerolog.ErrorLevel)
		name := ctx.String("name")
		pass := []byte(ctx.String("pass"))
		if name == "" {
			log.Fatal().Msg("--name must be set")
		}

		if len(pass) == 0 {
			var err error
			_, _ = fmt.Fprint(os.Stderr, "Enter Password: ")
			pass, err = terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				log.Fatal().Err(err).Msg("could not read stdin")
			}
			_, _ = fmt.Fprintln(os.Stderr, "")
		}
		hashedPw, err := bcrypt.GenerateFromPassword(pass, 12)
		if err != nil {
			log.Fatal().Err(err).Msg("could not generate password")
		}

		fmt.Printf("%s:%s", name, string(hashedPw))
		fmt.Println("")
	},
}
