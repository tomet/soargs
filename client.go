package soargs

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

var (
	PingCmd = &Cmd{}
)

// Repräsentiert die Verbindung zum Client.
type Client struct {
	conn net.Conn
}

// Sollte als erstes aufgerufen werden. Es wird das Kommando ([Cmd])
// vom Client gelesen.
//
// Es kann auch einfach ein [PingCmd] zurückgegeben werden, falls der
// Client nur prüfen wollte, ob der Server läuft ([Cmd.IsPing]). In diesem
// Fall wird die Verbindung einfach wieder geschlossen und es kann auf
// den nächsten Client gewartet werden.
func (c *Client) ReadCmd() (*Cmd, error) {
	reader := bufio.NewReader(c.conn)

	data, err := reader.ReadString(1)
	if err != nil {
		if errors.Is(err, io.EOF) {
			if data != "" {
				c.conn.Close()
				return PingCmd, nil
			}
			return nil, parseError("Ungültiges Kommando vom Client empfangen (endet nicht mit einem 1-Byte)")
		}

		return nil, connectionError("Fehler beim Empfangen des Kommandos vom Client: %s", err)
	}

	if len(data) < 1 {
		c.conn.Close()
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
			return nil, parseError("Parameter ist kein KEY=VALUE-Paar: %q", param)
		}
		switch name {
		case "lines":
			v, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return nil, parseError("Ungültiger int-Wert für lines-Parameter: %w", err)
			}
			if v > 0 && v < 1000000 {
				cmd.Lines = int(v)
			} else {
				return nil, parseError("Ungüliger Wert für lines-Parameter: %d", v)
			}
		case "columns":
			v, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return nil, parseError("Ungültiger int-Wert für columns-Parameter: %w", err)
			}
			if v > 5 && v < 1000 {
				cmd.Columns = int(v)
			} else {
				return nil, parseError("Ungüliger Wert für columns-Parameter: %d", v)
			}
		case "isatty":
			cmd.IsAtty = value == "true"
		case "env":
			name, value, found := strings.Cut(value, "=")
			if !found {
				return nil, parseError("Ungültiger env-Parameter: Wert ist KEIN Key-Value-Paar: %q", value)
			}
			if name != "" {
				cmd.Env[name] = value
			}
		case "arg":
			cmd.Args = append(cmd.Args, value)
		default:
			return nil, parseError("Unbekannter Parameter-Name: %q", name)
		}
	}

	return cmd, nil
}

// Wie [fmt.Println] auf der Client-Seite.
func (c *Client) Println(args ...any) {
	c.println('1', args...)
}

// Wie [fmt.Printf] auf der Client-Seite.
func (c *Client) Printf(format string, args ...any) {
	c.printf('1', format, args...)
}

// Wie [fmt.Fprintln](os.Stderr, ...) auf der Client-Seite.
func (c *Client) Eprintln(args ...any) {
	c.println('2', args...)
}

// Wie [fmt.Fprintf](os.Stderr, ...) auf der Client-Seite.
func (c *Client) Eprintf(format string, args ...any) {
	c.printf('2', format, args...)
}

// Sendet dem Client den Befehl zum Beenden mit dem angegeben `exitcode`.
func (c *Client) Exit(exitcode int) {
	fmt.Fprintf(c.conn, "x%d\x00", exitcode)
}

func (c *Client) println(prefix rune, args ...any) {
	c.writeRune(prefix)
	fmt.Fprintln(c.conn, args...)
	c.writeRune('\x00')
}

func (c *Client) printf(prefix rune, format string, args ...any) {
	c.writeRune(prefix)
	fmt.Fprintf(c.conn, format, args...)
	c.writeRune('\x00')
}

func (c *Client) writeRune(r rune) {
	fmt.Fprintf(c.conn, "%c", r)
}

// Schließt die Verbindung zum Client.
func (c *Client) Close() error {
	return c.conn.Close()
}
