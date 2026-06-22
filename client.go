package soargs

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type Client struct {
	conn net.Conn
}

func (c *Client) ReadCmd() (*Cmd, error) {
	bytes, err := io.ReadAll(c.conn)
	if err != nil {
		return nil, err
	}

	cmd := &Cmd{
		Lines:   40,
		Columns: 80,
		IsAtty:  false,
		Env:     make(map[string]string),
	}

	for _, param := range strings.Split(string(bytes), "\x00") {
		if param == "" {
			continue
		}

		name, value, found := strings.Cut(param, "=")
		if !found {
			return nil, fmt.Errorf("Parameter ist kein KEY=VALUE-Paar: %q", param)
		}
		switch name {
		case "lines":
			v, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("Ungültiger int-Wert für lines-Parameter: %w", err)
			}
			if v > 0 && v < 1000000 {
				cmd.Lines = int(v)
			} else {
				return nil, fmt.Errorf("Ungüliger Wert für lines-Parameter: %d", v)
			}
		case "columns":
			v, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("Ungültiger int-Wert für columns-Parameter: %w", err)
			}
			if v > 5 && v < 1000 {
				cmd.Columns = int(v)
			} else {
				return nil, fmt.Errorf("Ungüliger Wert für columns-Parameter: %d", v)
			}
		case "isatty":
			cmd.IsAtty = value == "true"
		case "env":
			name, value, found := strings.Cut(value, "=")
			if !found {
				return nil, fmt.Errorf("Ungültiger env-Parameter: Wert ist KEIN Key-Value-Paar: %q", value)
			}
			if name != "" {
				cmd.Env[name] = value
			}
		case "arg":
			cmd.Args = append(cmd.Args, value)
		default:
			return nil, fmt.Errorf("Unbekannter Parameter-Name: %q", name)
		}
	}

	return cmd, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
