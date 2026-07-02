package soargs

import (
	"errors"
	"net"
	"os"
	"path/filepath"
)

type Server struct {
	name       string
	socketPath string
	cacheDir   string
	listener   net.Listener
}

// Startet einen Server mit dem Namen `name`.
//
// Der Name bestimmt den Pfad des Sockets: `~/.cache/soargs/NAME.socket`.
//
// Es wird versucht einen eventuell bereits vorhandenen Socket zu löschen, falls
// an diesem nicht bereits ein Server lauscht.
//
// Danach wird ein neuer Socket angelegt.
// 
// Falls der Socket oder das Verzeichnis für den Socket auf Grund von fehlenden
// Permissions nicht angegelegt werden kann, wird ein "denied"-Fehler geliefert.
//
// Falls der Socket bereits existiert und darauf bereits ein Server läuft,
// wird ein "exists"-Fehler geliefert.
//
// Alle andere Fehler werden als "os"-Fehler zurückgegeben.
func StartServer(name string) (*Server, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = "/tmp"
	}

	cacheDir = filepath.Join(cacheDir, "soargs")

	if err := os.MkdirAll(cacheDir, 0775); err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil, deniedError("Fehler beim Erzeugen des Cache-Verzeichnisses: %w", err)
		} else {
			return nil, osError("Fehler beim Erzeugen des Cache-Verzeichnisses: %w", err)
		}
	}

	socketPath := filepath.Join(cacheDir, name+".socket")

	if err := deleteOldSocket(socketPath); err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, osError("Fehler beim Horchen am Socket: %w", err)
	}
	
	return &Server{
		name:       name,
		cacheDir:   cacheDir,
		socketPath: socketPath,
		listener:   listener,
	}, nil

}

// Löscht einen eventuell bereits vorhandenen Socket.
func deleteOldSocket(path string) error {
	info, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		// Socket existiert nicht -> neuer kann erzeugt werden
		return nil
	}

	if info.Mode().Type() != os.ModeSocket {
		return existsError("Socket-Pfad existiert bereits, ist jedoch KEIN Socket: %s", path)
	}

	conn, err := net.Dial("unix", path)
	if err != nil {
		if err := os.Remove(path); err != nil {
			if errors.Is(err, os.ErrPermission) {
				return deniedError("Fehler beim Löschen des alten Sockets: %w", err)
			}
			return osError("Fehler beim Löschen des alten Sockets: %w", err)
		}
		return nil
	}

	conn.Close()

	return existsError("Server läuft bereits")
}

// Liefert den Namen des Servers.
func (s *Server) Name() string {
	return s.name
}

// Liefert den Pfad zum Socket.
func (s *Server) SocketPath() string {
	return s.socketPath
}

// Liefert das Cache-Verzeichnis, in welchem sich der alle Sockets befinden.
//
// Das ist normalerweise ~/.cache/soargs
func (s *Server) CacheDir() string {
	return s.cacheDir
}

// Wartet auf dem Socket auf eingehende Verbindungen und
// liefert ein [Client]-Objekt, welches zur Kommunikation mit dem Client dient.
//
// Falls die Verbindung nicht klappt, wird ein "connection"-Fehler geliefert.
func (s *Server) WaitForClient() (*Client, error) {
	conn, err := s.listener.Accept()
	if err != nil {
		return nil, connectionError("Fehler beim Verbinden mit einem neuen Client: %w", err)
	}
	return &Client{conn: conn}, nil
}

// Wie [WaitForClient], nur wird ein entsprechender Channel geliefert.
//
// Der Channel liefert [ClientResult]-Objekte, welche einfach die Rückgabewerte
// von [WaitForClient] enthält. Also den [Client] ([ClientResult.Client]) oder einen Fehler ([ClientResult.Err]).
func (s *Server) ClientChannel() <-chan ClientResult {
	ch := make(chan ClientResult)
	go func() {
		for {
			c, err := s.WaitForClient()
			ch <- ClientResult{
				Client: c,
				Err:    err,
			}
		}
	}()
	return ch
}

// Schließt und löscht den Socket.
func (s *Server) Close() {
	s.listener.Close()
	os.Remove(s.socketPath)
}
