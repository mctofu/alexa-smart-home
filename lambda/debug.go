package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mctofu/alexa-smart-home/alexa"
)

// DebugLambdaRequestHandler logs the lambda request directly for debugging.
func DebugLambdaRequestHandler(handler alexa.Handler) func(context.Context, json.RawMessage) (*alexa.Response, error) {
	return func(ctx context.Context, reqJSON json.RawMessage) (*alexa.Response, error) {
		log.Printf("Debug request:\n%s\n", string(reqJSON))

		var req alexa.Request
		if err := json.Unmarshal(reqJSON, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal request: %v", err)
		}

		return alexa.ResponseDebugHandler(handler).HandleRequest(ctx, &req)
	}
}
