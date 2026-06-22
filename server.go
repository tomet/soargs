package soargs

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
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
		return nil, fmt.Errorf("Fehler beim Erzeugen des Konfigurations-Verzeichnisses: %w", err)
	}

	socketPath := filepath.Join(cacheDir, program+".socket")

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
