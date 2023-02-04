package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/abdusco/alman/pkg/duden"
)

type cliArgs struct {
	Word  string `arg:"" help:"Word to search"`
	JSON  bool   `help:"Output as JSON"`
	Debug bool   `help:"Write debugging info" default:"false"`
}

func (a cliArgs) Run() error {
	du := duden.NewDuden()
	entry, err := du.Find(a.Word)
	if err != nil {
		return err
	}

	if a.JSON {
		j, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal as json: %w", err)
		}
		fmt.Println(string(j))
		return nil
	}

	fmt.Println(entry.String())

	return nil
}

func main() {
	var args cliArgs
	cliCtx := kong.Parse(&args, kong.Name("duden"))

	log.Logger = log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.InfoLevel)
	if args.Debug {
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	}

	if err := cliCtx.Run(); err != nil {
		log.Fatal().Err(err)
	}
}
