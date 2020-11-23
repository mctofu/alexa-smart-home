package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	awslambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/mctofu/alexa-smart-home/alexa"
	"github.com/mctofu/alexa-smart-home/aws/s3store"
	"github.com/mctofu/alexa-smart-home/aws/sqsrelay"
	"github.com/mctofu/alexa-smart-home/lambda"
)

// Smart home skill lambda implementation that allows discovery of a mock temperature
// sensor and a mock power controller. Temperature requests are responded to
// with a canned response. Power controller requests return a deferred response
// and publish a SQS message to allow the sqsagent to handle it remotely.
func main() {
	sqsQueueURL := os.Getenv("SQS_QUEUE_URL")
	s3TokenBucket := os.Getenv("S3_TOKEN_BUCKET")
	authClientID := os.Getenv("AUTH_CLIENT_ID")
	authClientSecret := os.Getenv("AUTH_CLIENT_SECRET")

	session, err := session.NewSession()
	if err != nil {
		log.Fatalf("failed to init aws session: %v", err)
	}

	respBuilder := alexa.NewResponseBuilder()

	tempReader := tempReader{75, respBuilder}

	sqs := sqs.New(session)
	sqsRelay := &sqsrelay.RelayHandler{
		SQS:      sqs,
		QueueURL: sqsQueueURL,
	}

	s3Client := s3.New(session)
	tokenStorage := &alexa.DebugTokenStore{
		TokenStore: &s3store.TokenStorage{
			S3:     s3Client,
			Bucket: s3TokenBucket,
		},
	}
	userIDReader := &alexa.ProfileUserIDReader{HTTPDoer: http.DefaultClient}

	mux := alexa.NewNamespaceMux()
	mux.HandleFunc(alexa.NamespacePowerController, alexa.DeferredRelayHandler(sqsRelay, respBuilder))
	mux.HandleFunc(alexa.NamespaceDiscovery, alexa.StaticDiscoveryHandler(respBuilder, endpoints()...))
	mux.HandleFunc(alexa.NamespaceAlexa, tempReader.GetTemperature)
	mux.HandleFunc(alexa.NamespaceAuthorization,
		alexa.AuthorizationHandler(
			authClientID,
			authClientSecret,
			userIDReader,
			tokenStorage,
			respBuilder))

	awslambda.Start(lambda.DebugLambdaRequestHandler(mux))
}

func endpoints() []alexa.DiscoverEndpoint {
	return []alexa.DiscoverEndpoint{
		{
			EndpointID:        "temp-sensor-1",
			FriendlyName:      "Home Temperature",
			Description:       "Temp monitor",
			ManufacturerName:  "McTofu",
			DisplayCategories: []string{alexa.DisplayCategoryTemperatureSensor},
			Capabilities: []alexa.DiscoverCapability{
				{
					Type:      "AlexaInterface",
					Interface: alexa.InterfaceTemperatureSensor,
					Version:   "3",
					Properties: &alexa.DiscoverProperties{
						Supported: []alexa.DiscoverProperty{
							{
								Name: "temperature",
							},
						},
						ProactivelyReported: false,
						Retrievable:         true,
					},
				},
			},
		},
		{
			EndpointID:        "switch-1",
			FriendlyName:      "Fan",
			Description:       "Power switch for fan",
			ManufacturerName:  "McTofu",
			DisplayCategories: []string{alexa.DisplayCategorySwitch},
			Capabilities: []alexa.DiscoverCapability{
				{
					Type:      "AlexaInterface",
					Interface: alexa.InterfacePowerController,
					Version:   "3",
					Properties: &alexa.DiscoverProperties{
						Supported: []alexa.DiscoverProperty{
							{
								Name: "powerState",
							},
						},
						ProactivelyReported: true,
						Retrievable:         true,
					},
				},
			},
		},
	}
}

type tempReader struct {
	temperature float32
	respBuilder *alexa.ResponseBuilder
}

func (t *tempReader) GetTemperature(ctx context.Context, req *alexa.Request) (*alexa.Response, error) {
	now := time.Now()

	temp := alexa.TemperatureValue{
		Value: t.temperature,
		Scale: alexa.TemperatureScaleFahrenheit,
	}

	tempJSON, err := json.Marshal(temp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal temp: %v", err)
	}

	return t.respBuilder.StateReportResponse(req,
		alexa.ContextProperty{
			Namespace:                 alexa.NamespaceTemperatureSensor,
			Name:                      "temperature",
			Value:                     tempJSON,
			TimeOfSample:              now,
			UncertaintyInMilliseconds: 60000,
		}), nil
}
