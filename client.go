package soargs

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/tomet/terror"
)

var (
	PingCmd = &Cmd{}
)

type Client struct {
	conn net.Conn
}

func (c *Client) ReadCmd() (*Cmd, error) {
	reader := bufio.NewReader(c.conn)

	data, err := reader.ReadString(1)
	if err != nil {
		if errors.Is(err, io.EOF) {
			if data != "" {
				return PingCmd, nil
			}
			return nil, newParseError("Ungültiges Kommando empfangen (endet nicht mit einem 1-Byte)")
		}

		return nil, terror.Connection.Errorf("Fehler beim Empfangen des Kommandos: %w", err)
	}

	if len(data) < 1 {
		return PingCmd, nil
	}

	cmd := &Cmd{
		Lines:   240,
		Columns: 80,
		IsAtty:  false,
		Env:     make(map[string]string),
	}

	for _, param := range strings.Split(data[:len(data)-1], "\x00") {
		if param == "" {
			continue
		}

		name, value, found := strings.Cut(param, "=")
		if !found {
			return nil, newParseError("Parameter ist kein KEY=VALUE-Paar: %q", param)
		}
		switch name {
		case "lines":
			v, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return nil, newParseError("Ungültiger int-Wert für lines-Parameter: %w", err)
			}
			if v > 0 && v < 1000000 {
				cmd.Lines = int(v)
			} else {
				return nil, newParseError("Ungüliger Wert für lines-Parameter: %d", v)
			}
		case "columns":
			v, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return nil, newParseError("Ungültiger int-Wert für columns-Parameter: %w", err)
			}
			if v > 5 && v < 1000 {
				cmd.Columns = int(v)
			} else {
				return nil, newParseError("Ungüliger Wert für columns-Parameter: %d", v)
			}
		case "isatty":
			cmd.IsAtty = value == "true"
		case "env":
			name, value, found := strings.Cut(value, "=")
			if !found {
				return nil, newParseError("Ungültiger env-Parameter: Wert ist KEIN Key-Value-Paar: %q", value)
			}
			if name != "" {
				cmd.Env[name] = value
			}
		case "arg":
			cmd.Args = append(cmd.Args, value)
		default:
			return nil, newParseError("Unbekannter Parameter-Name: %q", name)
		}
	}

	return cmd, nil
}

func (c *Client) Println(args ...any) {
	c.println('1', args...)
}

func (c *Client) Printf(format string, args ...any) {
	c.printf('1', format, args...)
}

func (c *Client) EPrintln(args ...any) {
	c.println('2', args...)
}

func (c *Client) EPrintf(format string, args ...any) {
	c.printf('2', format, args...)
}

func (c *Client) Exit(exitcode int) {
	fmt.Fprintf(c.conn, "x%d\x01", exitcode)
}

func (c *Client) println(prefix rune, args ...any) {
	c.writeRune(prefix)
	fmt.Fprintln(c.conn, args...)
	c.writeRune('\x01')
}

func (c *Client) printf(prefix rune, format string, args ...any) {
	c.writeRune(prefix)
	fmt.Fprintf(c.conn, format, args...)
	c.writeRune('\x01')
}

func (c *Client) writeRune(r rune) {
	fmt.Fprintf(c.conn, "%c", r)
}

func (c *Client) Close() error {
	return c.conn.Close()
}
