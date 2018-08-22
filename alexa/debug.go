package alexa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/mctofu/alexa-smart-home/schema"
	"github.com/xeipuuv/gojsonschema"
	"golang.org/x/oauth2"
)

// DebugHandler wraps handler and logs the contents of the request and response for debugging.
// The response is also validated against the smart home schema.
func DebugHandler(handler Handler) Handler {
	return ResponseDebugHandler(RequestDebugHandler(handler))
}

// RequestDebugHandler wraps handler and logs the contents of the request for debugging.
func RequestDebugHandler(handler Handler) HandlerFunc {
	return func(ctx context.Context, req *Request) (*Response, error) {
		reqJSON, err := json.Marshal(req)
		if err != nil {
			log.Printf("RequestDebugHandler: Failed to marshal request: %v", err)
		} else {
			log.Printf("RequestDebugHandler: Debug request:\n%s\n", string(reqJSON))
		}

		return handler.HandleRequest(ctx, req)
	}
}

// ResponseDebugHandler wraps handler and logs the contents of the response for debugging.
// The response is also validated against the smart home schema.
func ResponseDebugHandler(handler Handler) HandlerFunc {
	return func(ctx context.Context, req *Request) (*Response, error) {
		resp, err := handler.HandleRequest(ctx, req)

		respJSON, jsonErr := json.Marshal(resp)
		if jsonErr != nil {
			log.Printf("Failed to marshal debug response: %v\n", jsonErr)
		}
		log.Printf("Debug response:\n%s\n", respJSON)

		if schemaErr := validateSchema(string(respJSON)); schemaErr != nil {
			log.Printf("Failed to validate schema: %v\n", schemaErr)
		} else {
			log.Printf("Schema validated!\n")
		}

		return resp, err
	}
}

func validateSchema(resp string) error {
	schemaLoader := gojsonschema.NewStringLoader(schema.AlexaSmartHome)
	result, err := gojsonschema.Validate(schemaLoader, gojsonschema.NewStringLoader(resp))
	if err != nil {
		return fmt.Errorf("Failed to validate schema: %v", err)
	}
	if !result.Valid() {
		log.Printf("Response is not valid:\n")
		for _, desc := range result.Errors() {
			log.Printf("- %s\n", desc)
		}
		return errors.New("Response is not valid")
	}
	return nil
}

// DebugTokenStore logs reads/writes to tokens
type DebugTokenStore struct {
	TokenStore TokenReaderWriter
}

func (d *DebugTokenStore) Write(ctx context.Context, id string, token *oauth2.Token) error {
	log.Printf("Writing token for %s\n", id)
	return d.TokenStore.Write(ctx, id, token)
}

func (d *DebugTokenStore) Read(ctx context.Context, id string) (*oauth2.Token, error) {
	log.Printf("Reading token for %s\n", id)
	return d.TokenStore.Read(ctx, id)
}
