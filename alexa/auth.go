package alexa

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
)

// TokenReaderWriter provides read & write access to tokens
type TokenReaderWriter interface {
	TokenReader
	TokenWriter
}

// TokenWriter provides secure storage for a user's oauth tokens
type TokenWriter interface {
	Write(ctx context.Context, id string, token *oauth2.Token) error
}

// TokenReader provides secure retrieval for a user's oauth tokens
type TokenReader interface {
	Read(ctx context.Context, id string) (*oauth2.Token, error)
}

// UserIDReader uses the bearerToken from the skill request to look up the user's id
type UserIDReader interface {
	Read(ctx context.Context, bearerToken string) (string, error)
}

// HTTPDoer performs a HTTP request (HTTPClient implements this)
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// ProfileUserIDReader retrieves the user's Amazon account user id.
// It also has access to the user's name and email but it is not returned.
type ProfileUserIDReader struct {
	HTTPDoer HTTPDoer
}

func (p *ProfileUserIDReader) Read(ctx context.Context, bearerToken string) (string, error) {
	profileReq, err := http.NewRequest(http.MethodGet, "https://api.amazon.com/user/profile", nil)
	if err != nil {
		return "", fmt.Errorf("failed to build profile request: %v", err)
	}

	profileReq = profileReq.WithContext(ctx)
	profileReq.Header.Set("Content-Type", "application/json")
	profileReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearerToken))

	profileResp, err := p.HTTPDoer.Do(profileReq)
	if err != nil {
		return "", fmt.Errorf("failed to perform profile request: %v", err)
	}
	defer profileResp.Body.Close()

	respBody, err := ioutil.ReadAll(profileResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read profile body: %v", err)
	}

	if profileResp.StatusCode != http.StatusOK && profileResp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("profile response unexpected status code: %s", profileResp.Status)
	}

	profileData := struct {
		UserID string `json:"user_id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
	}{}

	if err := json.Unmarshal(respBody, &profileData); err != nil {
		return "", fmt.Errorf("failed to unmarshal profile data: %v", err)
	}

	return profileData.UserID, nil
}
