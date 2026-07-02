package main

import (
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/tomet/soargs"

	"github.com/tomet/terror"
)

func start(program string) {
	server, err := soargs.StartServer(program)
	if err != nil {
		terror.FailWithError(err)
	}

	fmt.Printf("Program: %s\n", server.Program())
	fmt.Printf("Socket:  %s\n", server.SocketPath())

	clientCh := server.ClientChannel()
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case sig := <-signalCh:
			fmt.Fprintf(os.Stderr, "Abbruch durch %s\n", sig)
			server.Close()
			if sig == os.Interrupt {
				terror.SigInt.Exit()
			}
			terror.SigTerm.Exit()
			return
		case result := <-clientCh:
			if result.Err != nil {
				terror.FailWithError(err)
			}

			client := result.Client
			cmd, err := client.ReadCmd()
			if err != nil {
				terror.FailWithError(err)
			}

			if cmd.IsPing() {
				fmt.Println("Ping command received.")
			} else {
				fmt.Printf("Command received:\n")
				fmt.Printf(" lines=%d columns=%d isatty=%v\n", cmd.Lines, cmd.Columns, cmd.IsAtty)
				names := make([]string, 0, len(cmd.Env))
				for name := range cmd.Env {
					names = append(names, name)
				}
				slices.Sort(names)
				for _, name := range names {
					fmt.Printf(" env %s = %q\n", name, cmd.Env[name])
				}
				for i, arg := range cmd.Args {
					fmt.Printf(" arg[%d]=%q\n", i, arg)
				}
				client.Printf("STDOUT: lines=%d columns=%d isatty=%v\n", cmd.Lines, cmd.Columns, cmd.IsAtty)
				client.Eprintln("STDERR")
				client.Exit(terror.False.Code())
			}

			client.Close()
			fmt.Println("Client disconnected.")
		}
	}
}
