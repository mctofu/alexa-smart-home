package alexa

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

const sampleRequest = `{
    "directive": {
        "header": {
            "namespace": "Alexa",
            "name": "ReportState",
            "payloadVersion": "3",
            "messageId": "e9d21467-85db-4f34-90d7-0b9d92759f16",
            "correlationToken": "correlationTokenSample"
        },
        "endpoint": {
            "scope": {
                "type": "BearerToken",
                "token": "bearerTokenSample"
            },
            "endpointId": "temp-sensor-1",
            "cookie": {}
        },
        "payload": {}
    }
}`

const expectedResponse = `{
    "context": {
        "properties": [
            {
                "namespace": "Alexa.TemperatureSensor",
                "name": "temperature",
                "value": {
                    "value": 77,
                    "scale": "FAHRENHEIT"
                },
                "timeOfSample": "2018-08-20T05:57:00Z",
                "uncertaintyInMilliseconds": 60000
            }
        ]
    },
    "event": {
        "header": {
            "namespace": "Alexa",
            "name": "StateReport",
            "messageId": "843cf5d3-1923-4508-bc5e-8d30da3e593b",
            "correlationToken": "correlationTokenSample",
            "payloadVersion": "3"
        },
        "endpoint": {
            "endpointId": "temp-sensor-1",
            "scope": {
                "type": "BearerToken",
                "token": "bearerTokenSample"
            }
        },
        "payload": {}
    }
}`

func TestBasicHandler(t *testing.T) {
	tempReader := &mockTempReader{
		77,
		&ResponseBuilder{func() string { return "843cf5d3-1923-4508-bc5e-8d30da3e593b" }},
		func() time.Time { return time.Date(2018, 8, 20, 5, 57, 0, 0, time.UTC) },
	}
	mux := NewNamespaceMux()
	mux.HandleFunc(NamespaceAlexa, tempReader.GetTemperature)

	req := &Request{}
	if err := json.Unmarshal([]byte(sampleRequest), req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	ctx := context.Background()

	resp, err := mux.HandleRequest(ctx, req)
	if err != nil {
		t.Fatalf("Failed to handle request: %v", err)
	}

	respJSON, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	if string(respJSON) != expectedResponse {
		t.Fatalf("Response json does not match expected")
	}
}

type mockTempReader struct {
	temperature float32
	respBuilder *ResponseBuilder
	now         func() time.Time
}

func (t *mockTempReader) GetTemperature(ctx context.Context, req *Request) (*Response, error) {
	temp := TemperatureValue{
		Value: t.temperature,
		Scale: TemperatureScaleFahrenheit,
	}

	tempJSON, err := json.Marshal(temp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal temp: %v", err)
	}

	return t.respBuilder.StateReportResponse(req,
		ContextProperty{
			Namespace:                 NamespaceTemperatureSensor,
			Name:                      "temperature",
			Value:                     tempJSON,
			TimeOfSample:              t.now(),
			UncertaintyInMilliseconds: 60000,
		}), nil
}
