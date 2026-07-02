package soargs

// Wird vom [ClientChannel] geliefert.
type ClientResult struct {
	// Der verbundene Client oder `nil` (Fehler)
	Client *Client
	// Der eventuell von [WaitForClient] gelieferte Fehler
	Err    error
}
