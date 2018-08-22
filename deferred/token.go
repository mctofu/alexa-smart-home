package deferred

import "golang.org/x/oauth2"

// tokenSniffer wraps a TokenSource to detect token refreshes
// so the updated token can be persisted
type tokenSniffer struct {
	LastToken   *oauth2.Token
	TokenSource oauth2.TokenSource
}

func (t *tokenSniffer) Token() (*oauth2.Token, error) {
	token, err := t.TokenSource.Token()
	if err == nil {
		t.LastToken = token
	}
	return token, err
}
