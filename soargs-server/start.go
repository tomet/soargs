package main

import (
	"fmt"

	"github.com/tomet/soargs"

	"github.com/tomet/terror"
)

func start(program string) {
	serv, err := soargs.StartServer(program)
	if err != nil {
		terror.Fail(err)
	}

	fmt.Printf("Program: %s\n", serv.Program())
	fmt.Printf("Socket:  %s\n", serv.SocketPath())

	for {
		client, err := serv.WaitForClient()
		if err != nil {
			terror.Fail(err)
		}

		{
			defer func() {
				client.Close()
				fmt.Println("Client disconnected.")
			}()
			
			fmt.Println("Client connected.")

			cmd, err := client.ReadCmd()
			if err != nil {
				terror.Fail(err)
			}

			if cmd.IsPing() {
				fmt.Println("Ping command received.")
			} else {
				fmt.Printf("Command received:\n")
				tty := "no tty"
				if cmd.IsAtty {
					tty = "tty"
				}
				fmt.Printf(" %d lines, %d columns, %s\n", cmd.Lines, cmd.Columns, tty)
				client.Printf("STDOUT: %d lines, %d columns, %s\n", cmd.Lines, cmd.Columns, tty)
			}
		}
	}
}
