package alexa

import (
	"context"
	"fmt"
)

// Handler responds to an Alexa smart home skill request
type Handler interface {
	HandleRequest(ctx context.Context, req *Request) (*Response, error)
}

// HandlerFunc implements Handler as a func
type HandlerFunc func(ctx context.Context, req *Request) (*Response, error)

// HandleRequest calls the HandlerFunc
func (h HandlerFunc) HandleRequest(ctx context.Context, req *Request) (*Response, error) {
	return h(ctx, req)
}

// NamespaceMux performs routing of skill requests to handlers based on the namespace value
// in the request.
type NamespaceMux struct {
	handlerMap map[string]Handler
}

// NewNamespaceMux creates a NamespaceMux
func NewNamespaceMux() *NamespaceMux {
	return &NamespaceMux{make(map[string]Handler)}
}

// HandleRequest delegates the request to the handler registered for the request's namespace.
// An error is returned if the namespace is unregistered.
func (n *NamespaceMux) HandleRequest(ctx context.Context, req *Request) (*Response, error) {
	handler := n.handlerMap[req.Directive.Header.Namespace]
	if handler == nil {
		return nil, fmt.Errorf("NamespaceMux: unhandled namespace: %s", req.Directive.Header.Namespace)
	}
	return handler.HandleRequest(ctx, req)
}

// Handle registers a Handler for the namespace
func (n *NamespaceMux) Handle(namespace string, handler Handler) {
	n.handlerMap[namespace] = handler
}

// HandleFunc registers a HandlerFunc for the namespace
func (n *NamespaceMux) HandleFunc(namespace string, handler HandlerFunc) {
	n.Handle(namespace, handler)
}

// EndpointMux routes a request based on the requested endpoint
type EndpointMux struct {
	handlerMap map[string]Handler
}

// NewEndpointMux creates an EndpointMux
func NewEndpointMux() *EndpointMux {
	return &EndpointMux{make(map[string]Handler)}
}

// HandleRequest delegates the request to the handler registered for the request's endpoint.
// An error is returned if the endpoint is unregistered.
func (e *EndpointMux) HandleRequest(ctx context.Context, req *Request) (*Response, error) {
	handler := e.handlerMap[req.Directive.Endpoint.EndpointID]
	if handler == nil {
		return nil, fmt.Errorf("EndpointMux: unhandled endpoint: %s", req.Directive.Endpoint.EndpointID)
	}
	resp, err := handler.HandleRequest(ctx, req)
	if err != nil {
		return resp, fmt.Errorf("EndpointMux: failed to handle %s: %v", req.Directive.Endpoint.EndpointID, err)
	}

	return resp, nil
}

// Handle registers a Handler for the endpoint
func (e *EndpointMux) Handle(endpoint string, handler Handler) {
	e.handlerMap[endpoint] = handler
}

// HandleFunc registers a HandlerFunc for the namespace
func (e *EndpointMux) HandleFunc(endpoint string, handler HandlerFunc) {
	e.Handle(endpoint, handler)
}
