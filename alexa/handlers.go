package alexa

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/amazon"
)

// Relayer sends the request somewhere else for handling. It's expected
// that a response will be posted back to the smart home api.
type Relayer interface {
	Relay(ctx context.Context, req *Request) error
}

// DeferredRelayHandler handles a request by relaying it to the relayer and returning
// a DeferredResponse indicating that the actual response will be sent to the smart home
// api
func DeferredRelayHandler(relayer Relayer, builder *ResponseBuilder) HandlerFunc {
	return func(ctx context.Context, req *Request) (*Response, error) {
		if err := relayer.Relay(ctx, req); err != nil {
			return nil, fmt.Errorf("failed to relay: %v", err)
		}
		return builder.DeferredResponse(req), nil
	}
}

// StaticDiscoveryHandler handles discovery requests with a hardcoded set
// of endpoints
func StaticDiscoveryHandler(builder *ResponseBuilder, endpoints ...DiscoverEndpoint) HandlerFunc {
	return func(ctx context.Context, req *Request) (*Response, error) {
		resp, err := builder.DiscoverResponse(endpoints...)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

// AuthorizationHandler handles an Authorization AcceptGrant request and fetches credentials required
// to post events to the smart home api
func AuthorizationHandler(clientID, clientSecret string,
	userIDReader UserIDReader, tokenWriter TokenWriter, respBuilder *ResponseBuilder) HandlerFunc {
	return func(ctx context.Context, req *Request) (*Response, error) {
		var payload AcceptGrantPayload
		if err := json.Unmarshal(req.Directive.Payload, &payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %v", err)
		}

		config := oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     amazon.Endpoint,
		}

		token, err := config.Exchange(ctx, payload.Grant.Code)
		if err != nil {
			resp, err := respBuilder.BasicErrorResponse(req,
				"ACCEPT_GRANT_FAILED",
				fmt.Sprintf("failed to exchange token: %v", err))
			if err != nil {
				return nil, fmt.Errorf("failed to create error response: %v", err)
			}
			return resp, nil
		}

		userID, err := userIDReader.Read(ctx, payload.Grantee.Token)
		if err != nil {
			resp, err := respBuilder.BasicErrorResponse(req,
				"ACCEPT_GRANT_FAILED",
				fmt.Sprintf("failed to lookup userid: %v", err))
			if err != nil {
				return nil, fmt.Errorf("failed to create error response: %v", err)
			}
			return resp, nil
		}

		if err := tokenWriter.Write(ctx, userID, token); err != nil {
			resp, err := respBuilder.BasicErrorResponse(req,
				"ACCEPT_GRANT_FAILED",
				fmt.Sprintf("failed to store token: %v", err))
			if err != nil {
				return nil, fmt.Errorf("failed to create error response: %v", err)
			}
			return resp, nil
		}

		return respBuilder.AcceptGrantResponse(), nil
	}
}

// PercentageControllerHandler routes handling of set & adjust directives
func PercentageControllerHandler(setPct, adjustPct Handler) HandlerFunc {
	return func(ctx context.Context, req *Request) (*Response, error) {
		switch req.Directive.Header.Name {
		case "SetPercentage":
			return setPct.HandleRequest(ctx, req)
		case "AdjustPercentage":
			return adjustPct.HandleRequest(ctx, req)
		default:
			return nil, fmt.Errorf("PercentageControllerHandler: unexpected name: %s", req.Directive.Header.Name)
		}
	}
}

// PowerControllerHandler routes turn on & off requests
func PowerControllerHandler(turnOn, turnOff Handler) HandlerFunc {
	return func(ctx context.Context, req *Request) (*Response, error) {
		switch req.Directive.Header.Name {
		case "TurnOn":
			return turnOn.HandleRequest(ctx, req)
		case "TurnOff":
			return turnOff.HandleRequest(ctx, req)
		default:
			return nil, fmt.Errorf("PowerControllerHandler: unexpected name: %s", req.Directive.Header.Name)
		}
	}
}

// SceneControllerHandler routes activate & deactivate requests
func SceneControllerHandler(activate, deactivate Handler) HandlerFunc {
	return func(ctx context.Context, req *Request) (*Response, error) {
		switch req.Directive.Header.Name {
		case "Activate":
			return activate.HandleRequest(ctx, req)
		case "Deactivate":
			return deactivate.HandleRequest(ctx, req)
		default:
			return nil, fmt.Errorf("SceneControllerHandler: unexpected name: %s", req.Directive.Header.Name)
		}
	}
}
