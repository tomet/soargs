package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tomet/ansi"
	"github.com/tomet/soargs"
	"github.com/tomet/terror"
)

func main() {
	program := filepath.Base(os.Args[0])
	terror.Program = program
	args := os.Args

	if program == "soargs-client" {
		if len(os.Args) < 2 {
			terror.Syntax.Fail(
				"Bitte soargs-client über einen Symlink (Basename == Servername) aufrufen\n" +
				 "oder den Namen des Servers als erstes Argument angegeben!",
			)
		}
		program = os.Args[1]
		args = append([]string{os.Args[0]}, os.Args[2:]...)
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = "/tmp"
	}
	socketPath := filepath.Join(cacheDir, "soargs", program+".socket")

	info, err := os.Stat(socketPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			terror.Connection.Fail("%s-Server läuft nicht (Socket existiert nicht)!", program)
		}
	} else if info.Mode().Type() != os.ModeSocket {
		terror.Connection.Fail("Pfad ist KEIN Unix-Domain-Socket: %s", socketPath)
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		var sysErr *os.SyscallError
		if errors.As(errors.Unwrap(err), &sysErr) {
			if sysErr.Syscall == "connect" {
				terror.Connection.Fail("%s-Server läuft nicht (connection refused)!", program)
			}
		}
		terror.Connection.Fail("Fehler beim Verbinden mit dem Server: %s", err)
	}

	cmd := &soargs.Cmd{
		Args:    args,
		Lines:   ansi.Lines(),
		Columns: ansi.Columns(),
		IsAtty:  ansi.IsATTY(),
		Env:     buildEnvMap(),
	}

	io.WriteString(conn, encodeCmd(cmd))

	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString(0)
		if err != nil {
			if errors.Is(err, io.EOF) {
				os.Exit(0)
			}
			terror.Connection.Fail("Fehler beim Lesen der Server-Antwort: %s", err)
		}

		if len(msg) < 2 {
			terror.Connection.Fail("Ungültige Server-Antwort: %q", msg)
		}

		prefix := msg[0]
		msg = msg[1 : len(msg)-1]

		switch prefix {
		case '1':
			fmt.Fprint(os.Stdout, msg)
		case '2':
			fmt.Fprint(os.Stderr, msg)
		case 'x':
			ec, _ := strconv.Atoi(msg)
			os.Exit(ec)
		default:
			terror.Connection.Fail("Server-Antwort mit ungültigem Prefix empfangen: %c", prefix)
		}
	}

}

func buildEnvMap() map[string]string {
	env := os.Environ()
	envMap := make(map[string]string, len(env))
	for _, pair := range env {
		name, value, found := strings.Cut(pair, "=")
		if found {
			envMap[name] = value
		}
	}
	return envMap
}

func encodeCmd(c *soargs.Cmd) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "lines=%d\000", c.Lines)
	fmt.Fprintf(&sb, "columns=%d\000", c.Columns)
	fmt.Fprintf(&sb, "isatty=%v\000", c.IsAtty)
	for name, value := range c.Env {
		fmt.Fprintf(&sb, "env=%s=%s\000", name, value)
	}
	for _, arg := range c.Args {
		fmt.Fprintf(&sb, "arg=%s\000", arg)
	}
	sb.WriteByte(1)
	return sb.String()
}
