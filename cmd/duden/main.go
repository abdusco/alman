package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"

	"github.com/alecthomas/kong"

	"github.com/abdusco/alman/pkg/dwds"
)

type cliArgs struct {
	Word  string `arg:"" help:"Word to search"`
	JSON  bool   `help:"Output as JSON"`
	Debug bool   `help:"Write debugging info" default:"false"`
}

func (a cliArgs) Run() error {
	du := dwds.New()
	entry, err := du.Find(context.Background(), a.Word)
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

	if args.Debug {
		log.SetLevel(log.DebugLevel)
	}

	if err := cliCtx.Run(); err != nil {
		log.Fatal("exit with error", "error", err)
	}
}
