package main

import (
	"github.com/tomet/cmdr"
	"github.com/tomet/terror"
)

func main() {
	c := cmdr.NewWithHelp(
		cmdr.Cmd("start").Add(
			cmdr.Arg("PROGRAM"),
		),
	)

	program := ""

	cmd, err := c.ParseOsArgsOrPrintHelp(func(vars cmdr.Variables) {
		program = vars.Value("PROGRAM")
	})

	if err != nil {
		terror.FailWithError(err)
	}

	switch cmd {
	case "start":
		start(program)
	}
}
