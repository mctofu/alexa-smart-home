package deferred

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mctofu/alexa-smart-home/alexa"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/amazon"
)

// EventSender publishes a response back to the smart home event api
type EventSender interface {
	Send(ctx context.Context, resp *alexa.Response) error
}

// EventSenderFunc publishes a response back to the smart home event api
type EventSenderFunc func(ctx context.Context, resp *alexa.Response) error

// Send publishes a response back to the smart home event api
func (e EventSenderFunc) Send(ctx context.Context, resp *alexa.Response) error {
	return e(ctx, resp)
}

// Handler coordinates handling a request and sending the response back to the smart home event api
type Handler struct {
	RequestHandler alexa.Handler
	EventSender    EventSender
}

// HandleRequest passes the request to the RequestHandler. If response is returned it
// is published via the EventSender. An error of type SendError indicates the request
// was successful but the response failed to be sent.
func (h *Handler) HandleRequest(ctx context.Context, req *alexa.Request) error {
	resp, err := h.RequestHandler.HandleRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to handle request: %v", err)
	}
	if resp == nil {
		return nil
	}

	return h.EventSender.Send(ctx, resp)
}

// HTTPEventSender sends responses to the smart home api with the credentials of the user.
type HTTPEventSender struct {
	TokenStore   alexa.TokenReaderWriter
	UserIDReader alexa.UserIDReader
	ClientID     string
	ClientSecret string
}

// Send responses to the smart home api with the credentials of the user.
func (h *HTTPEventSender) Send(ctx context.Context, resp *alexa.Response) error {
	respJSON, err := json.Marshal(resp)
	if err != nil {
		return &SendError{fmt.Sprintf("failed to marshal response: %v", err)}
	}

	profile, err := h.UserIDReader.Read(ctx, resp.Event.Endpoint.Scope.Token)
	if err != nil {
		return &SendError{fmt.Sprintf("failed to retrieve user id: %v", err)}
	}

	token, err := h.TokenStore.Read(ctx, profile)
	if err != nil {
		return &SendError{fmt.Sprintf("failed to retrieve access token: %v", err)}
	}
	if token == nil {
		return &SendError{fmt.Sprintf("missing access token")}
	}

	eventReq, err := http.NewRequest(http.MethodPost, "https://api.amazonalexa.com/v3/events", bytes.NewReader(respJSON))
	if err != nil {
		return &SendError{fmt.Sprintf("failed to build event request: %v", err)}
	}

	eventReq = eventReq.WithContext(ctx)
	eventReq.Header.Set("Content-Type", "application/json")
	eventReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	oauth2Config := oauth2.Config{
		ClientID:     h.ClientID,
		ClientSecret: h.ClientSecret,
		Endpoint:     amazon.Endpoint,
	}

	tokenSniffer := &tokenSniffer{TokenSource: oauth2Config.TokenSource(ctx, token)}
	httpClient := oauth2.NewClient(ctx, tokenSniffer)

	eventResp, err := httpClient.Do(eventReq)
	if err != nil {
		return &SendError{fmt.Sprintf("failed to perform event request: %v", err)}
	}
	defer eventResp.Body.Close()

	if _, err := ioutil.ReadAll(eventResp.Body); err != nil {
		return &SendError{fmt.Sprintf("failed to read event body: %v", err)}
	}

	if eventResp.StatusCode != http.StatusOK && eventResp.StatusCode != http.StatusAccepted {
		return &SendError{fmt.Sprintf("event response unexpected status code: %s", eventResp.Status)}
	}

	if tokenSniffer.LastToken != nil && token.AccessToken != tokenSniffer.LastToken.AccessToken {
		if err := h.TokenStore.Write(ctx, profile, tokenSniffer.LastToken); err != nil {
			return fmt.Errorf("failed to update token: %v", err)
		}
	}

	return nil
}

// SendError is an error sending to the smart home event api
type SendError struct {
	msg string
}

func (r *SendError) Error() string {
	return r.msg
}
