package soargs

// Das vom Client empfangene Kommando samt aller
// benötigten Umgebungseinstellungen des Clients.
type Cmd struct {
	// wieviele Zeilen hat das Terminal-Fenster des Clients?
	Lines int
	// wieviele Spalten hat das Terminal-Fenster des Clients?
	Columns int
	// wird der Client in einem Terminal ausgeführt?
	IsAtty bool
	// die Umgebungsvariablen des Clients
	Env map[string]string
	// die Argumente des Clients
	Args []string
}

// Wurde nur ein Ping-Kommando vom Client gesandt?
//
// Das ist einfach ein leeres Kommando, mit dem der Client prüfen kann,
// ob der Server läuft.
func (c *Cmd) IsPing() bool {
	return c == PingCmd
}
