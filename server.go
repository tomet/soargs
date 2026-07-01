package soargs

import (
	"errors"
	"net"
	"os"
	"path/filepath"

	"github.com/tomet/terror"
)

type Server struct {
	program    string
	socketPath string
	cacheDir   string
	listener   net.Listener
}

func StartServer(program string) (*Server, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = "/tmp"
	}

	cacheDir = filepath.Join(cacheDir, "soargs")

	if err := os.MkdirAll(cacheDir, 0775); err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil, terror.Denied.Errorf("Fehler beim Erzeugen des Cache-Verzeichnisses: %w", err)
		} else {
			return nil, terror.Os.Errorf("Fehler beim Erzeugen des Cache-Verzeichnisses: %w", err)
		}
	}

	socketPath := filepath.Join(cacheDir, program+".socket")

	if err := deleteOldSocket(socketPath); err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}
	return &Server{
		program:    program,
		cacheDir:   cacheDir,
		socketPath: socketPath,
		listener:   listener,
	}, nil

}

func deleteOldSocket(path string) error {
	info, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		// Socket existiert nicht -> neuer kann erzeugt werden
		return nil
	}

	if info.Mode().Type() != os.ModeSocket {
		return terror.Exists.Errorf("Socket-Pfad existiert bereits, ist jedoch KEIN Socket: %s", path)
	}

	conn, err := net.Dial("unix", path)
	if err != nil {
		if err := os.Remove(path); err != nil {
			if errors.Is(err, os.ErrPermission) {
				return terror.Denied.Errorf("Fehler beim Löschen des alten Sockets: %w", err)
			}
			return terror.Os.Errorf("Fehler beim Löschen des alten Sockets: %w", err)
		}
		return nil
	}

	conn.Close()

	return terror.Exists.Errorf("Server läuft bereits")
}

func (s *Server) Program() string {
	return s.program
}

func (s *Server) SocketPath() string {
	return s.socketPath
}

func (s *Server) CacheDir() string {
	return s.cacheDir
}

func (s *Server) WaitForClient() (*Client, error) {
	conn, err := s.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}
