package alexa

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// UUIDMessageID generates a uuid suitable for use as a MessageID
func UUIDMessageID() string {
	return uuid.New().String()
}

// ResponseBuilder assists in generating proper responses for the smart home
// skill api
type ResponseBuilder struct {
	// MessageID should generate a unique identifier for a response. UUID recommended.
	MessageID func() string
}

// NewResponseBuilder creates a new ResponseBuilder with a UUID MessageID generator.
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{UUIDMessageID}
}

// DeferredResponse creates a response that indicates that a response will be
// sent to the smart home event api rather than being returned immediately.
func (r *ResponseBuilder) DeferredResponse(req *Request) *Response {
	return &Response{
		Event: Event{
			Header: Header{
				Namespace:        NamespaceAlexa,
				Name:             "DeferredResponse",
				PayloadVersion:   "3",
				MessageID:        r.MessageID(),
				CorrelationToken: req.Directive.Header.CorrelationToken,
			},
			Payload: EmptyPayload,
		},
	}
}

// DiscoverResponse creates a response that describes the available capabilities
func (r *ResponseBuilder) DiscoverResponse(endpoints ...DiscoverEndpoint) (*Response, error) {
	payload := DiscoverPayload{
		Endpoints: endpoints,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp := Response{
		Event: Event{
			Header: Header{
				Namespace:      "Alexa.Discovery",
				Name:           "Discover.Response",
				PayloadVersion: "3",
				MessageID:      r.MessageID(),
			},
			Payload: payloadJSON,
		},
	}

	return &resp, nil
}

// BasicErrorResponse creates a response for simple errors
func (r *ResponseBuilder) BasicErrorResponse(req *Request, errorType, msg string) (*Response, error) {
	payload := struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}{errorType, msg}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}
	return &Response{
		Event: Event{
			Header: Header{
				Namespace:        req.Directive.Header.Namespace,
				Name:             "ErrorResponse",
				PayloadVersion:   "3",
				MessageID:        r.MessageID(),
				CorrelationToken: req.Directive.Header.CorrelationToken,
			},
			Endpoint: &ResponseEndpoint{
				EndpointID: req.Directive.Endpoint.EndpointID,
				Scope:      req.Directive.Endpoint.Scope,
			},
			Payload: payloadJSON,
		},
	}, nil
}

// CustomErrorResponse returns an error response with custom payload
func (r *ResponseBuilder) CustomErrorResponse(req *Request, payload json.RawMessage) *Response {
	return &Response{
		Event: Event{
			Header: Header{
				Namespace:        req.Directive.Header.Namespace,
				Name:             "ErrorResponse",
				PayloadVersion:   "3",
				MessageID:        r.MessageID(),
				CorrelationToken: req.Directive.Header.CorrelationToken,
			},
			Endpoint: &ResponseEndpoint{
				EndpointID: req.Directive.Endpoint.EndpointID,
				Scope:      req.Directive.Endpoint.Scope,
			},
			Payload: payload,
		},
	}
}

// StateReportResponse builds a StateReport response with the provided properties
func (r *ResponseBuilder) StateReportResponse(req *Request, properties ...ContextProperty) *Response {
	return &Response{
		Event: Event{
			Header: Header{
				Namespace:        NamespaceAlexa,
				Name:             "StateReport",
				PayloadVersion:   "3",
				MessageID:        r.MessageID(),
				CorrelationToken: req.Directive.Header.CorrelationToken,
			},
			Endpoint: &ResponseEndpoint{
				EndpointID: req.Directive.Endpoint.EndpointID,
				Scope:      req.Directive.Endpoint.Scope,
			},
			Payload: EmptyPayload,
		},
		Context: &ResponseContext{
			Properties: properties,
		},
	}
}

// BasicResponse returns a response event response
func (r *ResponseBuilder) BasicResponse(req *Request, properties ...ContextProperty) *Response {
	return &Response{
		Event: Event{
			Header: Header{
				Namespace:        NamespaceAlexa,
				Name:             "Response",
				PayloadVersion:   "3",
				MessageID:        r.MessageID(),
				CorrelationToken: req.Directive.Header.CorrelationToken,
			},
			Endpoint: &ResponseEndpoint{
				EndpointID: req.Directive.Endpoint.EndpointID,
				Scope:      req.Directive.Endpoint.Scope,
			},
			Payload: EmptyPayload,
		},
		Context: &ResponseContext{
			Properties: properties,
		},
	}
}

// AcceptGrantResponse returns a successful accept grant response
func (r *ResponseBuilder) AcceptGrantResponse() *Response {
	return &Response{
		Event: Event{
			Header: Header{
				Namespace:      NamespaceAuthorization,
				Name:           "AcceptGrant.Response",
				PayloadVersion: "3",
				MessageID:      r.MessageID(),
			},
			Payload: EmptyPayload,
		},
	}
}
