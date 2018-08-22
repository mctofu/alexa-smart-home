package alexa

import (
	"encoding/json"
	"time"
)

// Structs/types for alexa smart home skill api requests/responses
// These should mirror the api reference at:
// https://developer.amazon.com/docs/smarthome/smart-home-skill-api-message-reference.html

// Request represents an incoming request from the smart home service
type Request struct {
	Directive RequestDirective `json:"directive"`
}

type RequestDirective struct {
	Header   Header          `json:"header"`
	Endpoint RequestEndpoint `json:"endpoint"`
	Payload  json.RawMessage `json:"payload"`
}

type Header struct {
	Namespace        string `json:"namespace"`
	Name             string `json:"name"`
	MessageID        string `json:"messageId"`
	CorrelationToken string `json:"correlationToken,omitempty"`
	PayloadVersion   string `json:"payloadVersion"`
}

type RequestEndpoint struct {
	Scope      Scope             `json:"scope,omitempty"`
	EndpointID string            `json:"endpointId,omitempty"`
	Cookie     map[string]string `json:"cookie,omitempty"`
}

type Scope struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// Response represents a response to a request from the smart home service
type Response struct {
	Context *ResponseContext `json:"context,omitempty"`
	Event   Event            `json:"event"`
}

type ResponseContext struct {
	Properties []ContextProperty `json:"properties,omitempty"`
}

// Namespace enums
const (
	NamespaceAlexa             = "Alexa"
	NamespaceAuthorization     = "Alexa.Authorization"
	NamespaceDiscovery         = "Alexa.Discovery"
	NamespacePowerController   = "Alexa.PowerController"
	NamespaceTemperatureSensor = "Alexa.TemperatureSensor"
)

type ContextProperty struct {
	Namespace                 string          `json:"namespace"`
	Name                      string          `json:"name"`
	Value                     json.RawMessage `json:"value"`
	TimeOfSample              time.Time       `json:"timeOfSample"`
	UncertaintyInMilliseconds int32           `json:"uncertaintyInMilliseconds"`
}

type Event struct {
	Header   Header            `json:"header"`
	Endpoint *ResponseEndpoint `json:"endpoint,omitempty"`
	Payload  json.RawMessage   `json:"payload"`
}

type ResponseEndpoint struct {
	EndpointID string            `json:"endpointId,omitempty"`
	Cookie     map[string]string `json:"cookie,omitempty"`
	Scope      Scope             `json:"scope,omitempty"`
}

// DisplayCategory enums
const (
	DisplayCategoryDoor              = "DOOR"
	DisplayCategorySwitch            = "SWITCH"
	DisplayCategoryTemperatureSensor = "TEMPERATURE_SENSOR"
	DisplayCategoryOther             = "OTHER"
)

// Interface enums
const (
	InterfaceTemperatureSensor = NamespaceTemperatureSensor
	InterfacePowerController   = NamespacePowerController
)

// EmptyPayload is a payload with no content
var EmptyPayload = json.RawMessage("{}")

type DiscoverPayload struct {
	Endpoints []DiscoverEndpoint `json:"endpoints"`
}

type DiscoverEndpoint struct {
	EndpointID        string               `json:"endpointId"`
	ManufacturerName  string               `json:"manufacturerName"`
	FriendlyName      string               `json:"friendlyName"`
	Description       string               `json:"description"`
	DisplayCategories []string             `json:"displayCategories"`
	Cookie            map[string]string    `json:"cookie,omitempty"`
	Capabilities      []DiscoverCapability `json:"capabilities"`
}

type DiscoverCapability struct {
	Type       string             `json:"type"`
	Interface  string             `json:"interface"`
	Version    string             `json:"version"`
	Properties DiscoverProperties `json:"properties"`
}

type DiscoverProperties struct {
	Supported           []DiscoverProperty `json:"supported,omitempty"`
	ProactivelyReported bool               `json:"proactivelyReported"`
	Retrievable         bool               `json:"retrievable"`
}

type DiscoverProperty struct {
	Name string `json:"name"`
}

type AcceptGrantPayload struct {
	Grant   AcceptGrantGrant   `json:"grant"`
	Grantee AcceptGrantGrantee `json:"grantee"`
}

type AcceptGrantGrant struct {
	Type string `json:"type"`
	Code string `json:"code"`
}

type AcceptGrantGrantee struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// TemperatureScale enums
const (
	TemperatureScaleFahrenheit = "FAHRENHEIT"
)

type TemperatureValue struct {
	Value float32 `json:"value"`
	Scale string  `json:"scale"`
}
